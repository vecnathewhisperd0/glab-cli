package list

import (
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/issuable"
	issuableListCmd "gitlab.com/gitlab-org/cli/commands/issuable/list"
)

func NewCmdList(f *cmdutils.Factory, runE func(opts *issuableListCmd.ListOptions) error) *cobra.Command {
	cmd := issuableListCmd.NewCmdList(f, runE, issuable.TypeIssue)
	cmd.Flags().Int("iteration", 0, "Filter issues by iteration ID")
	return cmd
}
