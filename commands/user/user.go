package user

import (
	"github.com/spf13/cobra"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	userEventsCmd "gitlab.com/gitlab-org/cli/commands/user/events"
)

func NewCmdUser(f *cmdutils.Factory) *cobra.Command {
	var userCmd = &cobra.Command{
		Use:   "user <command> [flags]",
		Short: "Interact with user",
		Long:  "",
	}

	userCmd.AddCommand(userEventsCmd.NewCmdEvents(f))

	return userCmd
}
