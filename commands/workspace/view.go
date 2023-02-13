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
)

const (
	gidWorkspaceFormat                = "gid://gitlab/RemoteDevelopment::Workspace/%s"
	defaultWatchIntervalWorkspaceView = 5 * time.Second
)

type ViewOptions struct {
	Group string
	ID    string
	Watch bool
	Json  bool

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

			watchCount, err := cmd.Flags().GetCount("watch")
			if err != nil {
				return err
			}
			opts.Watch = watchCount > 0

			jsonCount, err := cmd.Flags().GetCount("json")
			if err != nil {
				return err
			}
			opts.Json = jsonCount > 0

			return viewRun(opts)
		},
	}

	cmdutils.EnableRepoOverride(workspaceViewCmd, f)
	workspaceViewCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")
	workspaceViewCmd.Flags().CountP("watch", "w", "Watch for updates")
	workspaceViewCmd.Flags().Count("json", "Render data as json")

	return workspaceViewCmd
}

func viewRun(opts *ViewOptions) error {
	fetchAndRender := func() (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		workspace, err := api.ViewWorkspace(ctx, opts.GraphQLClient, opts.Group, opts.ID)
		if err != nil {
			return "", err
		}

		var output string

		// todo: watch mode doesn't particularly well with json
		// 	leads to errors in parsing and therefore rendering (maybe due to \n embedded in the devfiles being serialized)
		//
		//  need to discuss this
		if opts.Json {
			var raw []byte
			raw, err = json.Marshal(workspace)
			if err != nil {
				return "", err
			}

			output += string(raw)
		} else {
			output += RenderWorkspace(opts.IO, workspace)
			output += fmt.Sprintf("\nLatest data as of %s\n", time.Now().Format(time.Stamp))
		}

		return output, nil
	}

	if opts.Watch {
		writer := newWatchWriter(opts.IO.StdOut, defaultWatchIntervalWorkspaceView)
		return writer.runRenderLoop(fetchAndRender)
	} else {
		toRender, err := fetchAndRender()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(opts.IO.StdOut, toRender)
		return err
	}
	return nil
}
