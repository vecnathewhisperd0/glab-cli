package workspace

import (
	"fmt"
	"os"
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

const DesiredStateRunning = "Running"

type CreateOptions struct {
	Group                string
	CreateWorkspaceInput api.WorkspaceCreateInput

	IO            *iostreams.IOStreams
	GraphQLClient *graphql.Client
}

func NewCmdCreate(f *cmdutils.Factory) *cobra.Command {
	opts := &CreateOptions{
		IO: f.IO,
	}

	workspaceCreateCmd := &cobra.Command{
		Use:   "create [flags]",
		Short: `Create a workspace`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab workspace create --group="gitlab-org" --editor=ttyd -f devfile.yaml
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// supports repo override
			client, err := f.GraphQLClient()
			if err != nil {
				return err
			}
			opts.GraphQLClient = client

			editor, _ := cmd.Flags().GetString("editor")
			agent, _ := cmd.Flags().GetString("agent")
			devfileLocation, _ := cmd.Flags().GetString("devfile")

			devfileContents, err := os.ReadFile(devfileLocation)
			if err != nil {
				return err
			}

			group, err := flag.GroupOverride(cmd)
			if err != nil {
				return err
			}
			opts.Group = group

			opts.CreateWorkspaceInput = api.WorkspaceCreateInput{
				GroupPath:      group,
				Editor:         editor,
				ClusterAgentID: fmt.Sprintf("gid://gitlab/Clusters::Agent/%s", agent),
				DesiredState:   DesiredStateRunning,
				Devfile:        string(devfileContents),
			}
			return createRun(opts)
		},
	}

	cmdutils.EnableRepoOverride(workspaceCreateCmd, f)
	workspaceCreateCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")
	workspaceCreateCmd.Flags().StringP("editor", "e", "", "The editor to be injected")
	workspaceCreateCmd.Flags().StringP("agent", "a", "", "The Id of the agent to use for provisioning")
	workspaceCreateCmd.Flags().StringP("devfile", "f", "", "The path of the devfile")
	workspaceCreateCmd.MarkFlagRequired("editor")
	workspaceCreateCmd.MarkFlagRequired("agent")
	workspaceCreateCmd.MarkFlagRequired("devfile")

	return workspaceCreateCmd
}

func createRun(opts *CreateOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	workspace, err := api.CreateWorkspace(ctx, opts.GraphQLClient, opts.CreateWorkspaceInput)
	if err != nil {
		return err
	}

	DisplayWorkspace(opts.IO, workspace)

	// fmt.Fprintf(opts.IO.StdOut, "%s\n%s\n", title.Describe(), DisplayList(opts.IO, workspaces))

	return nil
}