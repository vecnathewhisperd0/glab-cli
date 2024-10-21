package list

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

func NewCmdList(f *cmdutils.Factory) *cobra.Command {
	var outputFormat string
	var groupName string
	var state string

	iterationListCmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   `List iterations in the group.`,
		Long:    ``,
		Aliases: []string{"ls"},
		Example: heredoc.Doc(`
			glab iteration list
			glab iteration ls
			glab iteration list --group mygroup
			glab iteration list --state active --output json
		`),
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			l := &gitlab.ListGroupIterationsOptions{}

			if p, _ := cmd.Flags().GetInt("page"); p != 0 {
				l.Page = p
			}
			if p, _ := cmd.Flags().GetInt("per-page"); p != 0 {
				l.PerPage = p
			}
			if state != "" {
				l.State = gitlab.Ptr(state)
			}

			iterations, err := api.ListGroupIterations(apiClient, groupName, l)
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				iterationsJSON, _ := json.Marshal(iterations)
				fmt.Fprintln(f.IO.StdOut, string(iterationsJSON))
			} else {
				fmt.Fprintf(f.IO.StdOut, "Showing %d iterations for group %s.\n\n", len(iterations), groupName)
				for _, iteration := range iterations {
					fmt.Fprintf(f.IO.StdOut, "ID: %d, Title: %s, State: %s\n", iteration.ID, iteration.Title, iteration.State)
				}
			}

			return nil
		},
	}

	iterationListCmd.Flags().IntP("page", "p", 1, "Page number.")
	iterationListCmd.Flags().IntP("per-page", "P", 30, "Number of items to list per page.")
	iterationListCmd.Flags().StringVarP(&outputFormat, "output", "F", "text", "Format output as: text, json.")
	iterationListCmd.Flags().StringVarP(&groupName, "group", "g", "", "Name of the group")
	iterationListCmd.Flags().StringVarP(&state, "state", "s", "", "Filter iterations by state (active, upcoming, opened, closed, or all)")

	return iterationListCmd
}
