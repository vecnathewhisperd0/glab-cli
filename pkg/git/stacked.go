package git

import (
	"fmt"
	"os"
	"path"

	"gitlab.com/gitlab-org/cli/internal/run"
)

type StackRef struct {
	Previous string `json:"previous"`
	Next     string `json:"next"`
	MR       string `json:"mr"`
	ChangeId string `json:"change-id"`
}

func GetCurrentStackTitle() (title string, err error) {
	title, err = Config("glab.currentstack")
	return
}

func SetLocalConfig(key, value string) error {
	found, err := configValueExists(key, value)
	if err != nil {
		return fmt.Errorf("git config value exists: %w", err)
	}

	if found {
		return nil
	}

	addCmd := GitCommand("config", "--local", key, value)
	_, err = run.PrepareCmd(addCmd).Output()
	if err != nil {
		return fmt.Errorf("setting local git config: %w", err)
	}
	return nil
}

func AddStackRefDir(dir string) (string, error) {
	baseDir, err := ToplevelDir()
	if err != nil {
		return "", fmt.Errorf("finding top level git directory: %w", err)
	}

	createdDir := path.Join(baseDir, "/.git/refs/stacked/", dir)

	err = os.MkdirAll(createdDir, 0o700)
	if err != nil {
		return "", fmt.Errorf("creating stacked diff directory: %w", err)
	}

	return createdDir, nil
}
