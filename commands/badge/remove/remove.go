package remove

import (
	"fmt"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

// create NewCmdRemove function that relies upon the DeleteBadge function
func NewCmdRemove(f *cmdutils.Factory) *cobra.Command {
	badgeRemoveCmd := &cobra.Command{
		Use:   "remove <badge-name>",
		Short: "Remove a badge from a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			projectID, err := cmd.Flags().GetInt("project-id")
			if err != nil {
				return err
			}

			if projectID <= 0 {
				return fmt.Errorf("error removing badge: --project-id is required")
			}

			badgeName := args[0]

			err = api.DeleteBadge(apiClient, projectID, badgeName)
			if err != nil {
				return fmt.Errorf("error removing badge: %v", err)
			}

			fmt.Fprintf(f.IO.StdOut, "Badge '%s' removed successfully\n", badgeName)
			return nil
		},
	}

	badgeRemoveCmd.Flags().Int("project-id", 0, "The ID of the project")
	_ = badgeRemoveCmd.MarkFlagRequired("project-id")

	return badgeRemoveCmd
}
