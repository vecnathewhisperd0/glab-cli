package api

import (
	"fmt"
	"net/url"

	"github.com/xanzy/go-gitlab"
)

func createOrUpdateBadge(client *gitlab.Client, projectID int, badgeName, badgeValue string) (*gitlab.ProjectBadge, error) {
	// List existing badges
	badges, _, err := client.ProjectBadges.ListProjectBadges(projectID, &gitlab.ListProjectBadgesOptions{})
	if err != nil {
		return nil, fmt.Errorf("error listing badges: %v", err)
	}

	// Check if badge exists
	var existingBadge *gitlab.ProjectBadge
	for _, badge := range badges {
		if badge.Name == badgeName {
			existingBadge = badge
			break
		}
	}

	// Prepare badge data
	imageURL := fmt.Sprintf("https://img.shields.io/badge/%s-%s-blue", url.PathEscape(badgeName), url.PathEscape(badgeValue))

	var badge *gitlab.ProjectBadge
	if existingBadge == nil {
		badgeOptions := &gitlab.AddProjectBadgeOptions{
			Name:     gitlab.Ptr(badgeName),
			ImageURL: gitlab.Ptr(imageURL),
		}
		// Create new badge
		badge, _, err = client.ProjectBadges.AddProjectBadge(projectID, badgeOptions)
		if err != nil {
			return nil, fmt.Errorf("error creating badge: %v", err)
		}
	} else {
		badgeOptions := &gitlab.EditProjectBadgeOptions{
			Name:     gitlab.Ptr(badgeName),
			ImageURL: gitlab.Ptr(imageURL),
		}
		// Update existing badge
		badge, _, err = client.ProjectBadges.EditProjectBadge(projectID, existingBadge.ID, badgeOptions)
		if err != nil {
			return nil, fmt.Errorf("error updating badge: %v", err)
		}
	}

	return badge, nil
}
