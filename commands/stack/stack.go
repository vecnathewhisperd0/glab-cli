package stack

import (
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	stackCreateCmd "gitlab.com/gitlab-org/cli/commands/stack/create"
	stackSaveCmd "gitlab.com/gitlab-org/cli/commands/stack/save"

	"github.com/spf13/cobra"
)

func NewCmdStack(f *cmdutils.Factory) *cobra.Command {
	stackCmd := &cobra.Command{
		Use:     "stack <command> [flags]",
		Short:   `Work with Stacked Diffs`,
		Long:    ``,
		Aliases: []string{"stacks"},
	}

	cmdutils.EnableRepoOverride(stackCmd, f)

	stackCmd.AddCommand(stackCreateCmd.NewCmdCreateStack(f))
	stackCmd.AddCommand(stackSaveCmd.NewCmdSaveStack(f))
	stackCmd.AddCommand(stackSaveCmd.NewCmdAmendStack(f))
	return stackCmd
}
