package save

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/prompt"
	"gitlab.com/gitlab-org/cli/test"
)

func runSaveCommand(rt http.RoundTripper, isTTY bool, args string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := cmdtest.InitIOStreams(isTTY, "")

	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := NewCmdSaveStack(factory)

	return cmdtest.ExecuteCommand(cmd, args, stdout, stderr)
}

func TestSaveNewStack(t *testing.T) {
	tests := []struct {
		desc     string
		args     []string
		files    []string
		message  string
		expected string
		wantErr  bool
	}{
		{
			desc:     "adding regular files",
			args:     []string{"testfile", "randomfile"},
			files:    []string{"testfile", "randomfile"},
			message:  "this is a commit message",
			expected: "• cool-test-feature: Saved with message: \"this is a commit message\".\n",
		},

		{
			desc:     "adding files with a dot argument",
			args:     []string{"."},
			files:    []string{"testfile", "randomfile"},
			message:  "this is a commit message",
			expected: "• cool-test-feature: Saved with message: \"this is a commit message\".\n",
		},

		{
			desc:     "omitting a message",
			args:     []string{"."},
			files:    []string{"testfile"},
			expected: "• cool-test-feature: Saved with message: \"oh ok fine how about blah blah\".\n",
		},

		{
			desc:     "with no changed files",
			args:     []string{"."},
			files:    []string{},
			expected: "could not save: \"no changes to save\"",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.message == "" {
				as, restoreAsk := prompt.InitAskStubber()
				defer restoreAsk()

				as.Stub([]*prompt.QuestionStub{
					{
						Name:  "description",
						Value: "oh ok fine how about blah blah",
					},
				})
			} else {
				tc.args = append(tc.args, "-m")
				tc.args = append(tc.args, "\""+tc.message+"\"")
			}

			dir := git.InitGitRepoWithCommit(t)
			err := git.SetLocalConfig("glab.currentstack", "cool-test-feature")
			require.Nil(t, err)

			createTemporaryFiles(t, dir, tc.files)

			args := strings.Join(tc.args, " ")

			output, err := runSaveCommand(nil, true, args)

			if tc.wantErr {
				require.Errorf(t, err, tc.expected)
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.expected, output.String())
			}
		})
	}
}

func Test_addFiles(t *testing.T) {
	tests := []struct {
		desc     string
		args     []string
		expected []string
	}{
		{
			desc:     "adding regular files",
			args:     []string{"file1", "file2"},
			expected: []string{"file1", "file2"},
		},
		{
			desc:     "adding files with a dot argument",
			args:     []string{"."},
			expected: []string{"file1", "file2"},
		},
		{
			desc:     "adding files with no argument",
			expected: []string{"file1", "file2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			dir := git.InitGitRepoWithCommit(t)
			err := git.SetLocalConfig("glab.currentstack", "cool-test-feature")
			require.Nil(t, err)

			createTemporaryFiles(t, dir, tc.expected)

			_, err = addFiles(tc.args)
			require.Nil(t, err)

			gitCmd := git.GitCommand("status", "--short", "-u")
			output, err := run.PrepareCmd(gitCmd).Output()
			require.Nil(t, err)

			normalizedFiles := []string{}
			for _, file := range tc.expected {
				file = "A  " + file

				normalizedFiles = append(normalizedFiles, file)
			}

			formattedOutput := strings.Replace(string(output), "\n", "", -1)
			require.Equal(t, formattedOutput, strings.Join(normalizedFiles, ""))
		})
	}
}

