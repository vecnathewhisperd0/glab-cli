package list

import (
	"fmt"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var factory *cmdutils.Factory

func NewCmdEnvironmentCreate(f *cmdutils.Factory) *cobra.Command {
	factory = f
	environmentCreateCmd := &cobra.Command{
		Use:   "create [flags]",
		Short: `Create a project-level environment`,
		Long:  ``,
		Args:  cobra.MaximumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			factory = f
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			url, err := cmd.Flags().GetString("url")
			if err != nil {
				return err
			}
			return createEnvironment(name, url)
		},
	}
	environmentCreateCmd.Flags().StringP("name", "n", "", "Name of the new environment")
	environmentCreateCmd.Flags().StringP("url", "e", "", "External URL of the new environment")

	return environmentCreateCmd
}

func createEnvironment(name, url string) error {
	apiClient, err := factory.HttpClient()
	if err != nil {
		return err
	}

	repo, err := factory.BaseRepo()
	if err != nil {
		return err
	}

	_, err = api.CreateEnvironment(apiClient, repo.FullName(), &gitlab.CreateEnvironmentOptions{
		Name:        &name,
		ExternalURL: &url,
	})
	if err != nil {
		return err
	}

	fmt.Fprintf(factory.IO.StdOut, "%s\n%s\n", "environment", name)
	return nil
}
