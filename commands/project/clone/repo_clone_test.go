package clone

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/spf13/cobra"

	"github.com/google/shlex"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/test"

	"github.com/stretchr/testify/assert"

	"gitlab.com/gitlab-org/cli/commands/cmdtest"
)

func TestMain(m *testing.M) {
	cmdtest.InitTest(m, "repo_clone_test")
}

func runCommand(cmd *cobra.Command, cli string, stds ...*bytes.Buffer) (*test.CmdOut, error) {
	var stdin *bytes.Buffer
	var stderr *bytes.Buffer
	var stdout *bytes.Buffer

	for i, std := range stds {
		if std != nil {
			if i == 0 {
				stdin = std
			}
			if i == 1 {
				stdout = std
			}
			if i == 2 {
				stderr = std
			}
		}
	}
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	argv, err := shlex.Split(cli)
	if err != nil {
		return nil, err
	}
	cmd.SetArgs(argv)
	_, err = cmd.ExecuteC()

	return &test.CmdOut{
		OutBuf: stdout,
		ErrBuf: stderr,
	}, err
}

func TestNewCmdClone(t *testing.T) {
	testCases := []struct {
		name        string
		args        string
		wantOpts    CloneOptions
		wantCtxOpts ContextOpts
		wantErr     string
	}{
		{
			name:    "no arguments",
			args:    "",
			wantErr: "specify repo argument or use --group flag to specify a group to clone all repos from the group",
		},
		{
			name: "repo argument",
			args: "NAMESPACE/REPO",
			wantOpts: CloneOptions{
				GitFlags: []string{},
			},
			wantCtxOpts: ContextOpts{
				Repo: "NAMESPACE/REPO",
			},
		},
		{
			name: "directory argument",
			args: "NAMESPACE/REPO mydir",
			wantOpts: CloneOptions{
				GitFlags: []string{"mydir"},
			},
			wantCtxOpts: ContextOpts{
				Repo: "NAMESPACE/REPO",
			},
		},
		{
			name: "git clone arguments",
			args: "NAMESPACE/REPO -- --depth 1 --recurse-submodules",
			wantOpts: CloneOptions{
				GitFlags: []string{"--depth", "1", "--recurse-submodules"},
			},
			wantCtxOpts: ContextOpts{
				Repo: "NAMESPACE/REPO",
			},
		},
		{
			name:    "unknown argument",
			args:    "NAMESPACE/REPO --depth 1",
			wantErr: "unknown flag: --depth\nSeparate git clone flags with '--'.",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			io, stdin, stdout, stderr := iostreams.Test()
			fac := &cmdutils.Factory{IO: io}

			var opts *CloneOptions
			var ctxOpts *ContextOpts
			cmd := NewCmdClone(fac, func(co *CloneOptions, cx *ContextOpts) error {
				opts = co
				ctxOpts = cx
				return nil
			})

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)

			cmd.SetIn(stdin)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			_, err = cmd.ExecuteC()
			if err != nil {
				assert.Equal(t, tt.wantErr, err.Error())
				return
			} else if tt.wantErr != "" {
				t.Errorf("expected error %q, got nil", tt.wantErr)
			}

			assert.Equal(t, "", stdout.String())
			assert.Equal(t, "", stderr.String())

			assert.Equal(t, tt.wantCtxOpts.Repo, ctxOpts.Repo)
			assert.Equal(t, tt.wantOpts.GitFlags, opts.GitFlags)
		})
	}
}

func Test_repoClone(t *testing.T) {
	defer config.StubConfig(`
hosts:
  gitlab.com:
    token: qRC87Xg9Wd46RhB8J8sp
`, "")()
	token := os.Getenv("GITLAB_TOKEN")
	if token != "" {
		_ = os.Setenv("GITLAB_TOKEN", "")
	}
	io, stdin, stdout, stderr := iostreams.Test()
	fac := &cmdutils.Factory{
		IO: io,
		Config: func() (config.Config, error) {
			return config.ParseConfig("config.yml")
		},
	}

	cs, restore := test.InitCmdStubber()
	// git clone
	cs.Stub("")
	// git remote add
	cs.Stub("")
	defer restore()

	cmd := NewCmdClone(fac, nil)
	out, err := runCommand(cmd, "test", stdin, stdout, stderr)
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}

	assert.Equal(t, "", out.String())
	assert.Equal(t, "", out.Stderr())
	assert.Equal(t, 1, cs.Count)
	assert.Equal(t, "git clone git@gitlab.com:clemsbot/test.git", strings.Join(cs.Calls[0].Args, " "))
	if token != "" {
		_ = os.Setenv("GITLAB_TOKEN", token)
	}
}

func Test_repoClone_group(t *testing.T) {
	var names = []string{"glab-cli/test", "glab-cli/test-pv"}
	var urls = []string{"git@gitlab.com:glab-cli/test.git", "git@gitlab.com:glab-cli/test-pv.git"}
	repoCloneTest(t, names, urls, 0, false)
}

func Test_repoClone_group_single(t *testing.T) {
	var names = []string{"glab-cli/test"}
	var urls = []string{"git@gitlab.com:glab-cli/test.git"}
	repoCloneTest(t, names, urls, 1, false)
}

func Test_repoClone_group_paginate(t *testing.T) {
	var names = []string{"glab-cli/test", "glab-cli/test-pv"}
	var urls = []string{"git@gitlab.com:glab-cli/test.git", "git@gitlab.com:glab-cli/test-pv.git"}
	repoCloneTest(t, names, urls, 1, true)
}

func repoCloneTest(t *testing.T, expectedRepoNames []string, expectedRepoUrls []string, perPage int, paginate bool) {
	assert.Equal(t, len(expectedRepoNames), len(expectedRepoUrls))

	defer config.StubConfig(`
hosts:
  gitlab.com:
    token: qRC87Xg9Wd46RhB8J8sp
`, "")()
	token := os.Getenv("GITLAB_TOKEN")
	if token != "" {
		_ = os.Setenv("GITLAB_TOKEN", "")
	}
	io, stdin, stdout, stderr := iostreams.Test()
	fac := &cmdutils.Factory{
		IO: io,
		Config: func() (config.Config, error) {
			return config.ParseConfig("config.yml")
		},
	}

	cs, restore := test.InitCmdStubber()
	for i := 0; i < len(expectedRepoUrls); i++ {
		cs.Stub("")
	}

	defer restore()

	cmd := NewCmdClone(fac, nil)
	var cli = "-g glab-cli"
	if perPage != 0 {
		cli += fmt.Sprintf(" --per-page %d", perPage)
	}
	if paginate {
		cli += " --paginate"
	}

	// TODO: stub api.ListGroupProjects endpoint
	out, err := runCommand(cmd, cli, stdin, stdout, stderr)
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}

	assert.Equal(t, "✓ "+strings.Join(expectedRepoNames, "\n✓ ")+"\n", out.String())
	assert.Equal(t, "", out.Stderr())
	assert.Equal(t, len(expectedRepoUrls), cs.Count)

	for i := 0; i < len(expectedRepoUrls); i++ {
		assert.Equal(t, fmt.Sprintf("git clone %s", expectedRepoUrls[i]), strings.Join(cs.Calls[i].Args, " "))
	}

	if token != "" {
		_ = os.Setenv("GITLAB_TOKEN", token)
	}
}
