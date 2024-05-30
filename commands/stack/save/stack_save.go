package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/git"
	"gitlab.com/gitlab-org/cli/pkg/prompt"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

var description string

const StackLocation = "/.git/refs/stacked/"

type StackRef struct {
	Prev        string `json:"prev"`
	Branch      string `json:"branch"`
	SHA         string `json:"sha"`
	Next        string `json:"next"`
	MR          string `json:"mr"`
	Description string `json:"description"`
}

func NewCmdSaveStack(f *cmdutils.Factory) *cobra.Command {
	stackSaveCmd := &cobra.Command{
		Use:   "save",
		Short: `Save your progress within stacked diff`,
		Long:  "\"save\" lets you save your current progress with a diff on the stack.\n",
		Example: heredoc.Doc(`
			glab stack save added_file
			glab stack save . -m "added a function"
			glab stack save -m "added a function"`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("message") && cmd.Flags().Changed("description") {
				return &cmdutils.FlagError{Err: errors.New("specify either of --message or --description")}
			}

			// check if there are even any changes before we start
			err := checkForChanges()
			if err != nil {
				return fmt.Errorf("could not save: %v", err)
			}

			// a description is required, so ask if one is not provided
			if description == "" {
				err := prompt.AskQuestionWithInput(&description, "description", "How would you describe this change?", "", true)
				if err != nil {
					return fmt.Errorf("error prompting for save description: %v", err)
				}
			}

			s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)

			// git add files
			_, err = addFiles(args[0:])
			if err != nil {
				return fmt.Errorf("error adding files: %v", err)
			}

			// get stack title
			title, err := git.GetCurrentStackTitle()
			if err != nil {
				return fmt.Errorf("error running git command: %v", err)
			}

			author, err := git.GitUserName()
			if err != nil {
				return fmt.Errorf("error getting git author: %v", err)
			}

			// generate a SHA based on: commit message, stack title, git author name
			sha, err := generateStackSha(description, title, string(author))
			if err != nil {
				return fmt.Errorf("error generating SHA command: %v", err)
			}

			// create branch name from SHA
			branch, err := createShaBranch(f, sha, title)
			if err != nil {
				return fmt.Errorf("error creating branch name: %v", err)
			}

			// create the branch prefix-stack_title-SHA
			err = git.CheckoutNewBranch(branch)
			if err != nil {
				return fmt.Errorf("error running branch checkout: %v", err)
			}

			// commit files to branch
			_, err = commitFiles(description)
			if err != nil {
				return fmt.Errorf("error committing files: %v", err)
			}

			refs, err := gatherStackRefs(title)
			if err != nil {
				return fmt.Errorf("error getting refs from file system: %v", err)
			}

			var stackRef StackRef
			if len(refs) > 0 {
				lastRef, err := lastRefInChain(refs)
				if err != nil {
					return fmt.Errorf("error finding last ref in chain: %v", err)
				}

				// update the ref before it (the current last ref)
				err = updateStackRefFile(title, StackRef{
					Prev:        lastRef.Prev,
					MR:          lastRef.MR,
					Description: lastRef.Description,
					SHA:         lastRef.SHA,
					Branch:      lastRef.Branch,
					Next:        sha,
				})
				if err != nil {
					return fmt.Errorf("error updating old ref: %v", err)
				}

				stackRef = StackRef{Prev: lastRef.SHA, SHA: sha, Branch: branch, Description: description}
			} else {
				stackRef = StackRef{SHA: sha, Branch: branch, Description: description}
			}

			err = addStackRefFile(title, stackRef)
			if err != nil {
				return fmt.Errorf("error creating stack file: %v", err)
			}

			if f.IO.IsOutputTTY() {
				color := f.IO.Color()

				fmt.Fprintf(
					f.IO.StdOut,
					"%s %s: Saved with message: \"%s\".\n",
					color.ProgressIcon(),
					color.Blue(title),
					description,
				)
			}

			s.Stop()

			return nil
		},
	}
	stackSaveCmd.Flags().StringVarP(&description, "description", "d", "", "a description of the change")
	stackSaveCmd.Flags().StringVarP(&description, "message", "m", "", "alias for description flag")

	return stackSaveCmd
}

func checkForChanges() error {
	gitCmd := git.GitCommand("status", "--porcelain")
	output, err := run.PrepareCmd(gitCmd).Output()
	if err != nil {
		return fmt.Errorf("error running git status: %v", err)
	}

	if string(output) == "" {
		return fmt.Errorf("no changes to save")
	}

	return nil
}

func addFiles(args []string) (files []string, err error) {
	if len(args) == 0 {
		args = []string{"."}
	}

	for _, file := range args {
		_, err = os.Stat(file)
		if err != nil {
			return
		}

		files = append(files, file)
	}

	cmdargs := append([]string{"add"}, args...)
	gitCmd := git.GitCommand(cmdargs...)

	_, err = run.PrepareCmd(gitCmd).Output()
	if err != nil {
		return []string{}, fmt.Errorf("error running git add: %v", err)
	}

	return files, err
}

