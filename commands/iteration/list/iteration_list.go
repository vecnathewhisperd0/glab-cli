package list

import (
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
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
			// Retrieve the configuration object
			cfg, err := f.Config()
			if err != nil {
				return err
			}

			// Get the token from the configuration using the correct section and key
			token, err := cfg.Get("auth", "token")
			if err != nil || token == "" {
				return fmt.Errorf("failed to retrieve GitLab token from configuration")
			}

			// Create a new GitLab client using the token
			apiClient, err := gitlab.NewClient(token)
			if err != nil {
				return err
			}

			// Set up the options for listing iterations
			listOptions := &gitlab.ListGroupIterationsOptions{}

			if p, _ := cmd.Flags().GetInt("page"); p != 0 {
				listOptions.Page = p
			}
			if p, _ := cmd.Flags().GetInt("per-page"); p != 0 {
				listOptions.PerPage = p
			}
			if state != "" {
				listOptions.State = gitlab.Ptr(state)
			}

			// Fetch the iterations for the specified group
			iterations, _, err := apiClient.GroupIterations.ListGroupIterations(groupName, listOptions)
			if err != nil {
				return err
			}

			// Output the result in the specified format
			if outputFormat == "json" {
				iterationsJSON, _ := json.Marshal(iterations)
				fmt.Fprintln(f.IO.StdOut, string(iterationsJSON))
			} else {
				fmt.Fprintf(f.IO.StdOut, "Showing %d iterations for group %s.\n\n", len(iterations), groupName)
				for _, iteration := range iterations {
					fmt.Fprintf(f.IO.StdOut, "ID: %d, Title: %s, State: %d\n", iteration.ID, iteration.Title, iteration.State)
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
