package todoutils

import (
	"github.com/xanzy/go-gitlab"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func TodoActionName(todo *gitlab.Todo) string {
	switch todo.ActionName {
	case "approval_required":
		return "Approval required"
	case "build_failed":
		return "Pipeline failed"
	case "directly_addressed":
		return "Mentioned"
	case "marked":
		return "Added todo"
	case "merge_train_removed":
		return "Removed from merge train"
	case "review_requested":
		return "Review requested"
	case "review_submitted":
		return "Review submitted"
	default:
		return cases.Title(language.English, cases.NoLower).String(string(todo.ActionName))
	}
}
