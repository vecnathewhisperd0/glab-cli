package create

import (
	"fmt"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
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
			color := f.IO.Color()

			title := utils.ReplaceNonAlphaNumericChars(titleString, "-")
			if title != titleString {
				fmt.Fprintf(f.IO.StdErr, "%s warning: non-usable characters have been replaced with dashes: %s\n",
					color.WarnIcon(),
					color.Blue(title))
			}

			err := git.SetLocalConfig("glab.currentstack", title)
			if err != nil {
				return fmt.Errorf("error setting local git config: %v", err)
			}

			_, err = git.AddStackRefDir(title)
			if err != nil {
				return fmt.Errorf("error adding stack metadata directory: %v", err)
			}

			if f.IO.IsOutputTTY() {
				fmt.Fprintf(f.IO.StdOut, "New stack created with title \"%s\".\n", title)
			}

			s.Stop()

			return nil
		},
	}
	return stackCreateCmd
}
