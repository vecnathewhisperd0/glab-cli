package list

import (
	"net/http"
	"regexp"
	"testing"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/MakeNowJust/heredoc"

	"github.com/alecthomas/assert"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(rt http.RoundTripper) (*test.CmdOut, error) {
	ios, _, stdout, stderr := iostreams.Test()
	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := NewCmdList(factory)

	_, err := cmd.ExecuteC()
	return &test.CmdOut{
		OutBuf: stdout,
		ErrBuf: stderr,
	}, err
}

func TestCiList(t *testing.T) {
	fakeHTTP := httpmock.New()
	defer fakeHTTP.Verify(t)

	fakeHTTP.RegisterResponder("GET", "/api/v4/projects/OWNER/REPO/pipelines",
		httpmock.NewStringResponse(200, `
		[
			{
				"id": 1,
				"iid": 3,
				"project_id": 5,
				"sha": "c366255c71600e17519e802850ddcf7105d3cf66",
				"ref": "refs/merge-requests/1107/merge",
				"status": "success",
				"source": "merge_request_event",
				"created_at": "2020-12-01T01:15:50.559Z",
				"updated_at": "2020-12-01T01:36:41.737Z",
				"web_url": "https://gitlab.com/OWNER/REPO/-/pipelines/710046436"
			},
			{
				"id": 2,
				"iid": 4,
				"project_id": 5,
				"sha": "c9a7c0d9351cd1e71d1c2ad8277f3bc7e3c47d1f",
				"ref": "main",
				"status": "success",
				"source": "push",
				"created_at": "2020-11-30T18:20:47.571Z",
				"updated_at": "2020-11-30T18:39:40.092Z",
				"web_url": "https://gitlab.com/OWNER/REPO/-/pipelines/709793838"
			}
	]
	`))

	output, err := runCommand(fakeHTTP)
	if err != nil {
		t.Errorf("error running command `ci list`: %v", err)
	}

	out := output.String()
	timeRE := regexp.MustCompile(`\d+ years`)
	out = timeRE.ReplaceAllString(out, "X years")

	assert.Equal(t, heredoc.Doc(`
		Showing 2 pipelines on OWNER/REPO (Page 1)
		
		(success) • #1	refs/merge-requests/1107/merge	(about X years ago)
		(success) • #2	main	(about X years ago)
		
		`), out)
	assert.Empty(t, output.Stderr())
}
