package trace

import (
	"bytes"
	"testing"

	"gitlab.com/gitlab-org/cli/test"

	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/cli/pkg/prompt"

	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/internal/config"

	"gitlab.com/gitlab-org/cli/commands/cmdtest"
)

var (
	stubFactory *cmdutils.Factory
	cmd         *cobra.Command
	stdout      *bytes.Buffer
)

func TestMain(m *testing.M) {
	cmdtest.InitTest(m, "ci_trace_test")
}

func TestNewCmdTrace_Integration(t *testing.T) {
	glTestHost := test.GetHostOrSkip(t)

	defer config.StubConfig(`---
git_protocol: https
hosts:
  gitlab.com:
    username: root
`, "")()

	var io *iostreams.IOStreams
	io, _, stdout, _ = iostreams.Test()
	stubFactory, _ = cmdtest.StubFactoryWithConfig(glTestHost + "/cli-automated-testing/test")
	stubFactory.IO = io
	stubFactory.IO.IsaTTY = true
	stubFactory.IO.IsErrTTY = true

	tests := []struct {
		name     string
		args     string
		wantOpts *TraceOpts
	}{
		{
			name: "Has no arg",
			args: ``,
			wantOpts: &TraceOpts{
				Branch: "master",
				JobID:  0,
			},
		},
		{
			name: "Has arg with job-id",
			args: `3509632552`,
			wantOpts: &TraceOpts{
				Branch: "master",
				JobID:  3509632552,
			},
		},
		{
			name: "On a specified repo with job ID",
			args: "3509632552 -X cli-automated-testing/test",
			wantOpts: &TraceOpts{
				Branch: "master",
				JobID:  3509632552,
			},
		},
	}

	var actualOpts *TraceOpts
	cmd = NewCmdTrace(stubFactory, func(opts *TraceOpts) error {
		actualOpts = opts
		return nil
	})
	cmd.Flags().StringP("repo", "X", "", "")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantOpts.IO = stubFactory.IO

			argv, err := shlex.Split(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs(argv)
			_, err = cmd.ExecuteC()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.wantOpts.JobID, actualOpts.JobID)
			assert.Equal(t, tt.wantOpts.Branch, actualOpts.Branch)
			assert.Equal(t, tt.wantOpts.Branch, actualOpts.Branch)
			assert.Equal(t, tt.wantOpts.IO, actualOpts.IO)
		})
	}
}

func TestTraceRun(t *testing.T) {
	glTestHost := test.GetHostOrSkip(t)

	var io *iostreams.IOStreams
	io, _, stdout, _ = iostreams.Test()
	stubFactory = cmdtest.StubFactory(glTestHost + "/cli-automated-testing/test")
	stubFactory.IO = io
	stubFactory.IO.IsaTTY = true
	stubFactory.IO.IsErrTTY = true

	tests := []struct {
		desc           string
		args           string
		assertContains func(t *testing.T, out string)
	}{
		{
			desc: "Has no arg",
			args: ``,
			assertContains: func(t *testing.T, out string) {
				assert.Contains(t, out, "Getting job trace...")
				assert.Contains(t, out, "Showing logs for ")
				assert.Contains(t, out, `Preparing the "docker+machine"`)
				assert.Contains(t, out, `$ echo "This is a after script section"`)
				assert.Contains(t, out, "Job succeeded")
			},
		},
		{
			desc: "Has arg with job-id",
			args: `3509632552`,
			assertContains: func(t *testing.T, out string) {
				assert.Contains(t, out, "Getting job trace...\n")
			},
		},
		{
			desc: "On a specified repo with job ID",
			args: "3509632552 -X cli-automated-testing/test",
			assertContains: func(t *testing.T, out string) {
				assert.Contains(t, out, "Getting job trace...\n")
			},
		},
	}

	cmd = NewCmdTrace(stubFactory, nil)
	cmd.Flags().StringP("repo", "X", "", "")

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if tt.args == "" {
				as, teardown := prompt.InitAskStubber()
				defer teardown()

				as.StubOne("cleanup4 (3509632552) - success")
			}
			argv, err := shlex.Split(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs(argv)
			_, err = cmd.ExecuteC()
			if err != nil {
				t.Fatal(err)
			}
			tt.assertContains(t, stdout.String())
		})
	}
}
