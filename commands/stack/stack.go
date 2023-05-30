package stack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/prompt"
)

const (
	configRef     = "refs/stack-config"
	stackDataRefs = "refs/diffs/"
)

func NewCmdStack(f *cmdutils.Factory) *cobra.Command {
	stackCmd := &cobra.Command{
		Use:     "stack [command] [flags]",
		Short:   "Work with stacked diffs",
		Aliases: []string{"st"},
		Long:    ``,
		Example: heredoc.Doc(`
			glab stack init
		`),
	}

	stackCmd.AddCommand(NewCmdStackInit(f))
	stackCmd.AddCommand(NewCmdStackCommit(f))
	return stackCmd
}

func NewCmdStackInit(f *cmdutils.Factory) *cobra.Command {
	var stack Stack
	cmd := &cobra.Command{
		Use:   "init [title]",
		Short: "Initialize a stacking configuration",
		Long:  ``,
		Example: heredoc.Doc(`
			glab stack init	
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			stack.Init()
			return nil
		},
	}
	return cmd
}

func NewCmdStackCommit(f *cmdutils.Factory) *cobra.Command {
	var stack Stack
	var branch string
	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Create a new diff on the diff stack",
		Long:  ``,
		Example: heredoc.Doc(`
		glab stack commit
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return stack.Commit(branch, args...)
		},
	}
	cmd.Flags().StringVarP(&branch, "", "b", "", "name of branch to use for the diff")
	cmd.MarkFlagRequired("b")
	return cmd
}

// Stack is a collection of changes that
// collectively introduce a change. The
// Stack abstraction allows for easier
// reviews of large changes that should
// be introduced as a single change set
// but would otherwise be too large to
// change.
type Stack struct {
	// Diffs contains the diff stack entries related to the Stack.
	Diffs []Diff `json:"diffs"`
}

// Diff represents an entry in the Stack of
// changes. Each Diff is a singular atomic
// change and is converted to a merge request
// when the stack is pushed.
type Diff struct {
	// Ready means that the branch contains changes
	Ready      bool   `json:"ready"`
	PrevBranch string `json:"prev_branch"`
	NextBranch string `json:"next_branch"`
	// Branch that holds the changes for the diff
	Branch string `json:"branch"`
	// MergeRequest created that will introduce change.
	MergeRequest string `json:"merge_request,omitempty"`
}

func listBranches() ([]string, error) {
	cmd := git.Command("branch", "--no-color")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("getting local branches cmd: %s output: %s: %w", cmd, out, err)
	}
	branches := strings.Fields(string(out))
	return branches, nil
}

func pushConfigRef() error {
	cmd := git.Command("push", "origin", configRef+":"+configRef)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", cmd, err)
	}
	return nil
}

type Config struct {
	DefaultBranch string `json:"default_branch"`
}

func loadConfig() (Config, error) {
	var cfg Config
	objectID, err := os.ReadFile(".git/" + configRef)
	if err != nil {
		return cfg, fmt.Errorf("reading config ref: %w", err)
	}
	objectID = bytes.TrimSpace(objectID)
	r, err := gitCatFile(string(objectID))
	if err != nil {
		return cfg, fmt.Errorf("reading git object: %w", err)
	}
	err = json.NewDecoder(r).Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("decoding stacking config: %w", err)
	}
	return cfg, nil
}

// Init initializes the stack config required to
// track the changes in a decentralized manner.
func (s Stack) Init() error {
	branches, err := listBranches()
	if err != nil {
		return fmt.Errorf("attempting to locate the default branch automatically: %w", err)
	}
	var defaultBranch string
	for _, branch := range branches {
		if branch == "main" || branch == "master" || branch == "trunk" {
			defaultBranch = branch
			break
		}
	}

	err = prompt.AskOne(&survey.Input{Message: "Select the default branch for the repository", Default: defaultBranch}, &defaultBranch)
	if err != nil {
		return fmt.Errorf("configuring default branch: %w", err)
	}

	cfg := Config{DefaultBranch: defaultBranch}
	b, err := json.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshalling stacking config to json: %w", err)
	}
	r := bytes.NewReader(b)
	sha, err := gitHashObject(r)
	if err != nil {
		return fmt.Errorf("saving initial config file: %w", err)
	}

	err = gitUpdateRef(configRef, sha)
	if err != nil {
		return fmt.Errorf("updating config ref: %w", err)
	}

	err = pushConfigRef()
	if err != nil {
		return fmt.Errorf("pushing config ref: %w", err)
	}

	return nil
}

