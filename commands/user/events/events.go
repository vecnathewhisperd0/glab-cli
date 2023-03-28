package events

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/utils"
)

func NewCmdEvents(f *cmdutils.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "View user events",
		Args:  cobra.MaximumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}

			repo, err := f.BaseRepo()
			if err != nil {
				return err
			}

			events, err := api.CurrentUserEvents(apiClient)
			if err != nil {
				return err
			}

			if err = f.IO.StartPager(); err != nil {
				return err
			}
			defer f.IO.StopPager()

			includeDate, _ := cmd.Flags().GetBool("include-dates")

			if lb, _ := cmd.Flags().GetBool("all"); lb {
				projects := make(map[int]*gitlab.Project)
				for _, e := range events {
					project, err := api.GetProject(apiClient, e.ProjectID)
					if err != nil {
						return err
					}
					projects[e.ProjectID] = project
				}

				title := utils.NewListTitle("User events")
				title.CurrentPageTotal = len(events)

				DisplayAllEvents(f.IO.StdOut, events, projects, includeDate)
				return nil
			}

			project, err := api.GetProject(apiClient, repo.FullName())
			if err != nil {
				return err
			}

			DisplayProjectEvents(f.IO.StdOut, events, project, includeDate)
			return nil
		},
	}

	cmd.Flags().BoolP("all", "a", false, "Get events from all projects")
	cmd.Flags().BoolP("include-dates", "d", false, "Include the date of each event")

	return cmd
}

func DisplayProjectEvents(w io.Writer, events []*gitlab.ContributionEvent, project *gitlab.Project, includeDate bool) {
	for _, e := range events {
		if e.ProjectID != project.ID {
			continue
		}
		printEvent(w, e, project, includeDate)
	}
}

func DisplayAllEvents(w io.Writer, events []*gitlab.ContributionEvent, projects map[int]*gitlab.Project, includeDate bool) {
	for _, e := range events {
		printEvent(w, e, projects[e.ProjectID], includeDate)
	}
}

func printEvent(w io.Writer, e *gitlab.ContributionEvent, project *gitlab.Project, includeDate bool) {
	dateString := "- "

	if includeDate {
		dateString = fmt.Sprintf("%s - ", e.CreatedAt.String())
	}

	switch e.ActionName {
	case "pushed to":
		fmt.Fprintf(w, "%sPushed to %s %s at %s\n%q\n", dateString, e.PushData.RefType, e.PushData.Ref, project.NameWithNamespace, e.PushData.CommitTitle)
	case "deleted":
		fmt.Fprintf(w, "%sDeleted %s %s at %s\n", dateString, e.PushData.RefType, e.PushData.Ref, project.NameWithNamespace)
	case "pushed new":
		fmt.Fprintf(w, "%sPushed new %s %s at %s\n", dateString, e.PushData.RefType, e.PushData.Ref, project.NameWithNamespace)
	case "commented on":
		fmt.Fprintf(w, "%sCommented on %s #%s at %s\n%q\n", dateString, e.Note.NoteableType, e.Note.Title, project.NameWithNamespace, e.Note.Body)
	case "accepted":
		fmt.Fprintf(w, "%sAccepted %s %s at %s\n", dateString, e.TargetType, e.TargetTitle, project.NameWithNamespace)
	case "opened":
		fmt.Fprintf(w, "%sOpened %s %s at %s\n", dateString, e.TargetType, e.TargetTitle, project.NameWithNamespace)
	case "closed":
		fmt.Fprintf(w, "%sClosed %s %s at %s\n", dateString, e.TargetType, e.TargetTitle, project.NameWithNamespace)
	case "joined":
		fmt.Fprintf(w, "%sJoined %s\n", dateString, project.NameWithNamespace)
	case "left":
		fmt.Fprintf(w, "%sLeft %s\n", dateString, project.NameWithNamespace)
	case "created":
		targetType := e.TargetType
		if e.TargetType == "WikiPage::Meta" {
			targetType = "%sWiki page"
		}
		fmt.Fprintf(w, "%sCreated %s %s at %s\n", dateString, targetType, e.TargetTitle, project.NameWithNamespace)
	default:
		fmt.Fprintf(w, "%s%s %q", dateString, e.TargetType, e.Title)
	}
	fmt.Fprintln(w) // to leave a blank line
}
