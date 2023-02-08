package workspace

import (
	"fmt"
	"time"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"golang.org/x/net/context"

	"github.com/MakeNowJust/heredoc"
	"github.com/hasura/go-graphql-client"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/flag"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/spf13/cobra"
)

type ListOptions struct {
	Group string

	IO            *iostreams.IOStreams
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

	fmt.Fprintf(opts.IO.StdOut, "%s\n%s\n", title.Describe(), DisplayList(opts.IO, workspaces))

	return nil
}