package incident

import (
	"github.com/MakeNowJust/heredoc"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	incidentCloseCmd "gitlab.com/gitlab-org/cli/commands/incident/close"
	incidentListCmd "gitlab.com/gitlab-org/cli/commands/incident/list"
	incidentReopenCmd "gitlab.com/gitlab-org/cli/commands/incident/reopen"
	incidentSubscribeCmd "gitlab.com/gitlab-org/cli/commands/incident/subscribe"
	incidentUnsubscribeCmd "gitlab.com/gitlab-org/cli/commands/incident/unsubscribe"
	incidentViewCmd "gitlab.com/gitlab-org/cli/commands/incident/view"

	"github.com/spf13/cobra"
)

func NewCmdIncident(f *cmdutils.Factory) *cobra.Command {
	incidentCmd := &cobra.Command{
		Use:   "incident [command] [flags]",
		Short: `Work with GitLab incidents`,
		Long:  ``,
		Example: heredoc.Doc(`
			glab incident list
		`),
		Annotations: map[string]string{
			"help:arguments": heredoc.Doc(`
				An incident can be supplied as argument in any of the following formats:
				- by number, e.g. "123"
				- by URL, e.g. "https://gitlab.com/NAMESPACE/REPO/-/issues/incident/123"
			`),
		},
	}

	cmdutils.EnableRepoOverride(incidentCmd, f)

	incidentCmd.AddCommand(incidentListCmd.NewCmdList(f, nil))
	incidentCmd.AddCommand(incidentViewCmd.NewCmdView(f))
	incidentCmd.AddCommand(incidentCloseCmd.NewCmdClose(f))
	incidentCmd.AddCommand(incidentReopenCmd.NewCmdReopen(f))
	incidentCmd.AddCommand(incidentSubscribeCmd.NewCmdSubscribe(f))
	incidentCmd.AddCommand(incidentUnsubscribeCmd.NewCmdUnsubscribe(f))
	return incidentCmd
}
