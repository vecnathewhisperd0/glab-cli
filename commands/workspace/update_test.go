package workspace

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
)

type WorkspaceUpdateTestSuite struct {
	suite.Suite
}

func TestWorkspaceUpdateCmd(t *testing.T) {
	suite.Run(t, new(WorkspaceUpdateTestSuite))
}

func (w *WorkspaceUpdateTestSuite) TestCases() {

	tests := []struct {
		name string
		args string

		expectedGraphqlCalls []httpmock.Stub
		expectedError        error
	}{
		{
			name: "when at least 1 valid params is passed in",
			args: "--group=test-group --workspaceId=1 --editor ttyd --status=Running",
			expectedGraphqlCalls: []httpmock.Stub{
				w.mockGetWorkspaceQuery(api.Workspace{
					Editor:       "vscode",
					Devfile:      "",
					DesiredState: "Stopped",
				}),
				w.mockUpdateWorkspaceMutation(nil),
			},
			expectedError: nil,
		},
		{
			name:                 "when no params are provided",
			args:                 "--group=test-group --workspaceId=1",
			expectedGraphqlCalls: nil,
			expectedError:        errors.New("no changes to status, editor or devfile"),
		},
		{
			name: "when an error is returned by the server during mutation",
			args: "--group=test-group --workspaceId=1 --editor ttyd --status=Running",
			expectedGraphqlCalls: []httpmock.Stub{
				w.mockGetWorkspaceQuery(api.Workspace{
					Editor:       "vscode",
					Devfile:      "",
					DesiredState: "Stopped",
				}),
				w.mockUpdateWorkspaceMutation([]string{"internal server error"}),
			},
			expectedError: errors.New("internal server error"),
		},
	}

	for _, tc := range tests {
		w.T().Run(tc.name, func(t *testing.T) {
			fakeHTTP := &httpmock.Mocker{
				MatchURL: httpmock.PathAndQuerystring,
			}
			defer fakeHTTP.Verify(t)

			for _, stub := range tc.expectedGraphqlCalls {
				fakeHTTP.RegisterStub(stub)
			}

			output, err := runCommand(NewCmdUpdate, fakeHTTP, false, tc.args)
			if tc.expectedError != nil {
				require.Equal(t, tc.expectedError, err, "error expected when running command `workspace list %s`", tc.args)
				return
			}

			require.Nil(t, err)
			require.Empty(t, output.Stderr())
		})
	}
}

func (w *WorkspaceUpdateTestSuite) graphqlMutationMatcher() httpmock.Matcher {
	return func(req *http.Request) bool {
		query, err := w.extractGraphqlQuery(req)
		if err != nil {
			w.T().Errorf("Error while extracting graphql query: %s", err.Error())
			return false
		}

		return strings.HasPrefix(query, "mutation")
	}
}

func (w *WorkspaceUpdateTestSuite) graphqlQueryMatcher() httpmock.Matcher {
	return func(req *http.Request) bool {
		query, err := w.extractGraphqlQuery(req)
		if err != nil {
			w.T().Errorf("Error while extracting graphql query: %s", err.Error())
			return false
		}

		return strings.HasPrefix(query, "query")
	}
}

func (w *WorkspaceUpdateTestSuite) extractGraphqlQuery(req *http.Request) (string, error) {
	reader, err := req.GetBody()
	if err != nil {
		return "", err
	}

	buf := bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return "", err
	}

	type graphqlBody struct {
		Query string `json:"query"`
	}

	var body graphqlBody
	_ = json.Unmarshal(buf.Bytes(), &body)

	return body.Query, nil
}

func (w *WorkspaceUpdateTestSuite) mockUpdateWorkspaceMutation(errors []string) httpmock.Stub {
	type response struct {
		Data struct {
			WorkspaceUpdate struct {
				Errors []string
			} `json:"workspaceUpdate"`
		} `json:"data"`
	}

	resp := &response{}
	resp.Data.WorkspaceUpdate.Errors = errors

	raw, _ := json.Marshal(resp)

	return httpmock.Stub{
		Matcher:   w.graphqlMutationMatcher(),
		Responder: httpmock.NewStringResponse(http.StatusOK, string(raw)),
	}
}

func (w *WorkspaceUpdateTestSuite) mockGetWorkspaceQuery(workspace api.Workspace) httpmock.Stub {
	type response struct {
		Data struct {
			Group struct {
				Workspaces struct {
					Nodes []api.Workspace `json:"nodes"`
				} `json:"workspaces"`
			} `json:"group"`
		} `json:"data"`
	}

	resp := &response{}
	resp.Data.Group.Workspaces.Nodes = []api.Workspace{workspace}

	raw, _ := json.Marshal(resp)

	return httpmock.Stub{
		Matcher:   w.graphqlQueryMatcher(),
		Responder: httpmock.NewStringResponse(http.StatusOK, string(raw)),
	}
}
