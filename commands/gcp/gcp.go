package variable

import (
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	gcpCreateIamPolicyCmd "gitlab.com/gitlab-org/cli/commands/gcp/create_iam_policy"
	gcpCreateWlifCmd "gitlab.com/gitlab-org/cli/commands/gcp/create_wlif"
)

func NewCmdGcp(f *cmdutils.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp <command> [flags]",
		Short: "Manage GCP integration of a GitLab project",
	}

	cmd.AddCommand(gcpCreateWlifCmd.NewCmdCreateWlif(f))
	cmd.AddCommand(gcpCreateIamPolicyCmd.NewCmdCreateIamPolicy(f))

	return cmd
}