func Test_checkForChanges(t *testing.T) {
	tests := []struct {
		desc     string
		args     []string
		expected bool
	}{
		{
			desc:     "check for changes with modified files",
			args:     []string{"file1", "file2"},
			expected: true,
		},
		{
			desc:     "check for changes without anything",
			args:     []string{},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			dir := git.InitGitRepoWithCommit(t)
			err := git.SetLocalConfig("glab.currentstack", "cool-test-feature")
			require.Nil(t, err)

			createTemporaryFiles(t, dir, tc.args)

			err = checkForChanges()
			if tc.expected {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func Test_commitFiles(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		message string
		wantErr bool
	}{
		{
			name:    "a regular commit message",
			message: "i am a test message",
			want:    "i am a test message\n 2 files changed, 0 insertions(+), 0 deletions(-)\n create mode 100644 test\n create mode 100644 yo\n",
		},
		{
			name:    "no message",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := git.InitGitRepoWithCommit(t)

			createTemporaryFiles(t, dir, []string{"yo", "test"})
			_, err := addFiles([]string{"."})
			require.Nil(t, err)

			got, err := commitFiles(tt.message)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.Contains(t, got, tt.want)
			}
		})
	}
}

func Test_generateStackSha(t *testing.T) {
	type args struct {
		message string
		title   string
		author  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic test",
			args: args{message: "hello", title: "supercool stack title", author: "norm maclean"},
			want: "f541d1d62a519a43e6242bcf3f2f6e7f4310c01e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			git.InitGitRepo(t)

			got, err := generateStackSha(tt.args.message, tt.args.title, tt.args.author)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)

				require.Equal(t, got, tt.want)
			}
		})
	}
}

func Test_createShaBranch(t *testing.T) {
	type args struct {
		sha   string
		title string
	}
	tests := []struct {
		name     string
		args     args
		prefix   string
		want     string
		wantErr  bool
		noConfig bool
	}{
		{
			name:   "standard test case",
			args:   args{sha: "237ec83c03d3", title: "cool-change"},
			prefix: "asdf",
			want:   "asdf-cool-change-237ec83c",
		},
		{
			name:     "with no config file",
			args:     args{sha: "237ec83c03d3", title: "cool-change"},
			prefix:   "",
			want:     "jawn-cool-change-237ec83c",
			noConfig: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			git.InitGitRepo(t)

			defer config.StubWriteConfig(io.Discard, io.Discard)()

			factory := createFactoryWithConfig("branch_prefix", tt.prefix)

			if tt.noConfig {
				t.Setenv("USER", "jawn")
			}

			got, err := createShaBranch(factory, tt.args.sha, tt.args.title)
			require.Nil(t, err)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_gatherStackRefs(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name   string
		args   args
		stacks []StackRef
	}{
		{
			name: "with multiple files",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "789", Prev: "456", Next: ""},
			},
		},
		{
			name: "with 1 file",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := git.InitGitRepo(t)

			for _, stack := range tt.stacks {
				err := addStackRefFile(tt.args.title, stack)
				require.Nil(t, err)
			}

			for _, stack := range tt.stacks {
				file := path.Join(dir, StackLocation, tt.args.title, stack.SHA+".json")

				data, err := os.ReadFile(file)
				require.Nil(t, err)

				temp := StackRef{}
				err = json.Unmarshal(data, &temp)
				require.Nil(t, err)

				require.Equal(t, stack, temp)
			}
		})
	}
}

func Test_lastRefInChain(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name        string
		args        args
		stacks      []StackRef
		expected    StackRef
		expectedErr bool
	}{
		{
			name: "with multiple files",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "789", Prev: "456", Next: ""},
			},
			expected: StackRef{SHA: "789", Prev: "456", Next: ""},
		},
		{
			name: "with multiple bad data that might infinite loop",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "789", Prev: "456", Next: "123"},
			},
			expectedErr: true,
		},
		{
			name: "with 1 file",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: ""},
			},
			expected: StackRef{SHA: "123", Prev: "", Next: ""},
		},
		{
			name: "large number",
			args: args{title: "title-123"},
			stacks: []StackRef{
				{SHA: "13", Prev: "12", Next: ""},
				{SHA: "3", Prev: "2", Next: "4"},
				{SHA: "10", Prev: "9", Next: "11"},
				{SHA: "2", Prev: "1", Next: "3"},
				{SHA: "5", Prev: "4", Next: "6"},
				{SHA: "6", Prev: "5", Next: "7"},
				{SHA: "7", Prev: "6", Next: "8"},
				{SHA: "4", Prev: "3", Next: "5"},
				{SHA: "12", Prev: "11", Next: "13"},
				{SHA: "9", Prev: "8", Next: "10"},
				{SHA: "1", Prev: "", Next: "2"},
				{SHA: "11", Prev: "10", Next: "12"},
				{SHA: "8", Prev: "7", Next: "9"},
			},
			expected: StackRef{SHA: "13", Prev: "12", Next: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lastRefInChain(tt.stacks)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

			require.Equal(t, tt.expected, got)
		})
	}
}

