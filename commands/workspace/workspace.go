package workspace

import (
	"github.com/MakeNowJust/heredoc"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
)

func NewCmdWorkspace(f *cmdutils.Factory) *cobra.Command {
	workspaceCmd := &cobra.Command{
		Use:   "workspace <command> [flags]",
		Short: `Create, view and manage workspaces`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab workspace list --group=gitlab-org 
			glab workspace create --group=gitlab-org --agent=1 -f devfile.yaml --editor=ttyd
			glab workspace view--group=gitlab-org 1
		`),
		Annotations: map[string]string{
			"help:arguments": heredoc.Doc(`
			`),
		},
	}

	cmdutils.EnableRepoOverride(workspaceCmd, f)

	workspaceCmd.AddCommand(NewCmdList(f))
	workspaceCmd.AddCommand(NewCmdCreate(f))
	workspaceCmd.AddCommand(NewCmdView(f))
	workspaceCmd.AddCommand(NewCmdUpdate(f))

	return workspaceCmd
}
