package create

import (
	"net/http"
	"os"
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

	cmd := NewCmdCreateStack(factory)

	return cmdtest.ExecuteCommand(cmd, args, stdout, stderr)
}

func TestCreateNewStack(t *testing.T) {
	tests := []struct {
		desc           string
		branch         string
		expectedBranch string
		warning        bool
	}{
		{
			desc:           "basic method",
			branch:         "test description here",
			expectedBranch: "test-description-here",
			warning:        false,
		},
		{
			desc:           "empty string",
			branch:         "",
			expectedBranch: "oh-ok-fine-how-about-blah-blah",
			warning:        true,
		},
		{
			desc:           "weird characters git won't like",
			branch:         "hey@#$!^$#)()*1234hmm",
			expectedBranch: "hey-1234hmm",
			warning:        true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.branch == "" {
				as, restoreAsk := prompt.InitAskStubber()
				defer restoreAsk()

				as.Stub([]*prompt.QuestionStub{
					{
						Name:  "title",
						Value: "oh ok fine how about blah blah",
					},
				})
			}

			tempDir, err := os.MkdirTemp("", "empty-git-directory")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			defer os.RemoveAll(tempDir)

			err = os.Chdir(tempDir)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			gitInit := git.GitCommand("init")
			_, err = run.PrepareCmd(gitInit).Output()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			output, err := runCommand(nil, true, tc.branch)
			require.Nil(t, err)

			require.Equal(t, "New stack created with branch \""+tc.expectedBranch+"\".\n", output.String())

			if tc.warning == true {
				require.Equal(t, "\nwarning: non-usable characters have been replaced with dashes\n", output.Stderr())
			} else {
				require.Empty(t, output.Stderr())
			}

			branchOutput, err := git.CurrentBranch()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			require.Equal(t, tc.expectedBranch, branchOutput)
		})
	}
}