func commitFiles(message string) (string, error) {
	commitCmd := git.GitCommand("commit", "-m", message)
	output, err := run.PrepareCmd(commitCmd).Output()
	if err != nil {
		return "", fmt.Errorf("error running git command: %v", err)
	}

	return string(output), nil
}

func generateStackSha(message string, title string, author string) (string, error) {
	toSha := message + title + author

	cmd := "echo \"" + toSha + "\" | git hash-object --stdin"
	sha, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("error running git hash-object: %v", err)
	}

	return strings.TrimSuffix(string(sha), "\n"), nil
}

func createShaBranch(f *cmdutils.Factory, sha string, title string) (string, error) {
	shortSha := string(sha)[0:8]

	cfg, err := f.Config()
	if err != nil {
		return "", fmt.Errorf("could not retrieve config file: %v", err)
	}

	prefix, err := cfg.Get("", "branch_prefix")
	if err != nil {
		return "", fmt.Errorf("could not get prefix config: %v", err)
	}

	if prefix == "" {
		prefix = os.Getenv("USER")
		if prefix == "" {
			prefix = "glab-stack"
		}
	}

	branchTitle := []string{prefix, title, shortSha}

	branch := strings.Join(branchTitle, "-")
	return string(branch), nil
}

func addStackRefFile(title string, stackRef StackRef) error {
	refDir, err := stackRootDir(title)
	if err != nil {
		return fmt.Errorf("error determining git root: %v", err)
	}

	initialJsonData, err := json.Marshal(stackRef)
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}

	if _, err = os.Stat(refDir); os.IsNotExist(err) {
		err = os.MkdirAll(refDir, 0o700) // create directory if it doesn't exist
		if err != nil {
			return fmt.Errorf("error creating directory: %v", err)
		}
	}

	fullPath := path.Join(refDir, stackRef.SHA+".json")

	err = os.WriteFile(fullPath, initialJsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error running writing file: %v", err)
	}

	return nil
}

func updateStackRefFile(title string, s StackRef) error {
	refDir, err := stackRootDir(title)
	if err != nil {
		return fmt.Errorf("error determining git root: %v", err)
	}

	fullPath := path.Join(refDir, s.SHA+".json")

	initialJsonData, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("error marshaling data: %v", err)
	}

	err = os.WriteFile(fullPath, initialJsonData, 0o644)
	if err != nil {
		return fmt.Errorf("error running writing file: %v", err)
	}

	return nil
}

func stackRootDir(title string) (string, error) {
	baseDir, err := git.ToplevelDir()
	if err != nil {
		return "", err
	}

	return path.Join(baseDir, StackLocation, title), nil
}

func gatherStackRefs(title string) ([]StackRef, error) {
	root, err := stackRootDir(title)
	if err != nil {
		return nil, err
	}

	var refs []StackRef
	err = filepath.WalkDir(root, func(dir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}
		// read files in the stacked ref directory
		// TODO: this may be quicker if we introduce a package
		// https://github.com/bmatcuk/doublestar
		if filepath.Ext(d.Name()) == ".json" {
			data, err := os.ReadFile(dir)
			if err != nil {
				return err
			}

			// marshal them into our StackRef type
			stackRef := StackRef{}
			err = json.Unmarshal(data, &stackRef)
			if err != nil {
				return err
			}

			refs = append(refs, stackRef)
		}

		return nil
	})
	if err != nil {
		if !os.IsNotExist(err) { // there might not be any refs yet, this is ok.
			return nil, err
		}
	}

	return refs, nil
}

func lastRefInChain(unsorted []StackRef) (StackRef, error) {
	index := slices.IndexFunc(unsorted, func(sr StackRef) bool { return sr.Next == "" })
	if index == -1 {
		return StackRef{}, fmt.Errorf("can't find the last ref in the chain. data might have been corrupted.")
	}

	return unsorted[index], nil
}

func sortRefs(unsorted []StackRef) ([]StackRef, error) {
	if len(unsorted) == 1 {
		return unsorted, nil // no need to sort if we only have one!
	}

	sorted := make([]StackRef, 0, len(unsorted))
	first := slices.IndexFunc(unsorted, func(sr StackRef) bool { return sr.Prev == "" })
	// find the first item in the chain: where the previous entry is nil
	sorted = append(sorted, unsorted[first])
	start := first

	for {
		next := slices.IndexFunc(unsorted, func(sr StackRef) bool { return sr.Prev == unsorted[start].SHA })
		// find the next item: find where the previous value == the current SHA

		if next == -1 {
			return []StackRef{}, errors.New("All entries have a non-empty Next ref and would infinite loop- check your data files")
		}

		sorted = append(sorted, unsorted[next])
		start = next
		// make the item we just found the start

		if unsorted[start].Next == "" {
			// if the item we found has a next == "", we're done
			break
		}
	}

	return sorted, nil
}
