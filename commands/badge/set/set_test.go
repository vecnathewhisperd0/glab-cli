// Create tests for the functions in set.go. Be sure to use the httpmock library for any external calls though. These are the libraries that should be imported:
// 	"net/http"
// "testing"
// "github.com/stretchr/testify/assert"
// "github.com/stretchr/testify/require"
// "gitlab.com/gitlab-org/cli/commands/cmdtest"
// "gitlab.com/gitlab-org/cli/pkg/httpmock"
// "gitlab.com/gitlab-org/cli/test"
package set

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func TestNewCmdSet(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wants    SetOptions
		wantsErr string
	}{
		{
			name: "valid input",
			cli:  "project 123 passing",
			wants: SetOptions{
				EntityID:   "123",
				EntityType: "project",
				State:      "passing",
			},
		},
		{
			name:     "missing arguments",
			cli:      "",
			wantsErr: "accepts 3 arg(s), received 0",
		},
		{
			name:     "invalid entity type",
			cli:      "invalid 123 passing",
			wantsErr: "invalid entity type: invalid",
		},
		{
			name:     "invalid state",
			cli:      "project 123 invalid",
			wantsErr: "invalid state: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io, _, _, _ := cmdtest.InitIOStreams(tt.cli, nil)
			cmd := NewCmdSet(io)
			cmd.SetArgs(cmdtest.SplitArgs(tt.cli))
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Out)
			cmd.SetErr(io.ErrOut)

			_, err := cmd.ExecuteC()
			if tt.wantsErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantsErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wants.EntityID, cmd.Flags().Lookup("id").Value.String())
			assert.Equal(t, tt.wants.EntityType, cmd.Flags().Lookup("type").Value.String())
			assert.Equal(t, tt.wants.State, cmd.Flags().Lookup("state").Value.String())
		})
	}
}

func TestSetRun(t *testing.T) {
	reg := httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.REST("POST", "projects/123/badges"),
		httpmock.StringResponse(`{"id": 1, "link_url": "https://example.com", "image_url": "https://example.com/badge.svg"}`),
	)

	io, _, stdout, stderr := cmdtest.InitIOStreams("", nil)

	opts := &SetOptions{
		EntityID:   "123",
		EntityType: "project",
		State:      "passing",
		IO:         io,
		HTTPClient: func() (*http.Client, error) {
			return &http.Client{Transport: &reg}, nil
		},
	}

	err := setRun(opts)
	require.NoError(t, err)

	assert.Equal(t, "", stderr.String())
	assert.Equal(t, "✓ Badge set successfully\n", stdout.String())
}

func TestSetRun_Error(t *testing.T) {
	reg := httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.REST("POST", "projects/123/badges"),
		httpmock.StatusStringResponse(400, `{"message": "Bad Request"}`),
	)

	io, _, _, _ := cmdtest.InitIOStreams("", nil)

	opts := &SetOptions{
		EntityID:   "123",
		EntityType: "project",
		State:      "passing",
		IO:         io,
		HTTPClient: func() (*http.Client, error) {
			return &http.Client{Transport: &reg}, nil
		},
	}

	err := setRun(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 400: Bad Request")
}

