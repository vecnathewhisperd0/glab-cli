package remove

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xanzy/go-gitlab"
	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

var tests = []struct {
	expectedUse   string
	expectedShort string
}{
	{
		expectedUse:   "remove [username | ID]",
		expectedShort: "Remove a user from a project",
	},
}

func TestNewCmdRemove(t *testing.T) {
	defer config.StubConfig(`---
hosts:
  gitlab.com:
    username: monalisa
    token: OTOKEN
no_prompt: true
`, "")()

	api.CurrentUser = func(client *gitlab.Client) (*gitlab.User, error) {
		return &gitlab.User{
			ID:       1,
			Username: "username",
			Name:     "name",
		}, nil
	}

	api.RemoveProjectMember = func(client *gitlab.Client, projectID interface{}, user int) (*gitlab.Response, error) {
		return &gitlab.Response{}, nil
	}

	io, _, stdout, stderr := iostreams.Test()
	stubFactory := cmdtest.StubFactory("")
	stubFactory.IO = io
	stubFactory.IO.IsaTTY = false
	stubFactory.IO.IsErrTTY = false

	cmd := NewCmdRemove(stubFactory)

	assert.Equal(t, tests[0].expectedUse, cmd.Use)
	assert.Equal(t, tests[0].expectedShort, cmd.Short)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe"})
	err := cmd.Execute()
	assert.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "Removed user john.doe from")
	assert.Equal(t, "", stderr.String())
}

func Test_userIdFromArgs(t *testing.T) {
	client := &gitlab.Client{}
	tests := []struct {
		name    string
		args    []string
		want    int
		wantErr bool
	}{
		{
			name:    "Valid user ID",
			args:    []string{"123"},
			want:    123,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userIdFromArgs(client, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("userIdFromArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("userIdFromArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
