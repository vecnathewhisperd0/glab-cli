package create

import (
	"fmt"
	"os"
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

func NewCmdSaveStack(f *cmdutils.Factory) *cobra.Command {
	stackSaveCmd := &cobra.Command{
		Use:   "save",
		Short: `Save your progress within stacked diff`,
		Long:  "\"save\" lets you add a branch of the stack with your current progress.\n",
		Example: heredoc.Doc(`
			glab save added_file
			glab save -m "added a function"`),
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
			_, err = git.GetCurrentStackTitle()
			if err != nil {
				return fmt.Errorf("error running git command: %v", err)
			}

			if f.IO.IsOutputTTY() {
				fmt.Fprintf(f.IO.StdOut, "Saved with message: \"%s\".\n", message)
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
