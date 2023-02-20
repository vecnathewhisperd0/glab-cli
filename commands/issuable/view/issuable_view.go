package view

import (
	"fmt"
	"strings"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"gitlab.com/gitlab-org/cli/commands/issuable"
	"gitlab.com/gitlab-org/cli/commands/issue/issueutils"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

type ViewOpts struct {
	ShowComments   bool
	ShowSystemLogs bool
	OpenInBrowser  bool
	Web            bool

	CommentPageNumber int
	CommentLimit      int

	Notes []*gitlab.Note
	Issue *gitlab.Issue

	IO *iostreams.IOStreams
}

func NewCmdView(f *cmdutils.Factory, issueType issuable.IssueType) *cobra.Command {
	examplePath := "issues/123"

	if issueType == issuable.TypeIncident {
		examplePath = "issues/incident/123"
	}

	opts := &ViewOpts{
		IO: f.IO,
	}
	issueViewCmd := &cobra.Command{
		Use:     "view <id>",
		Short:   fmt.Sprintf(`Display the title, body, and other information about an %s.`, issueType),
		Long:    ``,
		Aliases: []string{"show"},
		Example: heredoc.Doc(fmt.Sprintf(`
			glab %[1]s view 123
			glab %[1]s show 123
			glab %[1]s view --web 123
			glab %[1]s view --comments 123
			glab %[1]s view https://gitlab.com/NAMESPACE/REPO/-/%s
		`, issueType, examplePath)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			apiClient, err := f.HttpClient()
			if err != nil {
				return err
			}
			cfg, _ := f.Config()

			issue, baseRepo, err := issueutils.IssueFromArg(apiClient, f.BaseRepo, args[0])
			if err != nil {
				return err
			}

			opts.Issue = issue

			// Issues and incidents are the same kind, but with different issueType.
			// `issue view` can display issues of all types including incidents
			// `incident view` on the other hand, should display only incidents, and treat all other issue types as not found
			//
			// When using `incident view` with non incident's IDs, print an error.
			if issueType == issuable.TypeIncident && *opts.Issue.IssueType != string(issuable.TypeIncident) {
				fmt.Fprintln(opts.IO.StdErr, "Incident not found, but an issue with the provided ID exists. Run `glab issue view <id>` to view it.")
				return nil
			}

			// open in browser if --web flag is specified
			if opts.Web {
				if f.IO.IsaTTY && f.IO.IsErrTTY {
					fmt.Fprintf(opts.IO.StdErr, "Opening %s in your browser.\n", utils.DisplayURL(opts.Issue.WebURL))
				}

				browser, _ := cfg.Get(baseRepo.RepoHost(), "browser")
				return utils.OpenInBrowser(opts.Issue.WebURL, browser)
			}

			if opts.ShowComments {
				l := &gitlab.ListIssueNotesOptions{
					Sort: gitlab.String("asc"),
				}
				if opts.CommentPageNumber != 0 {
					l.Page = opts.CommentPageNumber
				}
				if opts.CommentLimit != 0 {
					l.PerPage = opts.CommentLimit
				}
				opts.Notes, err = api.ListIssueNotes(apiClient, baseRepo.FullName(), opts.Issue.IID, l)
				if err != nil {
					return err
				}
			}

			glamourStyle, _ := cfg.Get(baseRepo.RepoHost(), "glamour_style")
			f.IO.ResolveBackgroundColor(glamourStyle)
			err = f.IO.StartPager()
			if err != nil {
				return err
			}
			defer f.IO.StopPager()
			if f.IO.IsErrTTY && f.IO.IsaTTY {
				return printTTYIssuePreview(opts)
			}
			return printRawIssuePreview(opts)
		},
	}

	issueViewCmd.Flags().BoolVarP(&opts.ShowComments, "comments", "c", false, fmt.Sprintf("Show %s comments and activities", issueType))
	issueViewCmd.Flags().BoolVarP(&opts.ShowSystemLogs, "system-logs", "s", false, "Show system activities / logs")
	issueViewCmd.Flags().BoolVarP(&opts.Web, "web", "w", false, fmt.Sprintf("Open %s in a browser. Uses default browser or browser specified in BROWSER variable", issueType))
	issueViewCmd.Flags().IntVarP(&opts.CommentPageNumber, "page", "p", 1, "Page number")
	issueViewCmd.Flags().IntVarP(&opts.CommentLimit, "per-page", "P", 20, "Number of items to list per page")

	return issueViewCmd
}

func labelsList(opts *ViewOpts) string {
	return strings.Join(opts.Issue.Labels, ", ")
}

func assigneesList(opts *ViewOpts) string {
	assignees := utils.Map(opts.Issue.Assignees, func(a *gitlab.IssueAssignee) string {
		return a.Username
	})

	return strings.Join(assignees, ", ")
}

func issueState(opts *ViewOpts, c *iostreams.ColorPalette) (state string) {
	if opts.Issue.State == "opened" {
		state = c.Green("open")
	} else if opts.Issue.State == "locked" {
		state = c.Blue(opts.Issue.State)
	} else {
		state = c.Red(opts.Issue.State)
	}

	return
}

func printTTYIssuePreview(opts *ViewOpts) error {
	c := opts.IO.Color()
	issueTimeAgo := utils.TimeToPrettyTimeAgo(*opts.Issue.CreatedAt)
	// Header
	fmt.Fprint(opts.IO.StdOut, issueState(opts, c))
	fmt.Fprintf(opts.IO.StdOut, c.Gray(" • opened by %s %s\n"), opts.Issue.Author.Username, issueTimeAgo)
	fmt.Fprint(opts.IO.StdOut, c.Bold(opts.Issue.Title))
	fmt.Fprintf(opts.IO.StdOut, c.Gray(" #%d"), opts.Issue.IID)
	fmt.Fprintln(opts.IO.StdOut)

	// Description
	if opts.Issue.Description != "" {
		opts.Issue.Description, _ = utils.RenderMarkdown(opts.Issue.Description, opts.IO.BackgroundColor())
		fmt.Fprintln(opts.IO.StdOut, opts.Issue.Description)
	}

	fmt.Fprintf(opts.IO.StdOut, c.Gray("\n%d upvotes • %d downvotes • %d comments\n"), opts.Issue.Upvotes, opts.Issue.Downvotes, opts.Issue.UserNotesCount)

	// Meta information
	if labels := labelsList(opts); labels != "" {
		fmt.Fprint(opts.IO.StdOut, c.Bold("Labels: "))
		fmt.Fprintln(opts.IO.StdOut, labels)
	}
	if assignees := assigneesList(opts); assignees != "" {
		fmt.Fprint(opts.IO.StdOut, c.Bold("Assignees: "))
		fmt.Fprintln(opts.IO.StdOut, assignees)
	}
	if opts.Issue.Milestone != nil {
		fmt.Fprint(opts.IO.StdOut, c.Bold("Milestone: "))
		fmt.Fprintln(opts.IO.StdOut, opts.Issue.Milestone.Title)
	}
	if opts.Issue.State == "closed" {
		fmt.Fprintf(opts.IO.StdOut, "Closed By: %s %s\n", opts.Issue.ClosedBy.Username, issueTimeAgo)
	}

	// Comments
	if opts.ShowComments {
		fmt.Fprintln(opts.IO.StdOut, heredoc.Doc(` 
			--------------------------------------------
			Comments / Notes
			--------------------------------------------
			`))
		if len(opts.Notes) > 0 {
			for _, note := range opts.Notes {
				if note.System && !opts.ShowSystemLogs {
					continue
				}
				createdAt := utils.TimeToPrettyTimeAgo(*note.CreatedAt)
				fmt.Fprint(opts.IO.StdOut, note.Author.Username)
				if note.System {
					fmt.Fprintf(opts.IO.StdOut, " %s ", note.Body)
					fmt.Fprintln(opts.IO.StdOut, c.Gray(createdAt))
				} else {
					body, _ := utils.RenderMarkdown(note.Body, opts.IO.BackgroundColor())
					fmt.Fprint(opts.IO.StdOut, " commented ")
					fmt.Fprintf(opts.IO.StdOut, c.Gray("%s\n"), createdAt)
					fmt.Fprintln(opts.IO.StdOut, utils.Indent(body, " "))
				}
				fmt.Fprintln(opts.IO.StdOut)
			}
		} else {
			fmt.Fprintf(opts.IO.StdOut, "There are no comments on this %s\n", *opts.Issue.IssueType)
		}
	}

	fmt.Fprintf(opts.IO.StdOut, c.Gray("\nView this %s on GitLab: %s\n"), *opts.Issue.IssueType, opts.Issue.WebURL)

	return nil
}

func printRawIssuePreview(opts *ViewOpts) error {
	assignees := assigneesList(opts)
	labels := labelsList(opts)

	fmt.Fprintf(opts.IO.StdOut, "title:\t%s\n", opts.Issue.Title)
	fmt.Fprintf(opts.IO.StdOut, "state:\t%s\n", issueState(opts, opts.IO.Color()))
	fmt.Fprintf(opts.IO.StdOut, "author:\t%s\n", opts.Issue.Author.Username)
	fmt.Fprintf(opts.IO.StdOut, "labels:\t%s\n", labels)
	fmt.Fprintf(opts.IO.StdOut, "comments:\t%d\n", opts.Issue.UserNotesCount)
	fmt.Fprintf(opts.IO.StdOut, "assignees:\t%s\n", assignees)
	if opts.Issue.Milestone != nil {
		fmt.Fprintf(opts.IO.StdOut, "milestone:\t%s\n", opts.Issue.Milestone.Title)
	}

	fmt.Fprintln(opts.IO.StdOut, "--")
	fmt.Fprintln(opts.IO.StdOut, opts.Issue.Description)
	return nil
}
