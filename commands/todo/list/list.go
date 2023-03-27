package list

import (
	"encoding/json"
	"fmt"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

func NewCmdList(f *cmdutils.Factory) *cobra.Command {
	todoListCmd := &cobra.Command{
		Use:   "list [flags]",
		Short: `Get the list of TODO items`,
		Example: heredoc.Doc(`
	glab ci list
	glab ci list --format=json
	`),
		Long: ``,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var all bool
			all = true

			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			l := &gitlab.ListTodosOptions{}

			if m, _ := cmd.Flags().GetString("state"); m != "" {
				l.State = gitlab.String(m)
			}
			if m, _ := cmd.Flags().GetString("type"); m != "" {
				l.Type = gitlab.String(m)
			}
			if m, _ := cmd.Flags().GetBool("global"); !m {
				repo, err := f.BaseRepo()
				if err != nil {
					return err
				}
				project, _ := repo.Project(apiClient)
				l.ProjectID = &project.ID
			}
			if p, _ := cmd.Flags().GetInt("page"); p != 0 {
				l.Page = p
				all = false
			}
			if p, _ := cmd.Flags().GetInt("per-page"); p != 0 {
				l.PerPage = p
				all = false
			}

			todos, err := api.ListTodos(apiClient, l, all)
			if err != nil {
				return err
			}

			if m, _ := cmd.Flags().GetString("output-format"); m == "text" {
				for _, t := range todos {
					var milestone string
					if t.Target.Milestone != nil {
						milestone = t.Target.Milestone.Title
					} else {
						milestone = "nil"
					}
					fmt.Fprintf(f.IO.StdOut, "%s\t%s\t%s\t%s\t%s\n", t.Target.UpdatedAt, milestone, t.TargetType, t.Target.Title, t.Target.WebURL)
				}
				// fmt.Fprintf(f.IO.StdOut, "%s\n", todos)
			} else {
				todoListJSON, _ := json.Marshal(todos)
				fmt.Fprintln(f.IO.StdOut, string(todoListJSON))
			}
			return nil
		},
	}
	todoListCmd.Flags().StringP("output-format", "F", "text", "Format output as: text, json")
	todoListCmd.Flags().BoolP("global", "g", false, "Global list of TODO")
	todoListCmd.Flags().StringP("state", "s", "pending", "State of TODO. One of: pending, done")
	todoListCmd.Flags().StringP("type", "t", "", "Type of TODO. One of: MergeRequest, Commit, Epic, DesignManagement::Design or AlertManagement::Alert")
	// TODO todoListCmd.Flags().StringP("author", "a", "text", "Author of TODO")
	// TODO todoListCmd.Flags().StringP("group", "g", "text", "Group of TODO")

	return todoListCmd
}
