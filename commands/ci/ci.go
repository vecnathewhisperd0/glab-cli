package ci

import (
	jobArtifactCmd "gitlab.com/gitlab-org/cli/commands/ci/artifact"
	pipeDeleteCmd "gitlab.com/gitlab-org/cli/commands/ci/delete"
	legacyCICmd "gitlab.com/gitlab-org/cli/commands/ci/legacyci"
	ciLintCmd "gitlab.com/gitlab-org/cli/commands/ci/lint"
	pipeListCmd "gitlab.com/gitlab-org/cli/commands/ci/list"
	pipeRetryCmd "gitlab.com/gitlab-org/cli/commands/ci/retry"
	pipeRunCmd "gitlab.com/gitlab-org/cli/commands/ci/run"
	pipeStatusCmd "gitlab.com/gitlab-org/cli/commands/ci/status"
	ciTraceCmd "gitlab.com/gitlab-org/cli/commands/ci/trace"
	ciViewCmd "gitlab.com/gitlab-org/cli/commands/ci/view"
	pipeWrapperCmd "gitlab.com/gitlab-org/cli/commands/ci/wrapper"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
)

func NewCmdCI(f *cmdutils.Factory) *cobra.Command {
	var ciCmd = &cobra.Command{
		Use:     "ci <command> [flags]",
		Short:   `Work with GitLab CI pipelines and jobs`,
		Long:    ``,
		Aliases: []string{"pipe", "pipeline"},
	}

	cmdutils.EnableRepoOverride(ciCmd, f)

	ciCmd.AddCommand(legacyCICmd.NewCmdCI(f))
	ciCmd.AddCommand(ciTraceCmd.NewCmdTrace(f, nil))
	ciCmd.AddCommand(ciViewCmd.NewCmdView(f))
	ciCmd.AddCommand(ciLintCmd.NewCmdLint(f))
	ciCmd.AddCommand(pipeDeleteCmd.NewCmdDelete(f))
	ciCmd.AddCommand(pipeListCmd.NewCmdList(f))
	ciCmd.AddCommand(pipeStatusCmd.NewCmdStatus(f))
	ciCmd.AddCommand(pipeRetryCmd.NewCmdRetry(f))
	ciCmd.AddCommand(pipeRunCmd.NewCmdRun(f))
	ciCmd.AddCommand(pipeWrapperCmd.NewCmdWrapper(f))
	ciCmd.AddCommand(jobArtifactCmd.NewCmdRun(f))
	return ciCmd
}
