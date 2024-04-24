package create

import (
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/prompt"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdCreateStack(f *cmdutils.Factory) *cobra.Command {
	stackCreateCmd := &cobra.Command{
		Use:     "create",
		Short:   `Create new stacked diff`,
		Long:    ``,
		Aliases: []string{"new"},
		Example: heredoc.Doc(`
			glab stack create cool-new-feature
			glab stack new cool-new-feature
		`),
		Args: cobra.MaximumNArgs(10),
		RunE: func(cmd *cobra.Command, args []string) error {
			var titleString string

			if len(args) == 1 {
				titleString = args[0]
			} else if len(args) == 0 {
				err := prompt.AskQuestionWithInput(&titleString, "title", "New stack title?", "", true)
				if err != nil {
					return fmt.Errorf("error prompting for title: %v", err)
				}
			} else {
				titleString = strings.Join(args[:], "-")
			}

			s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)

			branch := utils.ReplaceNonAlphaNumericChars(titleString, "-")
			if branch != titleString {
				fmt.Fprintf(f.IO.StdErr, "\nwarning: non-usable characters have been replaced with dashes\n")
			}

			gitCmd := git.GitCommand("checkout", "-b", branch)
			_, err := run.PrepareCmd(gitCmd).Output()
			if err != nil {
				return fmt.Errorf("error running git command: %v", err)
			}

			if f.IO.IsOutputTTY() {
				fmt.Fprintf(f.IO.StdOut, "New stack created with branch \"%s\".\n", branch)
			}

			s.Stop()

			return nil
		},
	}
	return stackCreateCmd
}
