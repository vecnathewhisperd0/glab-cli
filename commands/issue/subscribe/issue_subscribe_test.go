package subscribe

import (
	"fmt"
	"testing"
	"time"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
)

func TestNewCmdSubscribe(t *testing.T) {
	t.Parallel()

	oldSubscribeIssue := api.SubscribeToIssue
	timer, _ := time.Parse(time.RFC3339, "2014-11-12T11:45:26.371Z")
	api.SubscribeToIssue = func(client *gitlab.Client, projectID interface{}, issueID int, opts gitlab.RequestOptionFunc) (*gitlab.Issue, error) {
		if projectID == "" || projectID == "WRONG_REPO" || projectID == "expected_err" || issueID == 0 {
			return nil, fmt.Errorf("error expected")
		}
		return &gitlab.Issue{
			ID:          issueID,
			IID:         issueID,
			State:       "closed",
			Description: "Dummy description for issue " + string(rune(issueID)),
			Author: &gitlab.IssueAuthor{
				ID:       1,
				Name:     "John Dev Wick",
				Username: "jdwick",
			},
			CreatedAt: &timer,
		}, nil
	}

	testCases := []struct {
		Name    string
		Issue   string
		stderr  string
		wantErr bool
	}{
		{
			Name:   "Issue Exists",
			Issue:  "1",
			stderr: "- Subscribing to Issue #1 in glab-cli/test\n✓ Subscribed\n",
		},
		{
			Name:    "Issue Does Not Exist",
			Issue:   "0",
			stderr:  "- Subscribing to Issue #0 in glab-cli/test\nerror expected\n",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			io, _, _, stderr := iostreams.Test()
			f := cmdtest.StubFactory("https://gitlab.com/glab-cli/test")
			f.IO = io
			f.IO.IsaTTY = true
			f.IO.IsErrTTY = true

			cmd := NewCmdSubscribe(f)
			cmd.Flags().StringP("repo", "R", "", "")

			_, err := cmdtest.RunCommand(cmd, tc.Issue)
			if tc.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.stderr, stderr.String())
		})
	}

	api.SubscribeToIssue = oldSubscribeIssue
}
