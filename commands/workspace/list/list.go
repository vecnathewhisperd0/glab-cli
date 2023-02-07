package list

import (
	"fmt"
	"time"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/tableprinter"
	"golang.org/x/net/context"

	"github.com/MakeNowJust/heredoc"
	"github.com/hasura/go-graphql-client"
	"gitlab.com/gitlab-org/cli/internal/glrepo"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/flag"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/spf13/cobra"
)

type ListOptions struct {
	Group string

	// Pagination
	Page    int
	PerPage int

	IO            *iostreams.IOStreams
	BaseRepo      func() (glrepo.Interface, error)
	GraphQLClient *graphql.Client
}

func NewCmdList(f *cmdutils.Factory) *cobra.Command {
	opts := &ListOptions{
		IO: f.IO,
	}

	workspaceListCmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   `List workspaces`,
		Long:    ``,
		Aliases: []string{"ls"},
		Example: heredoc.Doc(`
			glab workspace list --group-id="gitlab-org"
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// supports repo override
			opts.BaseRepo = f.BaseRepo

			client, err := f.GraphQLClient()
			if err != nil {
				return err
			}
			opts.GraphQLClient = client

			group, err := flag.GroupOverride(cmd)
			if err != nil {
				return err
			}
			opts.Group = group

			return listRun(opts)
		},
	}

	cmdutils.EnableRepoOverride(workspaceListCmd, f)
	workspaceListCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")

	return workspaceListCmd
}

func listRun(opts *ListOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	workspaces, err := api.ListWorkspaces(ctx, opts.GraphQLClient, opts.Group)
	if err != nil {
		return err
	}

	title := utils.NewListTitle("Workspace")
	title.Page = 1
	title.CurrentPageTotal = len(workspaces)
	title.Total = len(workspaces)
	// title.RepoName = repo.FullName()

	// if err = opts.IO.StartPager(); err != nil {
	// 	return err
	// }
	// defer opts.IO.StopPager()
	// // fmt.Fprintf(opts.IO.StdOut, "%s\n%s\n", title.Describe(), mrutils.DisplayAllMRs(opts.IO, mergeRequests))
	// fmt.Println(workspaces)

	fmt.Fprintf(opts.IO.StdOut, "%s\n%s\n", title.Describe(), DisplayList(opts.IO, workspaces))

	return nil
}

func DisplayList(streams *iostreams.IOStreams, workspaces []api.Workspace) string {
	c := streams.Color()
	table := tableprinter.NewTablePrinter()
	table.SetIsTTY(streams.IsOutputTTY())
	table.AddRow(c.Green("Id"), c.Green("Editor"), c.Green("Actual State"), c.Green("URL"))
	for _, workspace := range workspaces {
		table.AddCell(workspace.ID)
		table.AddCell(workspace.Editor)
		table.AddCell(workspace.ActualState)
		table.AddCell(workspace.Url)
		table.EndRow()
	}

	return table.Render()
}
