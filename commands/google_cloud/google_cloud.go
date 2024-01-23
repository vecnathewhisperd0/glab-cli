package google_cloud

import (
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	googleCloudCreateWLIFCmd "gitlab.com/gitlab-org/cli/commands/google_cloud/wlif"
)

func NewCmdGoogleCloud(f *cmdutils.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "google-cloud <command> [flags]",
		Short: "EXPERIMENTAL: Manage Google Cloud integration of a GitLab project",
	}

	cmd.AddCommand(googleCloudCreateWLIFCmd.NewCmdCreateWLIF(f))

	return cmd
}
