package workspace

import (
	"fmt"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/tableprinter"
)

func RenderWorkspaces(streams *iostreams.IOStreams, workspaces []api.Workspace) string {
	c := streams.Color()
	table := tableprinter.NewTablePrinter()
	table.SetIsTTY(streams.IsOutputTTY())
	table.AddRow(c.Green("Id"), c.Green("Editor"), c.Green("Actual State"), c.Green("URL"))
	for _, workspace := range workspaces {
		table.AddCell(workspace.ID)
		table.AddCell(workspace.Editor)
		table.AddCell(GetStatusWithColor(c, workspace.ActualState))
		table.AddCell(workspace.Url)
		table.EndRow()
	}

	return table.Render()
}

func DisplayWorkspace(streams *iostreams.IOStreams, workspace *api.Workspace) {
	c := streams.Color()

	fmt.Fprintf(streams.StdOut, "%s: %s\n", c.Bold("Workspace"), workspace.ID)
	fmt.Fprintf(streams.StdOut, "%s: %s\n", c.Bold("Editor"), workspace.Editor)
	fmt.Fprintf(streams.StdOut, "%s: %s\n", c.Bold("Actual State"), GetStatusWithColor(c, workspace.ActualState))
	fmt.Fprintf(streams.StdOut, "%s: %s\n", c.Bold("URL"), workspace.Url)
	fmt.Fprintf(streams.StdOut, "%s:\n%s\n", c.Bold("Devfile"), workspace.Devfile)
}

func GetStatusWithColor(cp *iostreams.ColorPalette, status string) string {

	switch status {
	case "Running":
		return cp.Green(status)
	case "Stopped":
		return cp.Red(status)
	case "Terminated":
		return cp.Gray(status)
	case "Failed":
		return cp.Red(status)
	}

	return status
}
