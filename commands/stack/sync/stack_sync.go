package sync

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/xanzy/go-gitlab"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/mr/mrutils"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/glrepo"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/text"
)

type SyncOptions struct {
	LabClient   *gitlab.Client
	CurrentUser *gitlab.User
	BaseRepo    func() (glrepo.Interface, error)
	Remotes     func() (glrepo.Remotes, error)
	Config      func() (config.Config, error)
}

var iostream *iostreams.IOStreams

func NewCmdSyncStack(f *cmdutils.Factory) *cobra.Command {
	opts := &SyncOptions{
		Remotes:  f.Remotes,
		Config:   f.Config,
		BaseRepo: f.BaseRepo,
	}

	iostream = f.IO

	stackSaveCmd := &cobra.Command{
		Use:   "sync",
		Short: `Sync and submit progress on a stacked diff.`,
		Long: heredoc.Doc(`Sync and submit progress on a stacked diff. This command runs these steps:

1. Creates a merge request for any branches without one.
1. Pushes any amended changes to their merge requests.
1. Rebases any changes that happened previously in the stack.
1. Removes any branches that were already merged, or with a closed merge request.
` + text.ExperimentalString),
		Example: heredoc.Doc(`
			glab sync
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := stackSync(f, opts)
			if err != nil {
				return fmt.Errorf("could not run sync: %v", err)
			}

			return nil
		},
	}

	return stackSaveCmd

}

func stackSync(f *cmdutils.Factory, opts *SyncOptions) error {
	iostream.StartSpinner("Syncing")

	repo, err := f.BaseRepo()
	if err != nil {
		return fmt.Errorf("error determining base repo: %v", err)
	}

	opts.LabClient, err = f.HttpClient()
	if err != nil {
		return fmt.Errorf("error utilizing API client: %v", err)
	}
	client := opts.LabClient

	stack, err := getStacks()
	if err != nil {
		return err
	}

	var needsToSyncAgain bool

	for {
		needsToSyncAgain = false

		ref, err := stack.First()
		if err != nil {
			return fmt.Errorf("error getting first stack: %v", err)
		}

		var gr git.StandardGitCommand
		for {
			status, err := branchStatus(ref, gr)
			if err != nil {
				return fmt.Errorf("error getting branch status: %v", err)
			}

			if strings.Contains(status, "Your branch is behind") {
				// possibly someone applied suggestions or someone else added a
				// different commit
				fmt.Println(progressString(ref.Branch + " is behind- pulling."))

				_, err = gitPull(ref, gr)
				if err != nil {
					return fmt.Errorf("error checking for running Git pull: %v", err)
				}

			} else if strings.Contains(status, "have diverged") {
				fmt.Println(progressString(ref.Branch + " has diverged. Rebasing..."))

				err := rebaseWithUpdateRefs(ref, stack, gr)
				if err != nil {
					return fmt.Errorf(errorString(
						"could not rebase due to a merge conflict.",
						"Fix the issues with Git and run `glab sync stack` again.",
					))
				}

				err = forcePushAll(stack, gr)
				if err != nil {
					return fmt.Errorf("error pushing branches %v", err)
				}

				// since a branch diverged and we need to rebase, we're going to have
				// to push all the subsequent stacks
				needsToSyncAgain = true
			} else {
				return fmt.Errorf("your Git branch is ahead when it shouldn't be. You might need to squash your commits.")
			}

			if ref.MR == "" {
				// no MR - lets create one!
				fmt.Println(progressString(ref.Branch + " needs a merge request. Creating it now."))

				mr, err := createMR(client, repo, stack, ref, gr)
				if err != nil {
					return fmt.Errorf("error updating stack ref files: %v", err)
				}

				fmt.Println(progressString("Merge request created!"))
				fmt.Println(mrutils.DisplayMR(iostream.Color(), mr, true))

				// update the ref
				ref.MR = mr.WebURL
				err = git.UpdateStackRefFile(stack.Title, ref)
				if err != nil {
					return fmt.Errorf("error updating stack ref files: %v", err)
				}

			} else {

				// we have an MR. let's make sure it's still open.
				mr, _, err := mrutils.MRFromArgsWithOpts(f, nil, nil, "opened")
				if err != nil {
					return fmt.Errorf("error getting merge request from branch: %v", err)
				}
				mergeOldMr(ref, mr, &stack)
			}

			if ref.Next == "" {
				break
			} else {
				ref = stack.Refs[ref.Next]
			}
		}

		if needsToSyncAgain != true {
			break
		}
	}

	iostream.StopSpinner("")

	fmt.Printf(progressString("Sync finished!"))

	return nil
}

func getStacks() (git.Stack, error) {
	title, err := git.GetCurrentStackTitle()
	if err != nil {
		return git.Stack{}, fmt.Errorf("error getting current stack: %v", err)
	}

	stack, err := git.GatherStackRefs(title)
	if err != nil {
		return git.Stack{}, fmt.Errorf("error getting current stack references: %v", err)
	}
	return stack, nil
}

func gitPull(ref git.StackRef, gr git.GitRunner) (string, error) {
	checkout, err := gr.Git("checkout", ref.Branch)
	if err != nil {
		return "", err
	}
	debug("Checked out:", checkout)

	upstream, err := gr.Git(
		"branch",
		"--set-upstream-to",
		fmt.Sprintf("%s/%s", git.DefaultRemote, ref.Branch),
	)
	if err != nil {
		return "", err
	}
	debug("Set upstream:", upstream)

	pull, err := gr.Git("pull")
	if err != nil {
		return "", err
	}
	debug("Pulled:", pull)

	return pull, nil
}

func branchStatus(ref git.StackRef, gr git.GitRunner) (string, error) {
	checkout, err := gr.Git("checkout", ref.Branch)
	if err != nil {
		return "", err
	}
	debug("Checked out:", checkout)

	output, err := gr.Git("status", "-uno")
	if err != nil {
		return "", err
	}
	debug("Git status:", output)

	return output, nil
}

func rebaseWithUpdateRefs(ref git.StackRef, stack git.Stack, gr git.GitRunner) error {
	lastRef, err := stack.Last()
	if err != nil {
		return err
	}

	checkout, err := gr.Git("checkout", lastRef.Branch)
	if err != nil {
		return err
	}
	debug("Checked out:", checkout)

	rebase, err := gr.Git("rebase", "--fork-point", "--update-refs", ref.Branch)
	if err != nil {
		return err
	}
	debug("Rebased:", rebase)

	return nil
}

func forcePushAll(stack git.Stack, gr git.GitRunner) error {
	for _, r := range stack.Refs {
		fmt.Printf(progressString("Updating branch", r.Branch))

		output, err := gr.Git("checkout", r.Branch)
		if err != nil {
			return err
		}
		fmt.Printf(progressString("Checked out: " + output))

		err = forcePush(gr)
		if err != nil {
			return err
		}
	}

	return nil
}

func forcePush(gr git.GitRunner) error {
	output, err := gr.Git("push", git.DefaultRemote, "--force-with-lease")
	if err != nil {
		return err
	}

	fmt.Printf(progressString("Push succeeded: " + output))
	return nil
}

func createMR(client *gitlab.Client, repo glrepo.Interface, stack git.Stack, ref git.StackRef, gr git.GitRunner) (*gitlab.MergeRequest, error) {
	_, err := gr.Git("push", "-u", git.DefaultRemote, ref.Branch)
	if err != nil {
		return &gitlab.MergeRequest{}, fmt.Errorf("error pushing branch: %v", err)
	}

	var previousBranch string
	if ref.Prev != "" {
		// if we have a previous branch, let's point to that
		previousBranch = stack.Refs[ref.Prev].Branch
	} else {
		// otherwise, we'll point to the default one
		previousBranch, err = git.GetDefaultBranch("origin")
		if err != nil {
			return &gitlab.MergeRequest{}, fmt.Errorf("error getting default branch: %v", err)
		}
	}

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return &gitlab.MergeRequest{}, err
	}

	l := &gitlab.CreateMergeRequestOptions{
		Title:              gitlab.Ptr(ref.Description),
		SourceBranch:       gitlab.Ptr(ref.Branch),
		TargetBranch:       gitlab.Ptr(previousBranch),
		AssigneeID:         gitlab.Ptr(user.ID),
		RemoveSourceBranch: gitlab.Ptr(true),
	}

	mr, err := api.CreateMR(client, repo.FullName(), l)
	if err != nil {
		return &gitlab.MergeRequest{}, fmt.Errorf("error creating merge request with the API: %v", err)
	}

	return mr, nil
}

func mergeOldMr(ref git.StackRef, mr *gitlab.MergeRequest, stack *git.Stack) {
	if mr.State == "merged" {
		string := fmt.Sprintf("MR !%v has merged, removing reference...", mr.IID)
		fmt.Println(progressString(string))

		stack.RemoveRef(ref)
	} else if mr.State == "closed" {
		string := fmt.Sprintf("MR !%v has closed, removing reference...", mr.IID)
		fmt.Println(progressString(string))

		stack.RemoveRef(ref)
	}
}

func errorString(errors ...string) string {
	redCheck := iostream.Color().Red("✘")

	title := errors[0]
	body := strings.Join(errors[1:], "\n  ")

	return fmt.Sprintf("\n%s %s \n  %s", redCheck, title, body)
}

func progressString(lines ...string) string {
	blueDot := iostream.Color().ProgressIcon()
	title := lines[0]

	var body string

	if len(lines) > 1 {
		body = strings.Join(lines[1:], "\n  ")
		return fmt.Sprintf("\n%s %s \n  %s", blueDot, title, body)
	} else {
		return fmt.Sprintf("\n%s %s\n", blueDot, title)
	}

}

func debug(output ...string) {
	if os.Getenv("DEBUG") != "" {
		log.Print(output)
	}
}
