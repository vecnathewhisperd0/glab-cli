package workspace

import (
	"encoding/json"
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
	Json  bool

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

			watchCount, err := cmd.Flags().GetCount("watch")
			if err != nil {
				return err
			}
			opts.Watch = watchCount > 0

			outputFormat, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}
			switch outputFormat {
			case "json":
				opts.Json = true
			case "": // default
			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}

			return listRun(opts)
		},
	}

	cmdutils.EnableRepoOverride(workspaceListCmd, f)
	workspaceListCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")
	workspaceListCmd.Flags().CountP("watch", "w", "Watch for updates")
	workspaceListCmd.Flags().StringP("output", "o", "", "Render data in different formats. Supported formats: json")

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

		if opts.Json {
			var raw []byte
			raw, err = json.Marshal(workspaces)
			if err != nil {
				return "", err
			}

			return string(raw), nil
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

		return output, nil
	}

	if opts.Watch {
		writer := newPollingWriter(opts.IO.StdOut, defaultWatchIntervalWorkspaceList)
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
