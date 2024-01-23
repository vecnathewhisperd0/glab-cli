package wlif

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(rt http.RoundTripper, isTTY bool, cli string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := cmdtest.InitIOStreams(isTTY, "")

	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := NewCmdCreateWLIF(factory)

	return cmdtest.ExecuteCommand(cmd, cli, stdout, stderr)
}

func TestNewCmdCreateWLIF(t *testing.T) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fakeHTTP := &httpmock.Mocker{
		MatchURL: httpmock.PathAndQuerystring,
	}
	defer fakeHTTP.Verify(t)

	fakeHTTP.RegisterResponder(http.MethodGet, "/api/v4/projects/OWNER/REPO/scripts/google_cloud/create_wlif?google_cloud_project_id=my_gcp_project",
		httpmock.NewStringResponse(http.StatusOK, `#!/bin/bash
echo "success from script"`))

	cli := `"OWNER/REPO" --google-cloud-project-id "my_gcp_project"`
	_, err := runCommand(fakeHTTP, false, cli)
	if err != nil {
		return
	}

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	assert.Contains(t, out, "success from script")
}
