package ask

import (
	"net/http"
	"strings"
	"testing"

	"gitlab.com/gitlab-org/cli/pkg/prompt"

	"gitlab.com/gitlab-org/cli/commands/cmdtest"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/cli/pkg/httpmock"
	"gitlab.com/gitlab-org/cli/test"
)

func runCommand(rt http.RoundTripper, isTTY bool, args string) (*test.CmdOut, error) {
	ios, _, stdout, stderr := cmdtest.InitIOStreams(isTTY, "")

	factory := cmdtest.InitFactory(ios, rt)

	_, _ = factory.HttpClient()

	cmd := NewCmdAsk(factory)

	return cmdtest.ExecuteCommand(cmd, args, stdout, stderr)
}

func TestAskCmd(t *testing.T) {
	t.Run("git commands", func(t *testing.T) {
		runGitCommandTests(t)
	})

	t.Run("shell commands", func(t *testing.T) {
		runShellCommandTests(t)
	})
}

func runShellCommandTests(t *testing.T) {
	t.Run("basic shell command", func(t *testing.T) {
		fakeHTTP := &httpmock.Mocker{
			MatchURL: httpmock.PathAndQuerystring,
		}
		defer fakeHTTP.Verify(t)

		body := `{"predictions": [{ "candidates": [ {"content": "ls -la"} ]}]}`
		response := httpmock.NewStringResponse(http.StatusOK, body)
		fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

		expectedOutput := "ls -la"
		output, err := runCommand(fakeHTTP, false, "--shell --git=false list files")
		require.NoError(t, err)
		require.Equal(t, expectedOutput, output.String())
	})

	t.Run("complex shell command", func(t *testing.T) {
		fakeHTTP := &httpmock.Mocker{
			MatchURL: httpmock.PathAndQuerystring,
		}
		defer fakeHTTP.Verify(t)

		body := `{"predictions": [{ "candidates": [ {"content": "find . -type f -name '*.txt' -mtime -7 | xargs grep 'pattern'"} ]}]}`
		response := httpmock.NewStringResponse(http.StatusOK, body)
		fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

		expectedOutput := "find . -type f -name '*.txt' -mtime -7 | xargs grep 'pattern'"
		output, err := runCommand(fakeHTTP, false, "--shell --git=false find text files modified in last week containing pattern")
		require.NoError(t, err)
		require.Equal(t, expectedOutput, output.String())
	})

	t.Run("shell command with special characters", func(t *testing.T) {
		fakeHTTP := &httpmock.Mocker{
			MatchURL: httpmock.PathAndQuerystring,
		}
		defer fakeHTTP.Verify(t)

		body := `{"predictions": [{ "candidates": [ {"content": "echo \"Hello, World!\" > output.txt && sed -i 's/World/Everyone/g' output.txt"} ]}]}`
		response := httpmock.NewStringResponse(http.StatusOK, body)
		fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

		expectedOutput := "echo \"Hello, World!\" > output.txt && sed -i 's/World/Everyone/g' output.txt"
		output, err := runCommand(fakeHTTP, false, "--shell --git=false create file saying Hello World and replace World with Everyone")
		require.NoError(t, err)
		require.Equal(t, expectedOutput, output.String())
	})

	t.Run("empty API response for shell command", func(t *testing.T) {
		fakeHTTP := &httpmock.Mocker{
			MatchURL: httpmock.PathAndQuerystring,
		}
		defer fakeHTTP.Verify(t)

		body := `{"predictions": []}`
		response := httpmock.NewStringResponse(http.StatusOK, body)
		fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

		_, err := runCommand(fakeHTTP, false, "--shell --git=false list files")
		require.Error(t, err)
		require.Contains(t, err.Error(), aiResponseErr)
	})

	t.Run("malformed shell command response", func(t *testing.T) {
		fakeHTTP := &httpmock.Mocker{
			MatchURL: httpmock.PathAndQuerystring,
		}
		defer fakeHTTP.Verify(t)

		body := `{"predictions": [{ "candidates": [{}]}]}`
		response := httpmock.NewStringResponse(http.StatusOK, body)
		fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

		_, err := runCommand(fakeHTTP, false, "--shell --git=false list files")
		require.Error(t, err)
		require.Contains(t, err.Error(), aiResponseErr)
	})

	t.Run("missing command in shell response", func(t *testing.T) {
		fakeHTTP := &httpmock.Mocker{
			MatchURL: httpmock.PathAndQuerystring,
		}
		defer fakeHTTP.Verify(t)

		body := `{"predictions": [{ "candidates": [{"content": ""}]}]}`
		response := httpmock.NewStringResponse(http.StatusOK, body)
		fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

		_, err := runCommand(fakeHTTP, false, "--shell --git=false list files")
		require.Error(t, err)
		require.Contains(t, err.Error(), aiResponseErr)
	})
}

