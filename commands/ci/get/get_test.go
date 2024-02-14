package status

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"gitlab.com/gitlab-org/cli/commands/cmdtest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(rt http.RoundTripper, isTTY bool, args string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := cmdtest.InitIOStreams(isTTY, "")

	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := NewCmdGet(factory)

	return cmdtest.ExecuteCommand(cmd, args, stdout, stderr)
}

const (
	FILE_BODY   = 1
	INLINE_BODY = 2
)

func TestCIGet(t *testing.T) {
	type httpMock struct {
		method   string
		path     string
		status   int
		body     string
		bodyType int
	}

	tests := []struct {
		name            string
		args            string
		httpMocks       []httpMock
		expectedOut     string
		expectedOutType int
	}{
		{
			name: "when get is called on an existing pipeline",
			args: "-p=123 -b=main",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123",
					http.StatusOK,
					`{
						"id": 123,
						"iid": 123,
						"status": "pending",
						"source": "push",
						"ref": "main",
						"sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
						"user": {
							"username": "test"
						},
						"yaml_errors": "-",
						"created_at": "2023-10-10T00:00:00Z",
						"started_at": "2023-10-10T00:00:00Z",
						"updated_at": "2023-10-10T00:00:00Z"
					}`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123/jobs?per_page=100",
					http.StatusOK,
					`[]`,
					INLINE_BODY,
				},
			},
			expectedOut: `# Pipeline:
id:	123
status:	pending
source:	push
ref:	main
sha:	0ff3ae198f8601a285adcf5c0fff204ee6fba5fd
tag:	false
yaml Errors:	-
user:	test
created:	2023-10-10 00:00:00 +0000 UTC
started:	2023-10-10 00:00:00 +0000 UTC
updated:	2023-10-10 00:00:00 +0000 UTC

# Jobs:

`,
		},
		{
			name: "when get is called on missing pipeline",
			args: "-b=main",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123",
					http.StatusOK,
					`{
						"id": 123,
						"iid": 123,
						"status": "pending",
						"source": "push",
						"ref": "main",
						"sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
						"user": {
							"username": "test"
						},
						"yaml_errors": "-",
						"created_at": "2023-10-10T00:00:00Z",
						"started_at": "2023-10-10T00:00:00Z",
						"updated_at": "2023-10-10T00:00:00Z"
					}`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123/jobs?per_page=100",
					http.StatusOK,
					`[]`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/repository/commits/main",
					http.StatusOK,
					`{
						"last_pipeline" : {
							"id": 123
						}
					}`,
					INLINE_BODY,
				},
			},
			expectedOut: `# Pipeline:
id:	123
status:	pending
source:	push
ref:	main
sha:	0ff3ae198f8601a285adcf5c0fff204ee6fba5fd
tag:	false
yaml Errors:	-
user:	test
created:	2023-10-10 00:00:00 +0000 UTC
started:	2023-10-10 00:00:00 +0000 UTC
updated:	2023-10-10 00:00:00 +0000 UTC

# Jobs:

`,
		},
		{
			name: "when get is called on an existing pipeline with job text",
			args: "-p=123 -b=main",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123",
					http.StatusOK,
					`{
						"id": 123,
						"iid": 123,
						"status": "pending",
						"source": "push",
						"ref": "main",
						"sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
						"user": {
							"username": "test"
						},
						"yaml_errors": "-",
						"created_at": "2023-10-10T00:00:00Z",
						"started_at": "2023-10-10T00:00:00Z",
						"updated_at": "2023-10-10T00:00:00Z"
					}`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123/jobs?per_page=100",
					http.StatusOK,
					`[{
							"id": 123,
							"name": "publish",
							"status": "failed"
						}]`,
					INLINE_BODY,
				},
			},
			expectedOut: `# Pipeline:
id:	123
status:	pending
source:	push
ref:	main
sha:	0ff3ae198f8601a285adcf5c0fff204ee6fba5fd
tag:	false
yaml Errors:	-
user:	test
created:	2023-10-10 00:00:00 +0000 UTC
started:	2023-10-10 00:00:00 +0000 UTC
updated:	2023-10-10 00:00:00 +0000 UTC

# Jobs:
publish:	failed

`,
		},
		{
			name: "when get is called on an existing pipeline with job details",
			args: "-p=123 -b=main --with-job-details",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123",
					http.StatusOK,
					`{
						"id": 123,
						"iid": 123,
						"status": "pending",
						"source": "push",
						"ref": "main",
						"sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
						"user": {
							"username": "test"
						},
						"yaml_errors": "-",
						"created_at": "2023-10-10T00:00:00Z",
						"started_at": "2023-10-10T00:00:00Z",
						"updated_at": "2023-10-10T00:00:00Z"
					}`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123/jobs?per_page=100",
					http.StatusOK,
					`[{
							"id": 123,
							"name": "publish",
							"status": "failed",
							"failure_reason": "bad timing"
						}]`,
					INLINE_BODY,
				},
			},
			expectedOut: `# Pipeline:
id:	123
status:	pending
source:	push
ref:	main
sha:	0ff3ae198f8601a285adcf5c0fff204ee6fba5fd
tag:	false
yaml Errors:	-
user:	test
created:	2023-10-10 00:00:00 +0000 UTC
started:	2023-10-10 00:00:00 +0000 UTC
updated:	2023-10-10 00:00:00 +0000 UTC

# Jobs:
ID	Name	Status	Duration	Failure reason
123	publish	failed	0	bad timing

`,
		},
		{
			name: "when get is called on an existing pipeline with variables",
			args: "-p=123 -b=main --with-variables",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123",
					http.StatusOK,
					`{
						"id": 123,
						"iid": 123,
						"project_id": 5,
						"status": "pending",
						"source": "push",
						"ref": "main",
						"sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
						"user": {
							"username": "test"
						},
						"yaml_errors": "-",
						"created_at": "2023-10-10T00:00:00Z",
						"started_at": "2023-10-10T00:00:00Z",
						"updated_at": "2023-10-10T00:00:00Z"
					}`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123/jobs?per_page=100",
					http.StatusOK,
					`[]`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/5/pipelines/123/variables",
					http.StatusOK,
					`[{
						"key": "RUN_NIGHTLY_BUILD",
				    "variable_type": "env_var",
						"value": "true"
					}]`,
					INLINE_BODY,
				},
			},
			expectedOut: `# Pipeline:
id:	123
status:	pending
source:	push
ref:	main
sha:	0ff3ae198f8601a285adcf5c0fff204ee6fba5fd
tag:	false
yaml Errors:	-
user:	test
created:	2023-10-10 00:00:00 +0000 UTC
started:	2023-10-10 00:00:00 +0000 UTC
updated:	2023-10-10 00:00:00 +0000 UTC

# Jobs:

# Variables:
RUN_NIGHTLY_BUILD:	true

`,
		},
		{
			name: "when get is called on an existing pipeline with variables however no variables are found",
			args: "-p=123 -b=main --with-variables",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123",
					http.StatusOK,
					`{
						"id": 123,
						"iid": 123,
						"project_id": 5,
						"status": "pending",
						"source": "push",
						"ref": "main",
						"sha": "0ff3ae198f8601a285adcf5c0fff204ee6fba5fd",
						"user": {
							"username": "test"
						},
						"yaml_errors": "-",
						"created_at": "2023-10-10T00:00:00Z",
						"started_at": "2023-10-10T00:00:00Z",
						"updated_at": "2023-10-10T00:00:00Z"
					}`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/123/jobs?per_page=100",
					http.StatusOK,
					`[]`,
					INLINE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/5/pipelines/123/variables",
					http.StatusOK,
					"[]",
					INLINE_BODY,
				},
			},
			expectedOut: `# Pipeline:
id:	123
status:	pending
source:	push
ref:	main
sha:	0ff3ae198f8601a285adcf5c0fff204ee6fba5fd
tag:	false
yaml Errors:	-
user:	test
created:	2023-10-10 00:00:00 +0000 UTC
started:	2023-10-10 00:00:00 +0000 UTC
updated:	2023-10-10 00:00:00 +0000 UTC

# Jobs:

# Variables:
No variables found in pipeline.
`,
		},
		{
			name: "when requesting output as JSON",
			args: "-p 452959326 -F json -b main",
			httpMocks: []httpMock{
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/452959326",
					http.StatusOK,
					"testdata/ci_get-0.json",
					FILE_BODY,
				},
				{
					http.MethodGet,
					"/api/v4/projects/OWNER%2FREPO/pipelines/452959326/jobs?per_page=100",
					http.StatusOK,
					"testdata/ci_get-1.json",
					FILE_BODY,
				},
			},
			expectedOut:     "testdata/ci_get.result",
			expectedOutType: FILE_BODY,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fakeHTTP := &httpmock.Mocker{
				MatchURL: httpmock.PathAndQuerystring,
			}
			defer fakeHTTP.Verify(t)

			for _, mock := range tc.httpMocks {
				fakeHTTP.RegisterResponder(mock.method, mock.path, httpmock.NewStringResponse(mock.status, mock.body))
			}

			output, err := runCommand(fakeHTTP, false, tc.args)
			require.Nil(t, err)

			fmt.Printf("++>> %s\n", output.String())
			err = os.WriteFile("/tmp/expected", []byte(tc.expectedOut), 0o644)
			require.Nil(t, err)
			err = os.WriteFile("/tmp/received", []byte(output.String()), 0o644)
			require.Nil(t, err)
			assert.Equal(t, tc.expectedOut, output.String())
			assert.Empty(t, output.Stderr())
		})
	}
}
