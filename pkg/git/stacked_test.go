package git

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddStackRefDir(t *testing.T) {
	tests := []struct {
		name   string
		branch string
		want   string
	}{
		{
			name:   "normal filename",
			branch: "thing",
			want:   "thing",
		},
		{
			name:   "advanced filename",
			branch: "something-with-dashes",
			want:   "something-with-dashes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := initGitRepo(t)

			_, err := AddStackRefDir(tt.branch)
			require.NoError(t, err)

			refDir := path.Join(baseDir, "/.git/refs/stacked/")

			_, err = os.Stat(path.Join(refDir, tt.branch))
			require.NoError(t, err)
		})
	}
}

func TestSetLocalConfig(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		existingConfig bool
	}{
		{
			name:           "config already exists",
			value:          "exciting new value",
			existingConfig: true,
		},
		{
			name:           "config doesn't exist",
			value:          "default value",
			existingConfig: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := initGitRepo(t)
			defer os.RemoveAll(tempDir)

			if tt.existingConfig {
				_ = GitCommand("config", "--local", "this.glabstacks", "prev-value")
			}

			err := SetLocalConfig("this.glabstacks", tt.value)
			require.NoError(t, err)

			config, err := GetAllConfig("this.glabstacks")
			require.NoError(t, err)

			// GetAllConfig() appends a new line. Let's get rid of that.
			compareString := strings.TrimSuffix(string(config), "\n")

			if compareString != tt.value {
				t.Errorf("config value = %v, want %v", compareString, tt.value)
			}
		})
	}
}
