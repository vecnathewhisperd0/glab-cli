package members

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	membersAddCmd "gitlab.com/gitlab-org/cli/commands/project/members/add"
	membersRemoveCmd "gitlab.com/gitlab-org/cli/commands/project/members/remove"
)

func NewCmdMembers(f *cmdutils.Factory) *cobra.Command {
	membersCmd := &cobra.Command{
		Use:   "members [command] [flags]",
		Short: `Manage project members.`,
		Long:  ``,
		Example: heredoc.Doc(`
glab repo members add john.doe --access-level=maintainer
glab repo members add 123 -a reporter
glab repo members remove john.doe
glab repo members remove 123`),
	}

	cmdutils.EnableRepoOverride(membersCmd, f)
	membersCmd.AddCommand(membersAddCmd.NewCmdAdd(f))
	membersCmd.AddCommand(membersRemoveCmd.NewCmdRemove(f))

	return membersCmd
}
