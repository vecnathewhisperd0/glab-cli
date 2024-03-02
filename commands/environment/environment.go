package cluster

import (
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	createCmd "gitlab.com/gitlab-org/cli/commands/environment/create"
)

func NewCmdEnvironment(f *cmdutils.Factory) *cobra.Command {
	environmentCmd := &cobra.Command{
		Use:   "environment <command> [flags]",
		Short: `Manage project-level Environments`,
		Long:  ``,
	}

	cmdutils.EnableRepoOverride(environmentCmd, f)

	environmentCmd.AddCommand(createCmd.NewCmdEnvironmentCreate(f))

	return environmentCmd
}
