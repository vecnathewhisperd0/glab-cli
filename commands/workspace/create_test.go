package workspace

import (
	"net/http"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
)

func TestWorkspaceCreate(t *testing.T) {
	type httpMock struct {
		method string
		path   string
		status int
		body   string
	}

	tests := []struct {
		name      string
		args      string
		httpMocks []httpMock

		expectedError error
		expectedOut   string
	}{
		{
			name: "when valid workspace details are passed in",
			args: "-g=MyGroup --editor ttyd --agent 4 --devfile='testdata/devfile.yaml'",
			httpMocks: []httpMock{
				{
					http.MethodPost,
					"/api/graphql",
					http.StatusOK,
					`{
						"data": {
							"workspaceCreate": {
								"workspace": {
									"id": "123",
									"url": null,
									"editor": "ttyd",
									"devfile": "test",
									"actualState": "CreationRequested"
								}
							}
						}
					}`,
				},
			},
			expectedError: nil,
			expectedOut:   "Workspace: 123\nEditor: ttyd\nActual State: CreationRequested\nURL: \nDevfile:\ntest\n",
		},
		{
			name: "when workspace creation fails",
			args: "-g=MyGroup --editor ttyd --agent 4 --devfile='testdata/devfile.yaml'",
			httpMocks: []httpMock{
				{
					http.MethodPost,
					"/api/graphql",
					http.StatusOK,
					`{
						"errors": [{
							"message": "Error Message 1",
							"raisedAt": "Test"
						}]
					}`,
				},
			},
			expectedError: graphql.Errors(graphql.Errors{
				graphql.Error{
					Message:    "Error Message 1",
					Extensions: map[string]interface{}(nil),
					Locations: []struct {
						Line   int "json:\"line\""
						Column int "json:\"column\""
					}(nil),
				},
			}),
			expectedOut: "",
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

			output, err := runCommand(NewCmdCreate, fakeHTTP, false, tc.args)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError, err, "error expected when running command `workspace list %s`", tc.args)
				return
			}

			require.Nil(t, err)

			require.Equal(t, tc.expectedOut, output.String())
			require.Empty(t, output.Stderr())
		})
	}
}
