package cluster

import (
	"github.com/spf13/cobra"
	agentBootstrapCmd "gitlab.com/gitlab-org/cli/commands/cluster/agent/bootstrap"
	agentGetTokenCmd "gitlab.com/gitlab-org/cli/commands/cluster/agent/get_token"
	agentListCmd "gitlab.com/gitlab-org/cli/commands/cluster/agent/list"
	agentUpdateKubeconfigCmd "gitlab.com/gitlab-org/cli/commands/cluster/agent/update_kubeconfig"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdAgent(f *cmdutils.Factory) *cobra.Command {
	agentCmd := &cobra.Command{
		Use:   "agent <command> [flags]",
		Short: `Manage GitLab Agents for Kubernetes`,
		Long:  ``,
	}

	cmdutils.EnableRepoOverride(agentCmd, f)

	agentCmd.AddCommand(agentBootstrapCmd.AgentBootstrapCmd(f))
	agentCmd.AddCommand(agentListCmd.NewCmdAgentList(f))
	agentCmd.AddCommand(agentGetTokenCmd.NewCmdAgentGetToken(f))
	agentCmd.AddCommand(agentUpdateKubeconfigCmd.NewCmdAgentUpdateKubeconfig(f))

	return agentCmd
}
