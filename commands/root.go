package commands

import (
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	aliasCmd "gitlab.com/gitlab-org/cli/commands/alias"
	apiCmd "gitlab.com/gitlab-org/cli/commands/api"
	askCmd "gitlab.com/gitlab-org/cli/commands/ask"
	authCmd "gitlab.com/gitlab-org/cli/commands/auth"
	changelogCmd "gitlab.com/gitlab-org/cli/commands/changelog"
	pipelineCmd "gitlab.com/gitlab-org/cli/commands/ci"
	clusterCmd "gitlab.com/gitlab-org/cli/commands/cluster"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	completionCmd "gitlab.com/gitlab-org/cli/commands/completion"
	configCmd "gitlab.com/gitlab-org/cli/commands/config"
	"gitlab.com/gitlab-org/cli/commands/help"
	incidentCmd "gitlab.com/gitlab-org/cli/commands/incident"
	issueCmd "gitlab.com/gitlab-org/cli/commands/issue"
	jobCmd "gitlab.com/gitlab-org/cli/commands/job"
	labelCmd "gitlab.com/gitlab-org/cli/commands/label"
	mrCmd "gitlab.com/gitlab-org/cli/commands/mr"
	projectCmd "gitlab.com/gitlab-org/cli/commands/project"
	releaseCmd "gitlab.com/gitlab-org/cli/commands/release"
	scheduleCmd "gitlab.com/gitlab-org/cli/commands/schedule"
	snippetCmd "gitlab.com/gitlab-org/cli/commands/snippet"
	sshCmd "gitlab.com/gitlab-org/cli/commands/ssh-key"
	updateCmd "gitlab.com/gitlab-org/cli/commands/update"
	userCmd "gitlab.com/gitlab-org/cli/commands/user"
	variableCmd "gitlab.com/gitlab-org/cli/commands/variable"
	versionCmd "gitlab.com/gitlab-org/cli/commands/version"
	"gitlab.com/gitlab-org/cli/internal/glrepo"
)

// NewCmdRoot is the main root/parent command
func NewCmdRoot(f *cmdutils.Factory, version, buildDate string) *cobra.Command {
	c := f.IO.Color()
	rootCmd := &cobra.Command{
		Use:           "glab <command> <subcommand> [flags]",
		Short:         "A GitLab CLI tool.",
		Long:          `GLab is an open source GitLab CLI tool that brings GitLab to your command line.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Annotations: map[string]string{
			"help:environment": heredoc.Doc(`
			GITLAB_TOKEN: An authentication token for API requests. Set this variable to
			avoid prompts to authenticate. Overrides any previously-stored credentials.
			Can be set in the config with 'glab config set token xxxxxx'.

			GITLAB_HOST or GL_HOST: Specify the URL of the GitLab server if self-managed.
			(Example: https://gitlab.example.com) Defaults to https://gitlab.com.

			REMOTE_ALIAS or GIT_REMOTE_URL_VAR: A 'git remote' variable or alias that contains
			the GitLab URL. Can be set in the config with 'glab config set remote_alias origin'.

			VISUAL, EDITOR (in order of precedence): The editor tool to use for authoring text.
			Can be set in the config with 'glab config set editor vim'.

			BROWSER: The web browser to use for opening links.
			Can be set in the config with 'glab config set browser mybrowser'.

			GLAMOUR_STYLE: The environment variable to set your desired Markdown renderer style.
			Available options: dark, light, notty. To set a custom style, read
			https://github.com/charmbracelet/glamour#styles

			NO_PROMPT: Set to 1 (true) or 0 (false) to disable or enable prompts.

			NO_COLOR: Set to any value to avoid printing ANSI escape sequences for color output.

			FORCE_HYPERLINKS: Set to 1 to force hyperlinks in output, even when not outputting to a TTY.

			GLAB_CONFIG_DIR: Set to a directory path to override the global configuration location.
		`),
			"help:feedback": heredoc.Docf(`
			Encountered a bug or want to suggest a feature?
			Open an issue using '%s'
		`, c.Bold(c.Yellow("glab issue create -R gitlab-org/cli"))),
		},
	}

	rootCmd.SetOut(f.IO.StdOut)
	rootCmd.SetErr(f.IO.StdErr)

	rootCmd.PersistentFlags().Bool("help", false, "Show help for command")
	rootCmd.SetHelpFunc(func(command *cobra.Command, args []string) {
		help.RootHelpFunc(f.IO.Color(), command, args)
	})
	rootCmd.SetUsageFunc(help.RootUsageFunc)
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		if errors.Is(err, pflag.ErrHelp) {
			return err
		}
		return &cmdutils.FlagError{Err: err}
	})

	formattedVersion := versionCmd.Scheme(version, buildDate)
	rootCmd.SetVersionTemplate(formattedVersion)
	rootCmd.Version = formattedVersion

	// Child commands
	rootCmd.AddCommand(aliasCmd.NewCmdAlias(f))
	rootCmd.AddCommand(configCmd.NewCmdConfig(f))
	rootCmd.AddCommand(completionCmd.NewCmdCompletion(f.IO))
	rootCmd.AddCommand(versionCmd.NewCmdVersion(f.IO, version, buildDate))
	rootCmd.AddCommand(updateCmd.NewCheckUpdateCmd(f, version))
	rootCmd.AddCommand(authCmd.NewCmdAuth(f))

	// the commands below require apiClient and resolved repos
	f.BaseRepo = resolvedBaseRepo(f)
	cmdutils.HTTPClientFactory(f) // Initialize HTTP Client

	rootCmd.AddCommand(changelogCmd.NewCmdChangelog(f))
	rootCmd.AddCommand(clusterCmd.NewCmdCluster(f))
	rootCmd.AddCommand(issueCmd.NewCmdIssue(f))
	rootCmd.AddCommand(incidentCmd.NewCmdIncident(f))
	rootCmd.AddCommand(jobCmd.NewCmdJob(f))
	rootCmd.AddCommand(labelCmd.NewCmdLabel(f))
	rootCmd.AddCommand(mrCmd.NewCmdMR(f))
	rootCmd.AddCommand(pipelineCmd.NewCmdCI(f))
	rootCmd.AddCommand(projectCmd.NewCmdRepo(f))
	rootCmd.AddCommand(releaseCmd.NewCmdRelease(f))
	rootCmd.AddCommand(sshCmd.NewCmdSSHKey(f))
	rootCmd.AddCommand(userCmd.NewCmdUser(f))
	rootCmd.AddCommand(variableCmd.NewVariableCmd(f))
	rootCmd.AddCommand(apiCmd.NewCmdApi(f, nil))
	rootCmd.AddCommand(scheduleCmd.NewCmdSchedule(f))
	rootCmd.AddCommand(snippetCmd.NewCmdSnippet(f))
	rootCmd.AddCommand(askCmd.NewCmd(f))

	rootCmd.Flags().BoolP("version", "v", false, "show glab version information")
	return rootCmd
}

func resolvedBaseRepo(f *cmdutils.Factory) func() (glrepo.Interface, error) {
	return func() (glrepo.Interface, error) {
		httpClient, err := f.HttpClient()
		if err != nil {
			return nil, err
		}
		remotes, err := f.Remotes()
		if err != nil {
			return nil, err
		}
		repoContext, err := glrepo.ResolveRemotesToRepos(remotes, httpClient, "")
		if err != nil {
			return nil, err
		}
		baseRepo, err := repoContext.BaseRepo(f.IO.PromptEnabled())
		if err != nil {
			return nil, err
		}

		return baseRepo, nil
	}
}
