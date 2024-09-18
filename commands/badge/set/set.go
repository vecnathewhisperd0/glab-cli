package set

import (
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdSet(f *cmdutils.Factory) *cobra.Command {
	badgeSetCmd := &cobra.Command{
		Use:   "set <project> <name> <value>",
		Short: "Set a badge for a project",
		Long: heredoc.Docf(`
			Set a badge for a project.

			The badge will be created if it doesn't exist, or updated if it already exists.
		`),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			// rewrite project := args[0] so that args[0] is an int
			projectID, err := cmdutils.ParseID(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse project ID: %w", err)
			}

			name := args[1]
			value := args[2]

			api.UpdateProjectBadge(apiClient, projectID, name, value)

			return nil
		},
	}

	return badgeSetCmd
}
