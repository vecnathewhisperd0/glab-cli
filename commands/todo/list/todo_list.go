package list

import (
	"fmt"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/tableprinter"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"

	"gitlab.com/gitlab-org/cli/commands/todo/todoutils"
)

type ListOptions struct {
	State  string

	// Pagination
	Page    int
	PerPage int

	// display opts
	TitleQualifier string

	IO         *iostreams.IOStreams
	HTTPClient func() (*gitlab.Client, error)
}

func DisplayAllTodos(streams *iostreams.IOStreams, todos []*gitlab.Todo) string {
	table := tableprinter.NewTablePrinter()
	table.SetIsTTY(streams.IsOutputTTY())
	for _, todo := range todos {
		table.AddCell(todoutils.TodoActionName(todo))
		table.AddCell(streams.Hyperlink(fmt.Sprintf("%s%s", todo.Project.PathWithNamespace, todo.Target.Reference), todo.TargetURL))
		table.AddCell(todo.Body)
		table.EndRow()
	}

	return table.Render()
}

func NewCmdList(f *cmdutils.Factory, runE func(opts *ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO: f.IO,
	}

	todoListCmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   `List your todos`,
		Long:    ``,
		Aliases: []string{"ls"},
		Example: heredoc.Doc(`
			glab todo list
			glab todo list --state done
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			opts.HTTPClient = f.HttpClient

			apiClient, err := opts.HTTPClient()
			if err != nil {
				return err
			}

			l := &gitlab.ListTodosOptions{}

			if p, _ := cmd.Flags().GetInt("page"); p != 0 {
				opts.Page = p
				l.Page = p
			}

			if p, _ := cmd.Flags().GetInt("per-page"); p != 0 {
				opts.PerPage = p
				l.PerPage = p
			}

			if state, _ := cmd.Flags().GetString("state"); state != "" {
				opts.State = state
				l.State = gitlab.String(state)
			}

			title := utils.NewListTitle(opts.TitleQualifier + " todo")
			todos, _, err := api.ListTodos(apiClient, l)

			if err != nil {
				return err
			}

			title.Page = l.Page
			title.CurrentPageTotal = len(todos)

			fmt.Fprintf(opts.IO.StdOut, "%s\n%s\n", title.Describe(), DisplayAllTodos(opts.IO, todos))

			return nil
		},
	}

	todoListCmd.Flags().IntP("page", "p", 1, "Page number")
	todoListCmd.Flags().IntP("per-page", "P", 30, "Number of items to list per page")
	todoListCmd.Flags().StringP("state", "s", "pending", "State of todo")

	return todoListCmd
}
