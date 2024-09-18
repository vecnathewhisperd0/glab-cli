func NewCmdSet(f *cmdutils.Factory) *cobra.Command {
	badgeSetCmd := &cobra.Command{
		Use:   "set <name> <value>",
		Short: "Set or update a badge for a project",
		Long: heredoc.Doc(`
			Set or update a badge for a project. If the badge doesn't exist, it will be created.
			If it already exists, it will be updated.
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			project, err := f.BaseRepo()
			if err != nil {
				return err
			}

			name := args[0]
			value := args[1]

			// Query existing badges
			badges, _, err := apiClient.ProjectBadges.ListProjectBadges(project.FullName(), &gitlab.ListProjectBadgesOptions{})
			if err != nil {
				return err
			}

			var existingBadge *gitlab.ProjectBadge
			for _, badge := range badges {
				if badge.Name == name {
					existingBadge = badge
					break
				}
			}

			imageURL := fmt.Sprintf("https://img.shields.io/badge/%s-%s-blue", name, value)

			if existingBadge == nil {
				// Create new badge
				badgeOptions := &gitlab.AddProjectBadgeOptions{
					LinkURL:  gitlab.String(""),
					ImageURL: gitlab.String(imageURL),
					Name:     gitlab.String(name),
				}
				_, _, err = apiClient.ProjectBadges.AddProjectBadge(project.FullName(), badgeOptions)
			} else {
				// Update existing badge
				badgeOptions := &gitlab.EditProjectBadgeOptions{
					LinkURL:  gitlab.String(""),
					ImageURL: gitlab.String(imageURL),
					Name:     gitlab.String(name),
				}
				_, _, err = apiClient.ProjectBadges.EditProjectBadge(project.FullName(), existingBadge.ID, badgeOptions)
			}

			if err != nil {
				return err
			}

			if existingBadge == nil {
				fmt.Printf("Badge '%s' created successfully\n", name)
			} else {
				fmt.Printf("Badge '%s' updated successfully\n", name)
			}

			return nil
		},
	}

	return badgeSetCmd
}

