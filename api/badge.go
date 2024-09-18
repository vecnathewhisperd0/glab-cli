package api

import (
	"fmt"
	"net/url"

	"github.com/xanzy/go-gitlab"
)

var CreateOrUpdateBadge(client *gitlab.Client, projectID int, badgeName, badgeValue string) (*gitlab.ProjectBadge, error) {
	// List existing badges
	badges, _, err := client.ProjectBadges.ListProjectBadges(projectID, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing project badges: %w", err)
	}

	// Check if badge exists
	var existingBadge *gitlab.ProjectBadge
	for _, badge := range badges {
		if badge.Name == badgeName {
			existingBadge = badge
			break
		}
	}

	// Prepare badge options
	imageURL := fmt.Sprintf("https://img.shields.io/badge/%s-%s-blue", url.PathEscape(badgeName), url.PathEscape(badgeValue))
	badgeOptions := &gitlab.AddProjectBadgeOptions{
		LinkURL:  gitlab.String("https://example.com"), // You might want to customize this
		ImageURL: gitlab.String(imageURL),
		Name:     gitlab.String(badgeName),
	}

	var badge *gitlab.ProjectBadge

	if existingBadge == nil {
		// Create new badge
		badge, _, err = client.ProjectBadges.AddProjectBadge(projectID, badgeOptions)
		if err != nil {
			return nil, fmt.Errorf("error creating project badge: %w", err)
		}
	} else {
		// Update existing badge
		updateOptions := &gitlab.EditProjectBadgeOptions{
			LinkURL:  badgeOptions.LinkURL,
			ImageURL: badgeOptions.ImageURL,
			Name:     badgeOptions.Name,
		}
		badge, _, err = client.ProjectBadges.EditProjectBadge(projectID, existingBadge.ID, updateOptions)
		if err != nil {
			return nil, fmt.Errorf("error updating project badge: %w", err)
		}
	}

	return badge, nil
}
