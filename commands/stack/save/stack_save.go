package create

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/prompt"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

var message string

const StackLocation = "/.git/refs/stacked/"

type StackRef struct {
	Prev   string `json:"prev"`
	Branch string `json:"branch"`
	SHA    string `json:"sha"`
	Next   string `json:"next"`
	MR     string `json:"mr"`
}

func NewCmdSaveStack(f *cmdutils.Factory) *cobra.Command {
	stackSaveCmd := &cobra.Command{
		Use:   "save",
		Short: `Save your progress within stacked diff`,
		Long:  "\"save\" lets you save your current progress with a diff on the stack.\n",
		Example: heredoc.Doc(`
			glab stack save added_file
			glab stack save . -m "added a function"
			glab stack save -m "added a function"`),
		RunE: func(cmd *cobra.Command, args []string) error {
			// check if there are even any changes before we start
			err := checkForChanges()
			if err != nil {
				return fmt.Errorf("could not save: %v", err)
			}

			// a title is required, so ask if one is not provided
			if message == "" {
				err := prompt.AskQuestionWithInput(&message, "title", "What would you like to name this change?", "", true)
				if err != nil {
					return fmt.Errorf("error prompting for save title: %v", err)
				}
			}

			s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)

			// git add files
			_, err = addFiles(args[0:])
			if err != nil {
				return fmt.Errorf("error adding files: %v", err)
			}

			// get stack title
			title, err := git.GetCurrentStackTitle()
			if err != nil {
				return fmt.Errorf("error running git command: %v", err)
			}

			author, err := git.GitUserName()
			if err != nil {
				return fmt.Errorf("error getting git author: %v", err)
			}

			// generate a SHA based on: commit message, stack title, git author name
			sha, err := generateStackSha(message, title, string(author))
			if err != nil {
				return fmt.Errorf("error generating SHA command: %v", err)
			}

			// create branch name from SHA
			branch, err := createShaBranch(f, sha, title)
			if err != nil {
				return fmt.Errorf("error creating branch name: %v", err)
			}

			// create the branch prefix-stack_title-SHA
			err = git.CheckoutNewBranch(branch)
			if err != nil {
				return fmt.Errorf("error running branch checkout: %v", err)
			}

			// commit files to branch
			_, err = commitFiles(message)
			if err != nil {
				return fmt.Errorf("error committing files: %v", err)
			}

			// create stack metadata
			stackRef := StackRef{SHA: sha, Branch: branch}
			err = addStackRefFile(title, stackRef)
			if err != nil {
				return fmt.Errorf("error creating stack file: %v", err)
			}

			if f.IO.IsOutputTTY() {
				color := f.IO.Color()

				fmt.Fprintf(
					f.IO.StdOut,
					"%s %s: Saved with message: \"%s\".\n",
					color.ProgressIcon(),
					color.Blue(title),
					message,
				)
			}

			s.Stop()

			return nil
		},
	}
	stackSaveCmd.Flags().StringVarP(&message, "message", "m", "", "name the change")

	return stackSaveCmd
}

func checkForChanges() error {
	gitCmd := git.GitCommand("status", "--porcelain")
	output, err := run.PrepareCmd(gitCmd).Output()
	if err != nil {
		return fmt.Errorf("error running git status: %v", err)
	}

	if string(output) == "" {
		return fmt.Errorf("no changes to save")
	}

	return nil
}

func addFiles(args []string) (files []string, err error) {
	for _, file := range args {
		_, err = os.Stat(file)
		if err != nil {
			return
		}

		files = append(files, file)
	}

	cmdargs := append([]string{"add"}, args...)
	gitCmd := git.GitCommand(cmdargs...)

	_, err = run.PrepareCmd(gitCmd).Output()
	if err != nil {
		return []string{}, fmt.Errorf("error running git add: %v", err)
	}

	return files, err
}

func commitFiles(message string) (string, error) {
	commitCmd := git.GitCommand("commit", "-m", message)
	output, err := run.PrepareCmd(commitCmd).Output()
	if err != nil {
		return "", fmt.Errorf("error running git command: %v", err)
	}

	return string(output), nil
}

func generateStackSha(message string, title string, author string) (string, error) {
	toSha := message + title + author

	cmd := "echo \"" + toSha + "\" | git hash-object --stdin"
	sha, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("error running git hash-object: %v", err)
	}

	return strings.TrimSuffix(string(sha), "\n"), nil
}

func createShaBranch(f *cmdutils.Factory, sha string, title string) (string, error) {
	shortSha := string(sha)[0:8]

	cfg, err := f.Config()
	if err != nil {
		return "", fmt.Errorf("could not retrieve config file: %v", err)
	}

	prefix, err := cfg.Get("", "branch_prefix")
	if err != nil {
		return "", fmt.Errorf("could not get prefix config: %v", err)
	}

	if prefix == "" {
		prefix = os.Getenv("USER")
		if prefix == "" {
			prefix = "glab-stack"
		}
	}

	branchTitle := []string{prefix, title, shortSha}

	branch := strings.Join(branchTitle, "-")
	return string(branch), nil
}

func addStackRefFile(title string, stackRef StackRef) error {
	baseDir, err := git.ToplevelDir()
	if err != nil {
		return fmt.Errorf("error running git command: %v", err)
	}

	refDir := path.Join(baseDir, StackLocation, title)

	initialJsonData, err := json.Marshal(stackRef)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err)
	}

	if _, err = os.Stat(refDir); os.IsNotExist(err) {
		err = os.MkdirAll(refDir, 0o700) // create directory if it doesn't exist
		if err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}

	fullPath := path.Join(refDir, stackRef.SHA+".json")

	err = os.WriteFile(fullPath, initialJsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error running writing file: %v", err)
	}

	return nil
}

func updateStackRefFile(title string, s StackRef) error {
	baseDir, err := git.ToplevelDir()
	if err != nil {
		return fmt.Errorf("error running git command: %v", err)
	}

	refDir := path.Join(baseDir, StackLocation, title)

	fullPath := path.Join(refDir, s.SHA+".json")

	initialJsonData, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err)
	}
	err = os.WriteFile(fullPath, initialJsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error running writing file: %v", err)
	}

	return nil
}
