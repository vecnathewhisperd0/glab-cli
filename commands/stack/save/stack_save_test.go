package create

import (
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/commands/cmdtest"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/prompt"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(rt http.RoundTripper, isTTY bool, args string) (*test.CmdOut, error) {
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
			expected: "Saved with message: \"this is a commit message\".\n",
		},

		{
			desc:     "adding files with a dot argument",
			args:     []string{"."},
			files:    []string{"testfile", "randomfile"},
			message:  "this is a commit message",
			expected: "Saved with message: \"this is a commit message\".\n",
		},

		{
			desc:     "omitting a message",
			args:     []string{"."},
			files:    []string{"testfile"},
			expected: "Saved with message: \"oh ok fine how about blah blah\".\n",
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
						Name:  "title",
						Value: "oh ok fine how about blah blah",
					},
				})
			} else {
				tc.args = append(tc.args, "-m")
				tc.args = append(tc.args, "\""+tc.message+"\"")
			}

			dir := git.InitGitRepoWithCommit(t)
			err := git.SetLocalConfig("glab.currentstack", "cool test feature")
			require.Nil(t, err)

			createTemporaryFiles(t, dir, tc.files)

			args := strings.Join(tc.args, " ")

			output, err := runCommand(nil, true, args)

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
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			dir := git.InitGitRepoWithCommit(t)
			err := git.SetLocalConfig("glab.currentstack", "cool test feature")
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
			err := git.SetLocalConfig("glab.currentstack", "cool test feature")
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

func createTemporaryFiles(t *testing.T, dir string, files []string) {
	for _, file := range files {
		file = path.Join(dir, file)
		_, err := os.Create(file)

		require.Nil(t, err)
	}
}