func runGitCommandTests(t *testing.T) {
	initialAiResponse := "The appropriate ```git log --pretty=format:'%h'``` Git command ```non-git cmd``` for listing ```git show``` commit SHAs."
	outputWithoutExecution := "Commands:\n" + `
git log --pretty=format:'%h'
non-git cmd
git show

Explanation:

The appropriate git log --pretty=format:'%h' Git command non-git cmd for listing git show commit SHAs.

`

	tests := []struct {
		desc           string
		content        string
		withPrompt     bool
		withExecution  bool
		expectedResult string
	}{
		{
			desc:           "agree to run commands",
			content:        initialAiResponse,
			withPrompt:     true,
			withExecution:  true,
			expectedResult: outputWithoutExecution + "git log executed\ngit show executed\n",
		},
		{
			desc:           "disagree to run commands",
			content:        initialAiResponse,
			withPrompt:     true,
			withExecution:  false,
			expectedResult: outputWithoutExecution,
		},
		{
			desc:           "no commands",
			content:        "There are no Git commands related to the text.",
			withPrompt:     false,
			expectedResult: "Commands:\n\n\nExplanation:\n\nThere are no Git commands related to the text.\n\n",
		},
	}
	cmdLogResult := "git log executed"
	cmdShowResult := "git show executed"

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fakeHTTP := &httpmock.Mocker{
				MatchURL: httpmock.PathAndQuerystring,
			}
			defer fakeHTTP.Verify(t)

			body := `{"predictions": [{ "candidates": [ {"content": "` + tc.content + `"} ]}]}`

			response := httpmock.NewStringResponse(http.StatusOK, body)
			fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

			if tc.withPrompt {
				restore := prompt.StubConfirm(tc.withExecution)
				defer restore()

				cs, restore := test.InitCmdStubber()
				defer restore()
				cs.Stub(cmdLogResult)
				cs.Stub(cmdShowResult)
			}

			output, err := runCommand(fakeHTTP, false, "git list 10 commits")
			require.Nil(t, err)

			require.Equal(t, output.String(), tc.expectedResult)
			require.Empty(t, output.Stderr())
		})
	}
}

func TestFlagCombinations(t *testing.T) {
	tests := []struct {
		desc           string
		args           string
		expectedOutput string
		expectedErr    string
	}{
		{
			desc:        "both git and shell flags",
			args:        "--git --shell list files",
			expectedErr: "cannot use both --git and --shell flags",
		},
		{
			desc:           "no flags provided",
			args:          "list files",
			expectedOutput: "Commands:", // Just checking start of output since default is git mode
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fakeHTTP := &httpmock.Mocker{
				MatchURL: httpmock.PathAndQuerystring,
			}
			defer fakeHTTP.Verify(t)

			if tc.expectedErr == "" {
				body := `{"predictions": [{ "candidates": [ {"content": "git status"} ]}]}`
				response := httpmock.NewStringResponse(http.StatusOK, body)
				fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)
			}

			output, err := runCommand(fakeHTTP, false, tc.args)
			
			if tc.expectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Contains(t, output.String(), tc.expectedOutput)
			}
		})
	}
}

func TestInputValidation(t *testing.T) {
	tests := []struct {
		desc        string
		input       string
		expectedErr string
	}{
		{
			desc:        "empty prompt",
			input:       "",
			expectedErr: "prompt required",
		},
		{
			desc:        "very long prompt",
			input:       strings.Repeat("a", 10000),
			expectedErr: "prompt too long",
		},
		{
			desc:        "prompt with special characters",
			input:       "--shell \"hello; rm -rf /\"",
			expectedErr: "invalid characters in prompt",
		},
		{
			desc:        "dangerous rm command",
			input:       "--shell \"remove all files from root\"",
			expectedErr: "dangerous command pattern detected: rm -rf /",
		},
		{
			desc:        "dangerous sudo command",
			input:       "--shell \"sudo apt-get update\"",
			expectedErr: "dangerous command pattern detected: sudo",
		},
		{
			desc:        "dangerous download command",
			input:       "--shell \"wget https://example.com/script.sh\"",
			expectedErr: "dangerous command pattern detected: wget",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := runCommand(nil, false, tc.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestFailedHttpResponse(t *testing.T) {
	tests := []struct {
		desc        string
		code        int
		response    string
		expectedMsg string
	}{
		{
			desc:        "API error",
			code:        http.StatusNotFound,
			response:    `{"message": "Error message"}`,
			expectedMsg: "404 Not Found",
		},
		{
			desc:        "Empty response",
			code:        http.StatusOK,
			response:    `{"choices": []}`,
			expectedMsg: aiResponseErr,
		},
		{
			desc:        "Bad JSON",
			code:        http.StatusOK,
			response:    `{"choices": [{"message": {"content": "hello"}}]}`,
			expectedMsg: aiResponseErr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			fakeHTTP := &httpmock.Mocker{
				MatchURL: httpmock.PathAndQuerystring,
			}
			defer fakeHTTP.Verify(t)

			response := httpmock.NewStringResponse(tc.code, tc.response)
			fakeHTTP.RegisterResponder(http.MethodPost, "/api/v4/ai/llm/git_command", response)

			_, err := runCommand(fakeHTTP, false, "git list 10 commits")
			require.NotNil(t, err)
			require.Contains(t, err.Error(), tc.expectedMsg)
		})
	}
}
