package todo

import (
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	todoListCmd "gitlab.com/gitlab-org/cli/commands/todo/list"
)

func NewCmdTodo(f *cmdutils.Factory) *cobra.Command {
	todoCmd := &cobra.Command{
		Use:   "todo <command> [flags]",
		Short: `List todos`,
		Long:  ``,
	}

	todoCmd.AddCommand(todoListCmd.NewCmdList(f, nil))
	return todoCmd
}
