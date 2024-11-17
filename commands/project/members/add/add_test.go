package add

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
	alvString     string
	alv           gitlab.AccessLevelValue
	expectedUse   string
	expectedShort string
}{
	{
		alvString:     "developer",
		alv:           gitlab.DeveloperPermissions,
		expectedUse:   "add [username | ID] [flags]",
		expectedShort: "Add a user to a project",
	},
}

func TestNewCmdAdd(t *testing.T) {
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

	api.AddProjectMember = func(client *gitlab.Client, projectID interface{}, opt *gitlab.AddProjectMemberOptions) (*gitlab.ProjectMember, error) {
		return &gitlab.ProjectMember{
			ID:          1,
			Username:    "john.doe",
			AccessLevel: gitlab.DeveloperPermissions,
		}, nil
	}

	io, _, stdout, stderr := iostreams.Test()
	stubFactory := cmdtest.StubFactory("")
	stubFactory.IO = io
	stubFactory.IO.IsaTTY = false
	stubFactory.IO.IsErrTTY = false

	cmd := NewCmdAdd(stubFactory)

	assert.Equal(t, tests[0].expectedUse, cmd.Use)
	assert.Equal(t, tests[0].expectedShort, cmd.Short)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"john.doe", "-a", "developer"})
	err := cmd.Execute()
	assert.NoError(t, err)

	out := stdout.String()
	assert.Contains(t, out, "User john.doe has been added to")
	assert.Equal(t, "", stderr.String())
}

func Test_getAccessLevelValue(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		want    gitlab.AccessLevelValue
		wantErr bool
	}{
		{
			name:    "Valid access level",
			level:   "developer",
			want:    gitlab.DeveloperPermissions,
			wantErr: false,
		},
		{
			name:    "Invalid access level",
			level:   "invalid",
			want:    gitlab.NoPermissions,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getAccessLevelValue(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAccessLevelValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAccessLevelValue() got = %v, want %v", got, tt.want)
			}
		})
	}
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
