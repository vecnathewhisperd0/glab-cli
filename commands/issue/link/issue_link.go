package link

import (
	"fmt"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/issue/issueutils"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

func NewCmdLink(f *cmdutils.Factory) *cobra.Command {
	issueLinkCmd := &cobra.Command{
		Use:   "link <issue> <target-issue> [flags]",
		Short: `Link an issue to another issue`,
		Long:  ``,
		Example: heredoc.Doc(`
	glab issue link 42 43
	glab issue link 42 43 --target-project 1234
	`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			out := f.IO.StdOut
			c := f.IO.Color()

			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			issueLink, repo, err := issueutils.IssueFromArg(apiClient, f.BaseRepo, args[0])
			if err != nil {
				return err
			}
			l := &gitlab.CreateIssueLinkOptions{}

			// If the user didn't specify a target project, use the current project
			if !cmd.Flags().Changed("target-project") {
				l.TargetProjectID = gitlab.Ptr(repo.FullName())
			}

			l.TargetIssueIID = gitlab.Ptr(args[1])

			fmt.Fprintf(out, "Linking issue %d to issue %s in project %s\n", issueLink.IID, args[1], repo.FullName())

			issueLink, _, err = api.LinkIssues(apiClient, repo.FullName(), issueLink.IID, l)
			if err != nil {
				return err
			}

			fmt.Fprintln(out, issueutils.DisplayIssue(c, issueLink, f.IO.IsaTTY))
			return nil
		},
	}

	issueLinkCmd.Flags().StringP("target-project", "p", "", "The target project")

	return issueLinkCmd
}
