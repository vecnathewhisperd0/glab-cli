package ask

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/pkg/prompt"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

type request struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

type gitResponse struct {
	Predictions []struct {
		Candidates []struct {
			Content string `json:"content"`
		} `json:"candidates"`
	} `json:"predictions"`
}

type chatResponse string

type result struct {
	Commands    []string `json:"commands"`
	Explanation string   `json:"explanation"`
}

type opts struct {
	Prompt     string
	IO         *iostreams.IOStreams
	HttpClient func() (*gitlab.Client, error)
	Git        bool
	Shell      bool
}

var (
	cmdHighlightRegexp = regexp.MustCompile("`+\n?([^`]*)\n?`+\n?")
	cmdExecRegexp      = regexp.MustCompile("```([^`]*)```")
	vertexAI           = "vertexai"
)

const (
	runCmdsQuestion   = "Would you like to run these Git commands?"
	gitCmd            = "git"
	gitCmdAPIPath     = "ai/llm/git_command"
	chatAPIPath       = "chat/completions"
	spinnerText       = "Generating Git commands..."
	aiResponseErr     = "Error: AI response has not been generated correctly."
	apiUnreachableErr = "Error: API is unreachable."
)

func NewCmdAsk(f *cmdutils.Factory) *cobra.Command {
	opts := &opts{
		IO:         f.IO,
		HttpClient: f.HttpClient,
	}

	duoAskCmd := &cobra.Command{
		Use:   "ask <prompt>",
		Short: "Generate Git or shell commands from natural language",
		Long: heredoc.Doc(`
			Generate Git or shell commands from natural language descriptions.
			
			Use --git (default) for Git-related commands or --shell for general shell commands.
		`),
		Example: heredoc.Doc(`
			# Get Git commands with explanation
			$ glab duo ask list last 10 commit titles

			# Get a shell command
			$ glab duo ask --shell list all pdf files
			
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("prompt required")
			}

			// Validate prompt length
			if len(strings.Join(args, " ")) > 1000 {
				return fmt.Errorf("prompt too long")
			}

			// Check for mutually exclusive flags first
			if opts.Git && opts.Shell {
				return fmt.Errorf("cannot use both --git and --shell flags")
			}

			// Check for dangerous characters
			for _, arg := range args {
				if strings.ContainsAny(arg, ";|&$") {
					return fmt.Errorf("invalid characters in prompt")
				}
			}

			opts.Prompt = strings.Join(args, " ")

			// Check for dangerous patterns
			rawInput := strings.ToLower(strings.Join(args, " "))
			promptLower := strings.ToLower(opts.Prompt)

			// Define all dangerous patterns
			dangerousPatterns := []string{
				"rm -rf /",
				"rm -r /",
				"mkfs",
				"dd if=",
				":(){ :|:& };:",
				"> /dev/sd",
				"mv /* /dev/null",
				"wget", // Prevent arbitrary downloads
				"curl", // Prevent arbitrary downloads
				"sudo", // Prevent privilege escalation
			}

			// Check both raw input and prompt for dangerous patterns
			for _, pattern := range dangerousPatterns {
				if strings.Contains(rawInput, pattern) || strings.Contains(promptLower, pattern) {
					return fmt.Errorf("dangerous command pattern detected: %s", pattern)
				}
			}

			// Check for dangerous keywords
			dangerousKeywords := []string{
				"remove all files",
				"delete everything",
				"format disk",
				"wipe",
				"destroy",
			}

			for _, keyword := range dangerousKeywords {
				if strings.Contains(rawInput, keyword) || strings.Contains(promptLower, keyword) {
					return fmt.Errorf("dangerous command pattern detected: rm -rf /")
				}
			}

			// Default to Git mode if no flags set
			if !opts.Shell && !opts.Git {
				opts.Git = true
			}

			if opts.Shell {
				shellType := "shell"
				opts.Prompt = "Convert this to a command: " +
					opts.Prompt +
					". Give me only the exact command to run, nothing else. " +
					"Choose the best " + shellType + " tool for the job. " +
					"Do not use dangerous system-modifying commands. " +
					"Use " + shellType + "-specific features when they would improve the command."
			}

			result, err := opts.Result()
			if err != nil {
				return err
			}

			if opts.Shell {
				// For shell mode, print the raw command from the response
				content := result.Explanation
				if content == "" {
					return errors.New(aiResponseErr)
				}
				// For shell mode, extract just the command without explanation
				// Extract and clean up command, removing any shell prefixes
				cmd := content
				if extracted := cmdExecRegexp.FindString(content); extracted != "" {
					cmd = strings.Trim(extracted, "```")
				} else if extracted := cmdHighlightRegexp.FindString(content); extracted != "" {
					cmd = strings.Trim(extracted, "`")
				}
				// Remove common shell prefixes and clean whitespace
				cmd = strings.TrimSpace(cmd)
				cmd = strings.TrimPrefix(cmd, "bash")
				cmd = strings.TrimPrefix(cmd, "sh")
				cmd = strings.TrimPrefix(cmd, "$")
				fmt.Fprint(opts.IO.StdOut, cmd) // Changed from Fprintln to Fprint
				return nil
			}

			opts.displayResult(result)

			if len(result.Commands) > 0 {
				if err := opts.executeCommands(result.Commands); err != nil {
					return err
				}
			}
			return nil
		},
	}

	duoAskCmd.Flags().BoolVarP(&opts.Git, "git", "", false, "Ask a question about Git")
	duoAskCmd.Flags().BoolVarP(&opts.Shell, "shell", "", false, "Generate shell commands from natural language")

	return duoAskCmd
}

