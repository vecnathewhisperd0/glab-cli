package list

import (
	"net/http"
	"testing"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/MakeNowJust/heredoc"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(rt http.RoundTripper) (*test.CmdOut, error) {
	ios, _, stdout, stderr := iostreams.Test()
	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := NewCmdList(factory, nil)

	_, err := cmd.ExecuteC()
	return &test.CmdOut{
		OutBuf: stdout,
		ErrBuf: stderr,
	}, err
}

func TestTodoList(t *testing.T) {
	fakeHTTP := &httpmock.Mocker{}
	defer fakeHTTP.Verify(t)

	fakeHTTP.RegisterResponder(http.MethodGet, "/api/v4/todos",
		httpmock.NewStringResponse(http.StatusOK, `
	[
		{
			"id": 102,
			"project": {
				"path_with_namespace": "gitlab-org/gitlab-foss"
			},
			"action_name": "marked",
			"target": {
				"reference": "!1"
			},
			"target_url": "https://gitlab.example.com/gitlab-org/gitlab-foss/-/merge_requests/7"
		},
		{
			"id": 102,
			"project": {
				"path_with_namespace": "gitlab-org/gitlab-foss"
			},
			"action_name": "build_failed",
			"target": {
				"reference": "!1"
			},
			"target_url": "https://gitlab.example.com/gitlab-org/gitlab-foss/-/merge_requests/7"
		}
	]
	`))

	output, err := runCommand(fakeHTTP)
	if err != nil {
		t.Errorf("error running command `todo list`: %v", err)
	}

	out := output.String()

	assert.Equal(t, heredoc.Doc(`
		Showing 2 todos (Page 1)

		Added todo	gitlab-org/gitlab-foss!1	
		Pipeline failed	gitlab-org/gitlab-foss!1	

	`), out)
	assert.Empty(t, output.Stderr())
}
