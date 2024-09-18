package badge

import (
	badgeSetCmd "gitlab.com/gitlab-org/cli/commands/badge/set"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
)

func NewCmdBadge() *cobra.Command {
	badgeCmd := &cobra.Command{
		Use:   "badge",
		Short: "Manage badges",
		Long:  `Manage badges for your project.`,
	}

	cmdutils.EnableRepoOverride(scheduleCmd, f)

	// Add subcommands
	badgeCmd.AddCommand(set.NewCmdBadgeSet())

	return badgeCmd
}

