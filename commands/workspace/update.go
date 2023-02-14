package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/hasura/go-graphql-client"
	"github.com/spf13/cobra"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/flag"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

type UpdateOptions struct {
	Group                string
	UpdateWorkspaceInput api.WorkspaceUpdateInput

	IO            *iostreams.IOStreams
	GraphQLClient *graphql.Client
}

func NewCmdUpdate(f *cmdutils.Factory) *cobra.Command {
	opts := &UpdateOptions{
		IO: f.IO,
	}

	workspaceUpdateCmd := &cobra.Command{
		Use:   "update [flags]",
		Short: `Update a workspace`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab workspace update --group=gitlab-org --workspaceId=1 --editor=ttyd -f devfile.yaml
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// supports repo override
			client, err := f.GraphQLClient()
			if err != nil {
				return err
			}
			opts.GraphQLClient = client

			workspaceId, _ := cmd.Flags().GetString("workspaceId")
			group, err := flag.GroupOverride(cmd)
			if err != nil {
				return err
			}

			status := new(string)
			*status, _ = cmd.Flags().GetString("status")
			if len(*status) == 0 {
				status = nil
			}

			editor := new(string)
			*editor, _ = cmd.Flags().GetString("editor")
			if len(*editor) == 0 {
				editor = nil
			}

			devfileContents := new(string)
			devfileLocation, _ := cmd.Flags().GetString("devfile")
			if len(devfileLocation) != 0 {
				rawData, err := os.ReadFile(devfileLocation)
				if err != nil {
					return err
				}
				*devfileContents = string(rawData)
			} else {
				devfileContents = nil
			}

			// return an error if nothing requires change
			if status == nil && editor == nil && devfileContents == nil {
				return errors.New("no changes to status, editor or devfile")
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			defer cancel()

			fullyQualifiedWorkspaceId := fmt.Sprintf(gidWorkspaceFormat, workspaceId)
			workspace, err := api.ViewWorkspace(
				ctx,
				client,
				group,
				fullyQualifiedWorkspaceId,
			)
			if err != nil {
				return err
			}

			updatePayload := api.WorkspaceUpdateInput{
				WorkspaceId:  api.RemoteDevelopmentWorkspaceID(workspace.ID),
				Editor:       workspace.Editor,
				Devfile:      workspace.Devfile,
				DesiredState: workspace.DesiredState,
			}

			if status != nil && *status != workspace.DesiredState {
				updatePayload.DesiredState = *status
			}

			if editor != nil && *editor != workspace.Editor {
				updatePayload.Editor = *editor
			}

			if devfileContents != nil && *devfileContents != workspace.Devfile {
				updatePayload.Devfile = *devfileContents
			}

			opts.UpdateWorkspaceInput = updatePayload
			return updateRun(ctx, opts)
		},
	}

	workspaceUpdateCmd.PersistentFlags().StringP("group", "g", "", "Select a group/subgroup. This option is ignored if a repo argument is set.")
	workspaceUpdateCmd.PersistentFlags().StringP("workspaceId", "i", "", "Set the ID of the workspace to update")
	workspaceUpdateCmd.Flags().StringP("editor", "e", "", "The editor to be injected")
	workspaceUpdateCmd.Flags().StringP("status", "s", "", "The desired status of the workspace")
	workspaceUpdateCmd.Flags().StringP("devfile", "f", "", "The path of the devfile")
	workspaceUpdateCmd.MarkFlagRequired("workspaceId")

	return workspaceUpdateCmd
}
func updateRun(ctx context.Context, opts *UpdateOptions) error {
	err := api.UpdateWorkspace(ctx, opts.GraphQLClient, opts.UpdateWorkspaceInput)
	if err != nil {
		return err
	}

	streams := opts.IO

	fmt.Fprintln(streams.StdOut, streams.Color().Green("Workspace successfully updated"))
	return nil
}
