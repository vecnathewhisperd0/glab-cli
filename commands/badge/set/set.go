package set

import (
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
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

			project := args[0]
			name := args[1]
			value := args[2]

			projectID, err := api.ProjectID(apiClient, project)
			if err != nil {
				return err
			}

			badge, err := api.UpdateProjectBadge(apiClient, projectID, name, value)
			if err != nil {
				return err
			}

			if badge != nil {
				fmt.Fprintf(f.IO.StdOut, "Badge '%s' set for project '%s'\n", name, project)
			}

			return nil
		},
	}

	return badgeSetCmd
}
