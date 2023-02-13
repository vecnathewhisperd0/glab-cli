package workspace

import (
	"fmt"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/hasura/go-graphql-client"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/flag"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/utils"
)

const (
	defaultWatchIntervalWorkspaceList = 5 * time.Second
)

type ListOptions struct {
	Group string
	Watch bool

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

			watchCount, err := cmd.Flags().GetCount("watchWriter")
			if err != nil {
				return err
			}
			opts.Watch = watchCount > 0

			return listRun(opts)
		},
	}

	cmdutils.EnableRepoOverride(workspaceListCmd, f)
	workspaceListCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")
	workspaceListCmd.Flags().CountP("watchWriter", "w", "Watch for updates")

	return workspaceListCmd
}

func listRun(opts *ListOptions) error {
	fetchData := func() ([]api.Workspace, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		return api.ListWorkspaces(ctx, opts.GraphQLClient, opts.Group)
	}

	fetchAndRender := func() (string, error) {
		workspaces, err := fetchData()
		if err != nil {
			return "", err
		}

		title := utils.NewListTitle("Workspace")
		title.EmptyMessage = fmt.Sprintf("No workspaces were found for group %s", opts.Group)
		title.Page = 1
		title.CurrentPageTotal = len(workspaces)
		title.Total = len(workspaces)

		var output string
		output += fmt.Sprintf("%s\n", title.Describe())

		if len(workspaces) != 0 {
			output += fmt.Sprintf("%s\n", RenderWorkspaces(opts.IO, workspaces))
		}

		output += fmt.Sprintf("\nLatest data as of %s\n", time.Now().Format(time.Stamp))

		return output, nil
	}

	if opts.Watch {
		writer := newWatchWriter(opts.IO.StdOut, defaultWatchIntervalWorkspaceList)
		return writer.runRenderLoop(fetchAndRender)
	} else {
		toRender, err := fetchAndRender()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(opts.IO.StdOut, toRender)
		return err
	}
}