func gitHashObject(r io.Reader) (string, error) {
	var stdout, stderr strings.Builder
	cmd := git.Command("hash-object", "-w", "--stdin")
	cmd.Stdin = r
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("initializing stack stderr: %s: %w", stderr.String(), err)
	}
	sha := strings.TrimSpace(stdout.String())
	return sha, nil
}

func gitUpdateRef(ref, value string) error {
	var stderr strings.Builder
	cmd := git.Command("update-ref", ref, value)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("updating stack entry stderr: %s: %w", stderr.String(), err)
	}
	return nil
}

func (s Stack) currentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running %s out: %s err: %w", cmd.String(), out, err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "", fmt.Errorf("no branch name: currently in a detached head state")
	}
	return branch, nil
}

func (s Stack) Stage(args ...string) error {
	addCmd := []string{"add"}
	args = append(addCmd, args...)
	cmd := git.Command(args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("staging changes in working diff: %w", err)
	}
	return nil
}

// Commit creates a new diff entry in the stack.
func (s *Stack) Commit(branch string, args ...string) error {
	prevBranch, err := s.currentBranch()
	if err != nil {
		return fmt.Errorf("getting previous branch name: %w", err)
	}

	diff := Diff{
		Ready:        false,
		PrevBranch:   prevBranch,
		NextBranch:   "",
		Branch:       branch,
		MergeRequest: "",
	}
	err = updateDiff(diff)
	if err != nil {
		return err
	}

	err = commit(args...)
	if err != nil {
		return fmt.Errorf("committing diff to stack: %w", err)
	}

	err = createBranch(branch)
	if err != nil {
		return fmt.Errorf("switching to new working diff: %w", err)
	}

	// After successfully committing changes update the previous branch
	// so that we link the previous diff to the current one.
	prevDiff, err := s.readDiffMetadata(prevBranch)
	if err != nil {
		return fmt.Errorf("getting previous diff: %w", err)
	}
	prevDiff.NextBranch = branch
	err = updateDiff(prevDiff)
	if err != nil {
		return err
	}
	return nil
}

func createBranch(branch string) error {
	cmd := git.Command("switch", "-c", branch)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("running %s: %w", cmd, err)
	}
	return nil
}

func commit(args ...string) error {
	cmd := git.Command("commit")
	cmd.Args = append(cmd.Args, args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("committing diff to stack: %w", err)
	}
	return nil
}

func gitCatFile(objectID string) (io.Reader, error) {
	var stdout, stderr bytes.Buffer
	cmd := git.Command("cat-file", "-p", objectID)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return &stdout, nil
}

func updateDiff(diff Diff) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// There's never a diff entry for the default branch by definition.
	if diff.Branch == cfg.DefaultBranch {
		return nil
	}
	b, err := json.Marshal(&diff)
	if err != nil {
		return err
	}

	r := bytes.NewReader(b)
	sha, err := gitHashObject(r)
	if err != nil {
		return err
	}

	path := refPath(diff.Branch)
	err = gitUpdateRef(path, sha)
	if err != nil {
		return err
	}
	return nil
}

func refPath(branch string) string {
	return stackDataRefs + branch
}

func (s Stack) readDiffMetadata(branch string) (diff Diff, err error) {
	b, err := os.ReadFile(refPath(branch))
	if err != nil {
		return
	}
	objectID := string(b)

	var stderr strings.Builder
	cmd := git.Command("cat-file", "-p", objectID)
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	err = json.NewDecoder(stdout).Decode(&diff)
	if err != nil {
		return
	}

	err = cmd.Wait()
	return
}

// List visualizes the stack of diffs.
func (s Stack) List() {
}

// Push upserts corresponding MRs for each Diff in the Stack.
func (s Stack) Push() {}

// Checkout pulls in the changes in a stack so they can be viewed locally.
func (s Stack) Checkout() {}
