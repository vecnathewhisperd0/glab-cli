package remove

import (
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdRemove(f *cmdutils.Factory) *cobra.Command {
	membersRemove := &cobra.Command{
		Use:   "remove [username | ID]",
		Short: `Remove a user from a project`,
		Example: heredoc.Doc(`
# Remove a user by name
$ glab repo members remove john.doe

# Remove a user by ID
$ glab repo members remove 123

`),
		Long: ``,
		Args: cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			repo, err := f.BaseRepo()
			if err != nil {
				return err
			}

			userID, err := api.UserIdFromArgs(apiClient, args)
			if err != nil {
				return err
			}

			_, err = api.RemoveProjectMember(apiClient, repo.FullName(), userID)
			if err == nil {
				fmt.Fprintf(f.IO.StdOut, "Removed user %s from %s\n", args[0], repo.FullName())
			}

			return err
		},
	}
	return membersRemove
}
