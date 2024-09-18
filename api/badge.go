package api

import "github.com/xanzy/go-gitlab"

func UpdateProjectBadge(client *gitlab.Client, projectID int, badgeName, badgeValue string) error {
    // List existing badges
    badges, _, err := client.ProjectBadges.ListProjectBadges(projectID, nil)
    if err != nil {
        return fmt.Errorf("error listing project badges: %v", err)
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
    badgeOptions := &gitlab.AddProjectBadgeOptions{
        LinkURL:  gitlab.String(fmt.Sprintf("https://img.shields.io/badge/%s-%s-blue", badgeName, badgeValue)),
        ImageURL: gitlab.String(fmt.Sprintf("https://img.shields.io/badge/%s-%s-blue", badgeName, badgeValue)),
        Name:     gitlab.String(badgeName),
    }

    if existingBadge == nil {
        // Create new badge
        _, _, err = client.ProjectBadges.AddProjectBadge(projectID, badgeOptions)
        if err != nil {
            return fmt.Errorf("error creating project badge: %v", err)
        }
    } else {
        // Update existing badge
        _, _, err = client.ProjectBadges.EditProjectBadge(projectID, existingBadge.ID, badgeOptions)
        if err != nil {
            return fmt.Errorf("error updating project badge: %v", err)
        }
    }

    return err
}
