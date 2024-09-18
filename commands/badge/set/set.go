package set

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/glrepo"
)

func NewCmdSet(f *cmdutils.Factory) *cobra.Command {
	badgeSetCmd := &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Set or update a project badge",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			repo, err := f.BaseRepo()
			if err != nil {
				return err
			}

			name := args[0]
			value := args[1]

			return setBadge(apiClient, repo, name, value)
		},
	}

	return badgeSetCmd
}

func setBadge(apiClient *api.Client, repo glrepo.Interface, name, value string) error {
	project := repo.FullName()

	// First, check if the badge exists
	badges, err := api.ListProjectBadges(apiClient, project, &api.ListProjectBadgesOptions{})
	if err != nil {
		return err
	}

	var existingBadge *api.Badge
	for _, badge := range badges {
		if badge.Name == name {
			existingBadge = badge
			break
		}
	}

	imageURL := fmt.Sprintf("https://img.shields.io/badge/%s-%s-blue", url.PathEscape(name), url.PathEscape(value))

	if existingBadge == nil {
		// Badge doesn't exist, create a new one
		_, err = api.AddProjectBadge(apiClient, project, &api.AddProjectBadgeOptions{
			ImageURL: &imageURL,
			Name:     &name,
		})
		if err != nil {
			return fmt.Errorf("failed to create badge: %w", err)
		}
		fmt.Printf("Badge '%s' created successfully\n", name)
	} else {
		// Badge exists, update it
		_, err = api.EditProjectBadge(apiClient, project, existingBadge.ID, &api.EditProjectBadgeOptions{
			ImageURL: &imageURL,
			Name:     &name,
		})
		if err != nil {
			return fmt.Errorf("failed to update badge: %w", err)
		}
		fmt.Printf("Badge '%s' updated successfully\n", name)
	}

	return nil
}
