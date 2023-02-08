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

	"github.com/spf13/cobra"
)

const gidWorkspaceFormat = "gid://gitlab/RemoteDevelopment::Workspace/%s"

type ViewOptions struct {
	Group string
	ID    string

	IO            *iostreams.IOStreams
	GraphQLClient *graphql.Client
}

func NewCmdView(f *cmdutils.Factory) *cobra.Command {
	opts := &ViewOptions{
		IO: f.IO,
	}

	workspaceViewCmd := &cobra.Command{
		Use:   "view [flags]",
		Short: `View details for a workspace`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab workspace view --group-id="gitlab-org" 1
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// supports repo override
			client, err := f.GraphQLClient()
			if err != nil {
				return err
			}
			opts.GraphQLClient = client

			opts.ID = fmt.Sprintf(gidWorkspaceFormat, args[0])

			group, err := flag.GroupOverride(cmd)
			if err != nil {
				return err
			}
			opts.Group = group

			return viewRun(opts)
		},
	}

	cmdutils.EnableRepoOverride(workspaceViewCmd, f)
	workspaceViewCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")

	return workspaceViewCmd
}

func viewRun(opts *ViewOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	workspace, err := api.ViewWorkspace(ctx, opts.GraphQLClient, opts.Group, opts.ID)
	if err != nil {
		return err
	}

	DisplayWorkspace(opts.IO, workspace)

	return nil
}