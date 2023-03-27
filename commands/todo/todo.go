package todo

import (
	"github.com/MakeNowJust/heredoc"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
	todoListCmd "gitlab.com/gitlab-org/cli/commands/todo/list"
)

func NewCmdTodo(f *cmdutils.Factory) *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "todo [command] [flags]",
		Short: `Work with GitLab todo`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab todo list
		`),
	}

	cmdutils.EnableRepoOverride(issueCmd, f)

	issueCmd.AddCommand(todoListCmd.NewCmdList(f))
	return issueCmd
}
