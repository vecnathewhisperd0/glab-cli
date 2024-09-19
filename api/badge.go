package api

import (
	"fmt"
	"net/url"

	"github.com/xanzy/go-gitlab"
)

func CreateOrUpdateBadge(client *gitlab.Client, projectID int, badgeName, badgeValue string) (*gitlab.ProjectBadge, error) {
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
			LinkURL:  gitlab.Ptr(imageURL),
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
			LinkURL:  gitlab.Ptr(imageURL),
		}
		// Update existing badge
		badge, _, err = client.ProjectBadges.EditProjectBadge(projectID, existingBadge.ID, badgeOptions)
		if err != nil {
			return nil, fmt.Errorf("error updating badge: %v", err)
		}
	}

	return badge, nil
}

// create a func to delete a badget given the project id and badge name
func DeleteBadge(client *gitlab.Client, projectID int, badgeName string) error {
	// List existing badges
	badges, _, err := client.ProjectBadges.ListProjectBadges(projectID, &gitlab.ListProjectBadgesOptions{})
	if err != nil {
		return fmt.Errorf("error listing badges: %v", err)
	}

	// Find the badge with the given name
	var badgeID int
	for _, badge := range badges {
		if badge.Name == badgeName {
			badgeID = badge.ID
			break
		}
	}

	if badgeID == 0 {
		return fmt.Errorf("badge with name '%s' not found", badgeName)
	}

	// Delete the badge
	_, err = client.ProjectBadges.DeleteProjectBadge(projectID, badgeID)
	if err != nil {
		return fmt.Errorf("error deleting badge: %v", err)
	}

	return nil
}
