package save

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/prompt"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdAmendStack(f *cmdutils.Factory) *cobra.Command {
	stackSaveCmd := &cobra.Command{
		Use:   "amend",
		Short: `Save additional progress on a stacked diff`,
		Long: `Amend lets you add a change to an already created stack

This is an experimental feature that might be broken or removed without any prior notice.
Read more about what experimental features mean at <https://docs.gitlab.com/ee/policy/experiment-beta-support.html#experiment>

This is an experimental feature. Use at your own risk.
`,
		Example: heredoc.Doc(`glab stack amend modifiedfile
			glab stack amend . -m "fixed a function"
			glab stack amend newfile -d "forgot to add this"`),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := amendFunc(f, args, description)
			if err != nil {
				return fmt.Errorf("could not run stack amend: %v", err)
			}

			if f.IO.IsOutputTTY() {
				fmt.Fprintf(f.IO.StdOut, output)
			}

			return nil
		},
	}
	stackSaveCmd.Flags().StringVarP(&description, "description", "d", "", "a description of the change")
	stackSaveCmd.Flags().StringVarP(&description, "message", "m", "", "alias for description flag")
	stackSaveCmd.MarkFlagsMutuallyExclusive("message", "description")

	return stackSaveCmd
}

func amendFunc(f *cmdutils.Factory, args []string, description string) (string, error) {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)

	// check if there are even any changes before we start
	err := checkForChanges()
	if err != nil {
		return "", fmt.Errorf("could not save: %v", err)
	}

	// get stack title
	title, err := git.GetCurrentStackTitle()
	if err != nil {
		return "", fmt.Errorf("error running git command: %v", err)
	}

	stack, err := checkForStack(title)
	if err != nil {
		return "", fmt.Errorf("error checking for stack: %v", err)
	}

	if stack.Branch == "" {
		return "", fmt.Errorf("not currently in a stack - change to the branch you want to amend")
	}

	// a description is required, so ask if one is not provided
	if description == "" {
		err := prompt.AskQuestionWithInput(&description, "description", "How would you describe this change?", stack.Description, true)
		if err != nil {
			return "", fmt.Errorf("error prompting for title description: %v", err)
		}
	}

	// git add files
	_, err = addFiles(args[0:])
	if err != nil {
		return "", fmt.Errorf("error adding files: %v", err)
	}

	// run the amend commit
	err = gitAmend(description)
	if err != nil {
		return "", fmt.Errorf("error amending commit with git: %v", err)
	}

	output := fmt.Sprintf("Amended stack item with description: \"%s\".\n", description)
	if f.IO.IsOutputTTY() {
	}

	s.Stop()

	return output, nil
}

func gitAmend(description string) error {
	amendCmd := git.GitCommand("commit", "--amend", "-m", description)
	output, err := run.PrepareCmd(amendCmd).Output()
	if err != nil {
		return fmt.Errorf("error running git command: %v", err)
	}

	fmt.Println("Amend commit: ", string(output))

	return nil
}

func checkForStack(title string) (git.StackRef, error) {
	stack, err := git.GatherStackRefs(title)
	if err != nil {
		return git.StackRef{}, err
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		return git.StackRef{}, err
	}

	for _, ref := range stack.Refs {
		if ref.Branch == branch {
			return ref, nil
		}
	}

	return git.StackRef{}, nil
}
