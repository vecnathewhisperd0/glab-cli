package run

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/internal/config"

	"github.com/acarl005/stripansi"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/test"
)

type ResponseJSON struct {
	Ref string `json:"ref"`
}

func runCommand(rt http.RoundTripper, isTTY bool, cli string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := cmdtest.InitIOStreams(isTTY, "")

	factory := cmdtest.InitFactory(ios, rt)

	factory.Branch = func() (string, error) {
		return "custom-branch-123", nil
	}

	_, _ = factory.HttpClient()

	cmd := NewCmdRun(factory)

	return cmdtest.ExecuteCommand(cmd, cli, stdout, stderr)
}

func TestCIRun(t *testing.T) {
	tests := []struct {
		name string
		cli  string

		expectedPOSTBody string
		expectedOut      string
	}{
		{
			name:             "when running `ci run` without any parameter, defaults to current branch",
			cli:              "",
			expectedPOSTBody: fmt.Sprintf(`"ref":"%s"`, "custom-branch-123"),
			expectedOut:      fmt.Sprintf("Created pipeline (id: 123 ), status: created , ref: %s , weburl:  https://gitlab.com/OWNER/REPO/-/pipelines/123 )\n", "custom-branch-123"),
		},
		{
			name:             "when running `ci run` with branch parameter, run CI at branch",
			cli:              "-b ci-cd-improvement-399",
			expectedPOSTBody: `"ref":"ci-cd-improvement-399"`,
			expectedOut:      "Created pipeline (id: 123 ), status: created , ref: ci-cd-improvement-399 , weburl:  https://gitlab.com/OWNER/REPO/-/pipelines/123 )\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeHTTP := &httpmock.Mocker{
				MatchURL: httpmock.PathAndQuerystring,
			}
			defer fakeHTTP.Verify(t)

			fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/projects/OWNER/REPO/pipeline",
				func(req *http.Request) (*http.Response, error) {
					rb, _ := io.ReadAll(req.Body)

					var response ResponseJSON
					err := json.Unmarshal(rb, &response)
					if err != nil {
						fmt.Printf("Error when parsing response body %s\n", rb)
					}

					// ensure CLI runs CI on correct branch
					assert.Contains(t, string(rb), tc.expectedPOSTBody)
					resp, _ := httpmock.NewStringResponse(http.StatusOK, fmt.Sprintf(`{
 						"id": 123,
 						"iid": 123,
 						"project_id": 3,
 						"status": "created",
 						"ref": "%s",
            "web_url": "https://gitlab.com/OWNER/REPO/-/pipelines/123"}`, response.Ref))(req)
					return resp, nil
				},
			)

			output, _ := runCommand(fakeHTTP, false, tc.cli)

			out := output.String()

			assert.Equal(t, tc.expectedOut, out)
			assert.Empty(t, output.Stderr())
		})
	}
}

func TestCIRunMR(t *testing.T) {
	defer config.StubConfig(`---
hosts:
  gitlab.com:
    username: monalisa
    token: OTOKEN
`, "")()

	io, _, stdout, stderr := iostreams.Test()
	stubFactory := cmdtest.StubFactory("")
	stubFactory.IO = io
	stubFactory.IO.IsaTTY = true
	stubFactory.IO.IsErrTTY = true
	oldCreateMRPipeline := api.CreateMRPipeline

	api.CreateMRPipeline = func(client *gitlab.Client, projectID interface{}, mrID int) (*gitlab.PipelineInfo, error) {
		if projectID == "" || projectID == "WRONG_REPO" || projectID == "expected_err" || mrID == 0 {
			return nil, fmt.Errorf("error expected")
		}
		repo, err := stubFactory.BaseRepo()
		if err != nil {
			return nil, err
		}

		mrPipelineInfo := &gitlab.PipelineInfo{
			ID:        1234,
			IID:       1234,
			ProjectID: 123,
			Status:    "created",
			Ref:       "branch-name",
			WebURL:    "https://" + repo.RepoHost() + "/" + repo.FullName() + "/-/pipelines/1234",
		}
		return mrPipelineInfo, nil
	}

	tests := []struct {
		name string
		args string

		expectedPOSTBody string
		expectedOut      string
	}{
		{
			name:             "when running `ci run` with --mr parameter, run CI at merge request",
			args:             "--mr 1234",
			expectedPOSTBody: "",
			expectedOut:      "Created pipeline (id: 1234), status: created, ref: branch-name, weburl: https://gitlab.com/gitlab-org/cli/-/pipelines/1234\n",
		},
	}

	cmd := NewCmdRun(stubFactory)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			argv, err := shlex.Split(tc.args)
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs(argv)
			_, err = cmd.ExecuteC()
			if err != nil {
				t.Fatal(err)
			}

			out := stripansi.Strip(stdout.String())
			fmt.Println(out)
			assert.Equal(t, tc.expectedOut, out)
			assert.Equal(t, "", stderr.String())
		})
	}

	api.CreateMRPipeline = oldCreateMRPipeline
}
