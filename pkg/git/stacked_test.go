package git

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/internal/config"
)

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
			tempDir := InitGitRepo(t)
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
			baseDir := InitGitRepo(t)

			_, err := AddStackRefDir(tt.branch)
			require.NoError(t, err)

			refDir := filepath.Join(baseDir, "/.git/refs/stacked/")

			_, err = os.Stat(filepath.Join(refDir, tt.branch))
			require.NoError(t, err)
		})
	}
}

func Test_StackRootDir(t *testing.T) {
	// TODO: write test
}

func Test_AddStackRefFile(t *testing.T) {
	type args struct {
		title    string
		stackRef StackRef
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no message",
			args: args{
				title: "sweet-title-123",
				stackRef: StackRef{
					Prev:   "hello",
					Branch: "gmh-feature-3ab3da",
					Next:   "goodbye",
					SHA:    "1a2b3c4d",
					MR:     "https://gitlab.com/",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := InitGitRepo(t)

			err := AddStackRefFile(tt.args.title, tt.args.stackRef)
			require.Nil(t, err)

			file := filepath.Join(dir, StackLocation, tt.args.title, tt.args.stackRef.SHA+".json")
			require.True(t, config.CheckFileExists(file))

			stackRef := StackRef{}
			readData, err := os.ReadFile(file)
			require.Nil(t, err)

			err = json.Unmarshal(readData, &stackRef)
			require.Nil(t, err)

			require.Equal(t, stackRef, tt.args.stackRef)
		})
	}
}

func Test_DeleteStackRefFile(t *testing.T) {
	// TODO: write test
}

func Test_UpdateStackRefFile(t *testing.T) {
	type args struct {
		title    string
		stackRef StackRef
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "no message",
			args: args{
				title: "sweet-title-123",
				stackRef: StackRef{
					Prev:   "hello",
					Branch: "gmh-feature-3ab3da",
					Next:   "goodbye",
					SHA:    "1a2b3c4d",
					MR:     "https://gitlab.com/",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := InitGitRepo(t)

			// add the initial data
			initial := StackRef{Prev: "123", Branch: "gmh"}
			err := AddStackRefFile(tt.args.title, initial)
			require.Nil(t, err)

			err = UpdateStackRefFile(tt.args.title, tt.args.stackRef)

			require.Nil(t, err)

			file := filepath.Join(dir, StackLocation, tt.args.title, tt.args.stackRef.SHA+".json")
			require.True(t, config.CheckFileExists(file))

			stackRef := StackRef{}
			readData, err := os.ReadFile(file)
			require.Nil(t, err)

			err = json.Unmarshal(readData, &stackRef)
			require.Nil(t, err)

			require.Equal(t, stackRef, tt.args.stackRef)
		})
	}
}

func Test_GetStacks(t *testing.T) {
	t.Run("two stacks", func(t *testing.T) {
		stacks := []Stack{
			{
				Title: "stack-0",
				Refs: map[string]StackRef{
					"0": {
						Description: "stack-0 initial commit",
					},
				},
			},
			{
				Title: "stack-1",
				Refs: map[string]StackRef{
					"0": {
						Description: "stack-1 initial commit",
					},
				},
			},
		}
		InitGitRepo(t)
		var want []Stack
		for _, v := range stacks {
			for _, ref := range v.Refs {
				err := AddStackRefFile(v.Title, ref)
				require.Nil(t, err)
			}
			want = append(want, Stack{Title: v.Title})
		}
		got, err := GetStacks()
		require.Nil(t, err)
		require.Equal(t, want, got)
	})
	t.Run("no stacks", func(t *testing.T) {
		InitGitRepo(t)
		got, err := GetStacks()
		var want []Stack = nil
		require.NotNil(t, err)
		require.Equal(t, want, got)
	})
}

func createRefFiles(refs map[string]StackRef, title string) error {
	for _, ref := range refs {
		err := AddStackRefFile(title, ref)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestCreateStack(t *testing.T) {
	tests := []struct {
		name    string
		stack   Stack
		wantErr bool
	}{
		{
			name: "create valid stack",
			stack: Stack{
				Name: "test-stack",
				Base: "main",
				Head: "feature",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory
			dir, err := os.MkdirTemp("", "TestCreateStack")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			// Change to the temporary directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(dir)
			require.NoError(t, err)
			defer func() {
				err := os.Chdir(oldWd)
				require.NoError(t, err)
			}()

			// Initialize Git repository
			cmd := exec.Command("git", "init")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Git init failed: %s", string(output))

			// Create an initial commit
			cmd = exec.Command("git", "commit", "--allow-empty", "-m", "Initial commit")
			output, err = cmd.CombinedOutput()
			require.NoError(t, err, "Initial commit failed: %s", string(output))

			// Set up Git user configuration
			cmd = exec.Command("git", "config", "user.name", "Test User")
			output, err = cmd.CombinedOutput()
			require.NoError(t, err, "Setting Git user.name failed: %s", string(output))

			cmd = exec.Command("git", "config", "user.email", "test@example.com")
			output, err = cmd.CombinedOutput()
			require.NoError(t, err, "Setting Git user.email failed: %s", string(output))

			// Log Git status before creating stack
			cmd = exec.Command("git", "status")
			output, err = cmd.CombinedOutput()
			require.NoError(t, err, "Git status failed: %s", string(output))

			err = CreateStack(tt.stack.Name, tt.stack.Base, tt.stack.Head)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				if err != nil {
					t.Fatalf("CreateStack failed: %v", err)
				}

				// Check if the stack metadata file was created
				metadataPath := filepath.Join(dir, ".git", "refs", "stacked", tt.stack.Name)
				require.FileExists(t, metadataPath)

				// Read the metadata file and verify its contents
				content, err := os.ReadFile(metadataPath)
				require.NoError(t, err)

				var stack Stack
				err = json.Unmarshal(content, &stack)
				require.NoError(t, err)

				require.Equal(t, tt.stack.Name, stack.Name)
				require.Equal(t, tt.stack.Base, stack.Base)
				require.Equal(t, tt.stack.Head, stack.Head)
				require.NotEmpty(t, stack.MetadataHash)
			}
		})
	}
}
