package workspace

import (
	"net/http"
	"testing"

	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(cmdCreateFn func(*cmdutils.Factory) *cobra.Command, rt http.RoundTripper, isTTY bool, cli string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := cmdtest.InitIOStreams(isTTY, "")

	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := cmdCreateFn(factory)

	return cmdtest.ExecuteCommand(cmd, cli, stdout, stderr)
}

func TestWorkspaceList(t *testing.T) {
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
			name: "when no workspaces exist and list is called",
			args: "-g=MyGroup",
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
			expectedError: nil,
			expectedOut:   "No workspaces were found for group MyGroup\n",
		},
		{
			name: "when workspaces exist and list is called",
			args: "-g=MyGroup",
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
			expectedOut:   "Showing 1 of 1 Workspace on  (Page 1)\n\nId\tEditor\tActual State\tURL\n123\tttyd\tRunning\thttp://something.remotedev.com\n\n",
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

			output, err := runCommand(NewCmdList, fakeHTTP, false, tc.args)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError, err, "error expected when running command `workspace list %s`", tc.args)
				return
			}

			require.Nil(t, err, "error running command `workspace list %s`: %v")

			require.Equal(t, tc.expectedOut, output.String())
			require.Empty(t, output.Stderr())
		})
	}
}