func Test_sortRefs(t *testing.T) {
	type args struct {
		title string
	}
	tests := []struct {
		name        string
		args        args
		stacks      []StackRef
		expected    []StackRef
		expectedErr bool
	}{
		{
			name: "with multiple files",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "789", Prev: "456", Next: ""},
			},
			expected: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "789", Prev: "456", Next: ""},
			},
		},
		{
			name: "with multiple bad data that might infinite loop",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: "456"},
				{SHA: "456", Prev: "123", Next: "789"},
				{SHA: "789", Prev: "456", Next: "123"},
			},
			expectedErr: true,
		},
		{
			name: "with 1 file",
			args: args{title: "sweet-title-123"},
			stacks: []StackRef{
				{SHA: "123", Prev: "", Next: ""},
			},
			expected: []StackRef{
				{SHA: "123", Prev: "", Next: ""},
			},
		},
		{
			name: "large number",
			args: args{title: "title-123"},
			stacks: []StackRef{
				{SHA: "13", Prev: "12", Next: ""},
				{SHA: "3", Prev: "2", Next: "4"},
				{SHA: "10", Prev: "9", Next: "11"},
				{SHA: "2", Prev: "1", Next: "3"},
				{SHA: "5", Prev: "4", Next: "6"},
				{SHA: "6", Prev: "5", Next: "7"},
				{SHA: "7", Prev: "6", Next: "8"},
				{SHA: "4", Prev: "3", Next: "5"},
				{SHA: "12", Prev: "11", Next: "13"},
				{SHA: "9", Prev: "8", Next: "10"},
				{SHA: "1", Prev: "", Next: "2"},
				{SHA: "11", Prev: "10", Next: "12"},
				{SHA: "8", Prev: "7", Next: "9"},
			},
			expected: []StackRef{
				{SHA: "1", Prev: "", Next: "2"},
				{SHA: "2", Prev: "1", Next: "3"},
				{SHA: "3", Prev: "2", Next: "4"},
				{SHA: "4", Prev: "3", Next: "5"},
				{SHA: "5", Prev: "4", Next: "6"},
				{SHA: "6", Prev: "5", Next: "7"},
				{SHA: "7", Prev: "6", Next: "8"},
				{SHA: "8", Prev: "7", Next: "9"},
				{SHA: "9", Prev: "8", Next: "10"},
				{SHA: "10", Prev: "9", Next: "11"},
				{SHA: "11", Prev: "10", Next: "12"},
				{SHA: "12", Prev: "11", Next: "13"},
				{SHA: "13", Prev: "12", Next: ""},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			git.InitGitRepo(t)

			_, err := sortRefs(tt.stacks)
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

			for k, stack := range tt.expected {
				require.Equal(t, stack, tt.expected[k])
			}
		})
	}
}

func Test_addStackRefFile(t *testing.T) {
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
			dir := git.InitGitRepo(t)

			err := addStackRefFile(tt.args.title, tt.args.stackRef)
			require.Nil(t, err)

			file := path.Join(dir, StackLocation, tt.args.title, tt.args.stackRef.SHA+".json")
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

func Test_updateStackRefFile(t *testing.T) {
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
			dir := git.InitGitRepo(t)

			// add the initial data
			initial := StackRef{Prev: "123", Branch: "gmh"}
			err := addStackRefFile(tt.args.title, initial)
			require.Nil(t, err)

			err = updateStackRefFile(tt.args.title, tt.args.stackRef)
			require.Nil(t, err)

			file := path.Join(dir, StackLocation, tt.args.title, tt.args.stackRef.SHA+".json")
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

func createTemporaryFiles(t *testing.T, dir string, files []string) {
	for _, file := range files {
		file = path.Join(dir, file)
		_, err := os.Create(file)

		require.Nil(t, err)
	}
}

func createFactoryWithConfig(key string, value string) *cmdutils.Factory {
	strconfig := heredoc.Doc(`
				` + key + `: ` + value + `
			`)

	cfg := config.NewFromString(strconfig)

	ios, _, _, _ := iostreams.Test()

	return &cmdutils.Factory{
		IO: ios,
		Config: func() (config.Config, error) {
			return cfg, nil
		},
	}
}
