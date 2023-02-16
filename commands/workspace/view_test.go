package workspace

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
)

func TestWorkspaceView(t *testing.T) {
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
			name:          "when unsupported output options are provided",
			args:          "-g=MyGroup --output=non-existent 1",
			httpMocks:     nil,
			expectedError: errors.New("unsupported output format: non-existent"),
		},
		{
			name: "when workspace passed in does not exist",
			args: "-g=MyGroup 1",
			httpMocks: []httpMock{
				{
					http.MethodPost,
					"/api/graphql",
					http.StatusOK,
					`{
						"data": {
							"group": {
								"workspaces": {
									"nodes": [								
									]
								}
							}
						}
					}`,
				},
			},
			expectedError: api.ErrWorkspaceNotFound,
			expectedOut:   "",
		},
		{
			name: "when the workspace passed in does exist",
			args: "-g=MyGroup 1",
			httpMocks: []httpMock{
				{
					http.MethodPost,
					"/api/graphql",
					http.StatusOK,
					`{
						"data": {
							"group": {
								"workspaces": {
									"nodes": [
										{
											"id": "123",
											"name": "test",
											"editor": "ttyd",
											"url": "http://something.remotedev.com",
											"actualState": "Running",
											"devfile": "test"
										}
									]
								}
							}
						}
					}`,
				},
			},
			expectedError: nil,
			expectedOut:   "Workspace: 123\nEditor: ttyd\nActual State: Running\nURL: http://something.remotedev.com\nDevfile:\ntest\n",
		},
		{
			name: "when the workspace exists and view is called with json output",
			args: "-g=MyGroup -o=json 1",
			httpMocks: []httpMock{
				{
					http.MethodPost,
					"/api/graphql",
					http.StatusOK,
					`{
						"data": {
							"group": {
								"workspaces": {
									"nodes": [
										{
											"id": "123",
											"name": "test",
											"editor": "ttyd",
											"url": "http://something.remotedev.com",
											"actualState": "Running",
											"devfile": "test"
										}
									]
								}
							}
						}
					}`,
				},
			},
			expectedError: nil,
			expectedOut:   "{\"ID\":\"123\",\"Name\":\"test\",\"Url\":\"http://something.remotedev.com\",\"Editor\":\"ttyd\",\"ActualState\":\"Running\",\"Devfile\":\"test\",\"DesiredState\":\"\"}",
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

			output, err := runCommand(NewCmdView, fakeHTTP, false, tc.args)
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
