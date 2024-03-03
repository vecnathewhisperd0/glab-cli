package agentutils

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/tableprinter"
	"gitlab.com/gitlab-org/cli/pkg/utils"
)

func DisplayAllAgents(io *iostreams.IOStreams, agents []*gitlab.Agent) string {
	c := io.Color()
	table := tableprinter.NewTablePrinter()
	table.AddRow(c.Bold("ID"), c.Bold("Name"), c.Bold(c.Gray("Created At")))
	for _, r := range agents {
		table.AddRow(r.ID, r.Name, c.Gray(utils.TimeToPrettyTimeAgo(*r.CreatedAt)))
	}
	return table.Render()
}

func RunCommandToFile(cmd *exec.Cmd, p string) error {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return err
	}

	outFile, err := os.Create(p)
	if err != nil {
		return err
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	cmd.Wait()

	return nil
}

func KubectlApply(io *iostreams.IOStreams, path string) error {
	cmd := exec.Command("kubectl", "apply", "-f", path)
	cmd.Stdout = io.StdOut
	cmd.Stderr = io.StdErr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func WriteTemplateToFile(t *template.Template, p string, data interface{}) error {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return err
	}

	outFile, err := os.Create(p)
	if err != nil {
		return err
	}
	defer outFile.Close()

	err = t.Execute(outFile, data)
	if err != nil {
		return err
	}
	return nil
}

func DownloadFile(url string, p string, isTemp bool) (string, error) {
	var outFile io.Writer
	if isTemp {
		outFile, err := os.CreateTemp("", p)
		if err != nil {
			return "", err
		}
		p = outFile.Name()
		defer os.Remove(p)
		defer outFile.Close()
	} else {
		if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
			return "", err
		}
		outFile, err := os.Create(p)
		if err != nil {
			return "", err
		}
		defer outFile.Close()
	}

	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return "", err
	}
	return p, nil
}
