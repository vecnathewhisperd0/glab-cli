package badge

import (
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/badge/remove"
	"gitlab.com/gitlab-org/cli/commands/badge/set"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdBadge(f *cmdutils.Factory) *cobra.Command {
	badgeCmd := &cobra.Command{
		Use:   "badge",
		Short: "Manage project badges",
		Long:  `Work with GitLab project badges`,
	}

	badgeCmd.AddCommand(set.NewCmdSet(f))
	badgeCmd.AddCommand(remove.NewCmdRemove(f))

	return badgeCmd
}
