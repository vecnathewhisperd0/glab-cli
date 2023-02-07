package workspace

import (
	"github.com/MakeNowJust/heredoc"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"

	"gitlab.com/gitlab-org/cli/commands/workspace/list"
)

func NewCmdWorkspace(f *cmdutils.Factory) *cobra.Command {
	workspaceCmd := &cobra.Command{
		Use:   "workspace <command> [flags]",
		Short: `Create, view and manage workspaces`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab workspace list --group=gitlab-org 
			glab workspace create --group=gitlab-org --agent=1 -f devfile.yaml --editor=ttyd
			glab workspace get --group=gitlab-org --id=1
		`),
		Annotations: map[string]string{
			"help:arguments": heredoc.Doc(`
			`),
		},
	}

	cmdutils.EnableRepoOverride(workspaceCmd, f)

	workspaceCmd.AddCommand(list.NewCmdList(f))
	// workspaceCmd.AddCommand(mrApproversCmd.NewCmdApprovers(f))
	// workspaceCmd.AddCommand(mrCheckoutCmd.NewCmdCheckout(f))
	// workspaceCmd.AddCommand(mrCloseCmd.NewCmdClose(f))
	// workspaceCmd.AddCommand(mrCreateCmd.NewCmdCreate(f, nil))
	// workspaceCmd.AddCommand(mrDeleteCmd.NewCmdDelete(f))
	// workspaceCmd.AddCommand(mrDiffCmd.NewCmdDiff(f, nil))
	// workspaceCmd.AddCommand(mrForCmd.NewCmdFor(f))
	// workspaceCmd.AddCommand(mrIssuesCmd.NewCmdIssues(f))
	// workspaceCmd.AddCommand(mrListCmd.NewCmdList(f, nil))
	// workspaceCmd.AddCommand(mrMergeCmd.NewCmdMerge(f))
	// workspaceCmd.AddCommand(mrNoteCmd.NewCmdNote(f))
	// workspaceCmd.AddCommand(mrRebaseCmd.NewCmdRebase(f))
	// workspaceCmd.AddCommand(mrReopenCmd.NewCmdReopen(f))
	// workspaceCmd.AddCommand(mrRevokeCmd.NewCmdRevoke(f))
	// workspaceCmd.AddCommand(mrSubscribeCmd.NewCmdSubscribe(f))
	// workspaceCmd.AddCommand(mrUnsubscribeCmd.NewCmdUnsubscribe(f))
	// workspaceCmd.AddCommand(mrTodoCmd.NewCmdTodo(f))
	// workspaceCmd.AddCommand(mrUpdateCmd.NewCmdUpdate(f))
	// workspaceCmd.AddCommand(mrViewCmd.NewCmdView(f))

	return workspaceCmd
}
