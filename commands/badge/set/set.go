package set

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/internal/glrepo"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

type BadgeOptions struct {
	ProjectID int
	APIClient *gitlab.Client

	IO   *iostreams.IOStreams
	Repo glrepo.Interface
}

func NewCmdSet(f *cmdutils.Factory) *cobra.Command {
	badgeSetCmd := &cobra.Command{
		Use:   "set <badge-name> <badge-value>",
		Short: "Set a badge for a project",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			opts := &BadgeOptions{}

			badgeName := args[0]
			badgeValue := args[1]

			badge, err := api.CreateOrUpdateBadge(apiClient, opts.ProjectID, badgeName, badgeValue)
			if err != nil {
				return fmt.Errorf("error setting badge: %v", err)
			}

			if badge != nil {
				fmt.Fprintf(f.IO.StdOut, "Badge '%s' set successfully with value '%s'\n", badgeName, badgeValue)
			}

			return nil
		},
	}

	return badgeSetCmd
}
