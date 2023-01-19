package award_emoji

import (
	"errors"
	"fmt"

	"gitlab.com/gitlab-org/cli/commands/issue/issueutils"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/spf13/cobra"
	gitlab "github.com/xanzy/go-gitlab"
)

func NewCmdAwardEmoji(f *cmdutils.Factory) *cobra.Command {
	issueAwardEmojiCreateCmd := &cobra.Command{
		Use:     "award-emoji <issue-id>",
		Aliases: []string{"comment"},
		Short:   "Award an emoji to an issue on GitLab",
		Long:    ``,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			out := f.IO.StdOut

			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			issue, repo, err := issueutils.IssueFromArg(apiClient, f.BaseRepo, args[0])
			if err != nil {
				return err
			}

			name, _ := cmd.Flags().GetString("name")

			if name == "" {
				name = utils.Editor(utils.EditorOptions{
					Label:    "Award Emoji Name:",
					Help:     "Enter the award emoji's name. ",
					FileName: "ISSUE_AWARD_EMOJI_EDITMSG",
				})
			}

			if name == "" {
				return errors.New("aborted... Award Emoji name is empty")
			}

			emoji, err := api.CreateIssueAwardEmoji(apiClient, repo.FullName(), issue.IID, &gitlab.CreateAwardEmojiOptions{
				Name: name,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(out, "Added award emoji %d\n", emoji.AwardableID)
			return nil
		},
	}
	issueAwardEmojiCreateCmd.Flags().StringP("name", "n", "", "Award Emoji name")

	return issueAwardEmojiCreateCmd
}