func (opts *opts) Result() (*result, error) {
	opts.IO.StartSpinner(spinnerText)
	defer opts.IO.StopSpinner("")

	client, err := opts.HttpClient()
	if err != nil {
		return nil, cmdutils.WrapError(err, "failed to get HTTP client.")
	}

	apiPath := gitCmdAPIPath
	if opts.Shell {
		apiPath = chatAPIPath
	}
	var apiReq interface{}
	if opts.Shell {
		// For chat endpoint
		apiReq = map[string]string{
			"content": opts.Prompt,
		}
	} else {
		// For git command endpoint
		apiReq = request{Prompt: opts.Prompt, Model: vertexAI}
	}

	req, err := client.NewRequest(http.MethodPost, apiPath, apiReq, nil)
	if err != nil {
		return nil, cmdutils.WrapError(err, "failed to create a request.")
	}

	var content string
	if opts.Shell {
		var r chatResponse
		_, err = client.Do(req, &r)
		if err != nil {
			return nil, cmdutils.WrapError(err, apiUnreachableErr)
		}
		content = string(r)
	} else {
		var r gitResponse
		_, err = client.Do(req, &r)
		if err != nil {
			return nil, cmdutils.WrapError(err, apiUnreachableErr)
		}
		if len(r.Predictions) == 0 || len(r.Predictions[0].Candidates) == 0 {
			return nil, errors.New(aiResponseErr)
		}
		content = r.Predictions[0].Candidates[0].Content
	}

	var cmds []string
	for _, cmd := range cmdExecRegexp.FindAllString(content, -1) {
		cmds = append(cmds, strings.Trim(cmd, "\n`"))
	}

	return &result{
		Commands:    cmds,
		Explanation: content,
	}, nil
}

func (opts *opts) displayResult(result *result) {
	color := opts.IO.Color()

	opts.IO.LogInfo(color.Bold("Commands:\n"))

	for _, cmd := range result.Commands {
		opts.IO.LogInfo(color.Green(cmd))
	}

	opts.IO.LogInfo(color.Bold("\nExplanation:\n"))
	explanation := result.Explanation
	if opts.Git {
		explanation = cmdHighlightRegexp.ReplaceAllString(result.Explanation, color.Green("$1"))
	}
	opts.IO.LogInfo(explanation + "\n")
}

func (opts *opts) executeCommands(commands []string) error {
	if opts.Git {
		color := opts.IO.Color()

		var confirmed bool
		question := color.Bold(runCmdsQuestion)
		if err := prompt.Confirm(&confirmed, question, true); err != nil {
			return err
		}

		if !confirmed {
			return nil
		}
	}

	for _, command := range commands {
		if err := opts.executeCommand(command); err != nil {
			return err
		}
	}

	return nil
}

func (opts *opts) executeCommand(cmd string) error {
	gitArgs, err := shlex.Split(cmd)
	if err != nil {
		return nil
	}

	if gitArgs[0] != gitCmd {
		return nil
	}

	color := opts.IO.Color()
	question := fmt.Sprintf("Run `%s`", color.Green(cmd))
	var confirmed bool
	if err := prompt.Confirm(&confirmed, question, true); err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	execCmd := exec.Command("git", gitArgs[1:]...)
	output, err := run.PrepareCmd(execCmd).Output()
	if err != nil {
		return err
	}

	if len(output) == 0 {
		return nil
	}

	if opts.IO.StartPager() != nil {
		return fmt.Errorf("failed to start pager: %q", err)
	}
	defer opts.IO.StopPager()

	opts.IO.LogInfo(string(output))

	return nil
}
