package add

import (
	"fmt"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"golang.org/x/exp/maps"
)

const (
	FlagAccessLevel = "access-level"
)

var AccessLevelMap = map[string]gitlab.AccessLevelValue{
	"no-access":  gitlab.NoPermissions,
	"minimal":    gitlab.MinimalAccessPermissions,
	"guest":      gitlab.GuestPermissions,
	"reporter":   gitlab.ReporterPermissions,
	"developer":  gitlab.DeveloperPermissions,
	"maintainer": gitlab.MaintainerPermissions,
	"owner":      gitlab.OwnerPermissions,
	"admin":      gitlab.AdminPermissions,
}

func getAccessLevelValue(level string) (gitlab.AccessLevelValue, error) {
	if val, ok := AccessLevelMap[strings.ToLower(level)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("invalid access level, must be one of: %s", strings.Join(maps.Keys(AccessLevelMap), ","))
}

func NewCmdAdd(f *cmdutils.Factory) *cobra.Command {
	membersAdd := &cobra.Command{
		Use:   "add [username | ID] [flags]",
		Short: `Add a user to a project`,
		Example: heredoc.Doc(`
# Add a user and specify a role, long form
$ glab repo members add john.doe --access-level=developer

# Add a user by ID and specify a role, short form
$ glab repo members add 123 -a reporter

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

			accessLevelInput, err := cmd.Flags().GetString(FlagAccessLevel)
			if err != nil {
				return err
			}
			accessLevelValue, err := getAccessLevelValue(accessLevelInput)
			if err != nil {
				return err
			}

			userID, err := api.UserIdFromArgs(apiClient, args)
			if err != nil {
				return err
			}

			c := &gitlab.AddProjectMemberOptions{
				UserID:      userID,
				AccessLevel: &accessLevelValue,
			}
			_, err = api.AddProjectMember(apiClient, repo.FullName(), c)
			if err == nil {
				fmt.Fprintf(f.IO.StdOut, "Added user %s to %s\n", args[0], repo.FullName())
			}

			return err
		},
	}
	SetupCommandFlags(membersAdd.Flags())

	return membersAdd
}

func SetupCommandFlags(flags *pflag.FlagSet) {
	keys := maps.Keys(AccessLevelMap)
	sort.Strings(keys)
	flags.StringP(FlagAccessLevel, "a", "", fmt.Sprintf("Access level of the user. Possible values are: %s", strings.Join(keys, ", ")))
}
