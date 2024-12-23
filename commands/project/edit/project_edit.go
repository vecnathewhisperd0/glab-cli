package edit

import (
	"fmt"
	"path"
	"strings"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/internal/glrepo"
	"gitlab.com/gitlab-org/cli/pkg/git"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdEdit(f *cmdutils.Factory) *cobra.Command {
	projectCreateCmd := &cobra.Command{
		Use:   "edit [path] [flags]",
		Short: `Edit a GitLab project/repository.`,
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEditProject(cmd, args, f)
		},
		Example: heredoc.Doc(`
			# Update a repository with name and visibilty.
			$ glab repo edit glab-cli/my-project --private
	  `),
	}

	// TODO name should be optional and use the directory where glab is invoked
	projectCreateCmd.Flags().StringP("name", "n", "", "Name of the new project.")
	projectCreateCmd.Flags().Bool("internal", false, "Make project internal: visible to any authenticated user. Default.")
	projectCreateCmd.Flags().BoolP("private", "p", false, "Make project private: visible only to project members.")
	projectCreateCmd.Flags().BoolP("public", "P", false, "Make project public: visible without any authentication.")

	return projectCreateCmd
}

func runEditProject(cmd *cobra.Command, args []string, f *cmdutils.Factory) error {
	var (
		projectPath string
		visiblity   gitlab.VisibilityValue
		err         error
		namespace   string
		pid         string
	)
	c := f.IO.Color()

	apiClient, err := f.HttpClient()
	if err != nil {
		return err
	}

	if len(args) == 1 {
		var host string

		pid, host, namespace, projectPath = projectPathFromArgs(args)
		if host != "" {
			cfg, _ := f.Config()
			client, err := api.NewClientWithCfg(host, cfg, false)
			if err != nil {
				return err
			}
			apiClient = client.Lab()
		}
		user, err := api.CurrentUser(apiClient)
		if err != nil {
			return err
		}
		if user.Username == namespace {
			namespace = ""
		}
	} else {
		projectPath, err = git.ToplevelDir()
		projectPath = path.Base(projectPath)
		if err != nil {
			return err
		}
	}

	name, _ := cmd.Flags().GetString("name")

	if projectPath == "" && name == "" {
		fmt.Println("ERROR: path or name required to edit a project.")
		return cmd.Usage()
	} else if name == "" {
		name = projectPath
	}

	if internal, _ := cmd.Flags().GetBool("internal"); internal {
		visiblity = gitlab.InternalVisibility
	} else if private, _ := cmd.Flags().GetBool("private"); private {
		visiblity = gitlab.PrivateVisibility
	} else if public, _ := cmd.Flags().GetBool("public"); public {
		visiblity = gitlab.PublicVisibility
	}

	opts := &gitlab.EditProjectOptions{
		Name: gitlab.Ptr(name),
		Path: gitlab.Ptr(projectPath),
	}

	if visiblity != "" {
		opts.Visibility = &visiblity
	}

	// Deprecated/Legacy attribute
	opts.ContainerExpirationPolicyAttributes = nil

	project, err := api.EditProject(apiClient, pid, opts)

	greenCheck := c.Green("✓")

	if err == nil {
		fmt.Fprintf(f.IO.StdOut, "%s Updated repository %s on GitLab: %s\n", greenCheck, project.NameWithNamespace, project.WebURL)
		fmt.Fprintf(f.IO.StdOut, "   : visibility : %s\n", project.Visibility)
		fmt.Fprintf(f.IO.StdOut, "   : taglist    : %s\n", project.TagList)
		fmt.Fprintf(f.IO.StdOut, "   : created at : %s\n", project.CreatedAt)
	} else {
		return fmt.Errorf("error updating project: %v", err)
	}
	return err
}

func projectPathFromArgs(args []string) (pid, host, namespace, project string) {
	// sanitize input by removing trailing "/"
	project = strings.TrimSuffix(args[0], "/")

	if strings.Contains(project, "/") {
		pp, _ := glrepo.FromFullName(project)
		pid = pp.FullName()
		host = pp.RepoHost()
		project = pp.RepoName()
		namespace = pp.RepoNamespace()
	}
	return
}
