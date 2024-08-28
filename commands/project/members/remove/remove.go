package remove

import (
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"strconv"
)

func NewCmdRemove(f *cmdutils.Factory) *cobra.Command {
	membersRemove := &cobra.Command{
		Use:   "remove [username | ID]",
		Short: `Remove a user from a project`,
		Example: heredoc.Doc(`
glab repo members remove john.doe
glab repo members remove 123
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

			userID, err := userIdFromArgs(apiClient, args)
			if err != nil {
				return err
			}

			_, err = api.RemoveProjectMember(apiClient, repo.FullName(), userID)

			return err
		},
	}
	return membersRemove
}

func userIdFromArgs(client *gitlab.Client, args []string) (int, error) {
	user := args[0]
	if userID, err := strconv.Atoi(user); err == nil {
		return userID, nil
	}

	userByName, err := api.UserByName(client, user)
	if err != nil {
		return 0, err
	}
	return userByName.ID, nil
}
