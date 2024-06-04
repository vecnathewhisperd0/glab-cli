package edit

import (
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var (
	variablesToCreate []string
	variablesToEdit   []string
	variablesToDelete []string
)

func NewCmdEdit(f *cmdutils.Factory) *cobra.Command {
	scheduleEditCmd := &cobra.Command{
		Use:   "edit [flags]",
		Short: `Edit a pipeline.`,
		Example: heredoc.Doc(`
			glab schedule edit 10 --cron "0 * * * *" --description "Describe your pipeline here" --ref "main" --create-variable "foo:bar" --edit-variable "baz:baz" --delete-variable "qux"
		`),
		Long: ``,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			repo, err := f.BaseRepo()
			if err != nil {
				return err
			}

			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			scheduleId := int(id)

			opts := &gitlab.EditPipelineScheduleOptions{}

			description, _ := cmd.Flags().GetString("description")
			ref, _ := cmd.Flags().GetString("ref")
			cron, _ := cmd.Flags().GetString("cron")
			cronTimeZone, _ := cmd.Flags().GetString("cronTimeZone")
			active, _ := cmd.Flags().GetBool("active")
			variablesToCreate, _ = cmd.Flags().GetStringSlice("create-variable")
			variablesToEdit, _ = cmd.Flags().GetStringSlice("edit-variable")
			variablesToDelete, _ = cmd.Flags().GetStringSlice("delete-variable")

			if cmd.Flags().Lookup("active").Changed {
				opts.Active = &active
			}

			if description != "" {
				opts.Description = &description
			}

			if ref != "" {
				opts.Ref = &ref
			}

			if cron != "" {
				opts.Cron = &cron
			}

			if cronTimeZone != "" {
				opts.CronTimezone = &cronTimeZone
			}

			// skip API call if no changes are made
			if opts.Active != nil || opts.Description != nil || opts.Ref != nil || opts.Cron != nil || opts.CronTimezone != nil {
				_, err := api.EditSchedule(apiClient, repo.FullName(), scheduleId, opts)
				if err != nil {
					return err
				}
			}

			// create variables
			for _, v := range variablesToCreate {
				split := strings.SplitN(v, ":", 2)
				if len(split) != 2 {
					return fmt.Errorf("Invalid format for --create-variable: %s", v)
				}
				err = api.CreateScheduleVariable(apiClient, repo.FullName(), scheduleId, &gitlab.CreatePipelineScheduleVariableOptions{
					Key:   &split[0],
					Value: &split[1],
				})
				if err != nil {
					return err
				}
			}

			// edit variables
			for _, v := range variablesToEdit {
				split := strings.SplitN(v, ":", 2)
				if len(split) != 2 {
					return fmt.Errorf("Invalid format for --edit-variable: %s", v)
				}
				err = api.EditScheduleVariable(apiClient, repo.FullName(), scheduleId, split[0], &gitlab.EditPipelineScheduleVariableOptions{
					Value: &split[1],
				})
				if err != nil {
					return err
				}
			}

			// delete variables
			for _, v := range variablesToDelete {
				err = api.DeleteScheduleVariable(apiClient, repo.FullName(), scheduleId, v)
				if err != nil {
					return err
				}
			}

			fmt.Fprintln(f.IO.StdOut, "Edited schedule with ID", scheduleId)

			return nil
		},
	}

	scheduleEditCmd.Flags().String("description", "", "Description of the schedule")
	scheduleEditCmd.Flags().String("ref", "", "Target branch or tag")
	scheduleEditCmd.Flags().String("cron", "", "Cron interval pattern")
	scheduleEditCmd.Flags().String("cronTimeZone", "", "Cron timezone")
	scheduleEditCmd.Flags().Bool("active", true, "Whether or not the schedule is active")
	scheduleEditCmd.Flags().StringSliceVar(&variablesToCreate, "create-variable", []string{}, "Pass new variables to schedule in format <key>:<value>")
	scheduleEditCmd.Flags().StringSliceVar(&variablesToEdit, "edit-variable", []string{}, "Pass updated variables to schedule in format <key>:<value>")
	scheduleEditCmd.Flags().StringSliceVar(&variablesToDelete, "delete-variable", []string{}, "Pass variables you want to delete from schedule in format <key>")
	scheduleEditCmd.Flags().Lookup("active").DefValue = "to not change"

	return scheduleEditCmd
}
