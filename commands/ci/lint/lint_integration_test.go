package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
	"gitlab.com/gitlab-org/cli/test"

	"gitlab.com/gitlab-org/cli/commands/cmdtest"
)

func TestMain(m *testing.M) {
	cmdtest.InitTest(m, "ci_lint_test")
}

func Test_pipelineCILint_Integration(t *testing.T) {
	glTestHost := test.GetHostOrSkip(t)

	io, _, stdout, stderr := iostreams.Test()
	fac := cmdtest.StubFactory(glTestHost + "/cli-automated-testing/test")
	fac.IO = io
	fac.IO.StdErr = stderr
	fac.IO.StdOut = stdout

	tests := []struct {
		Name    string
		Args    string
		StdOut  string
		WantErr error
	}{
		{
			Name:   "with no path specified",
			Args:   "",
			StdOut: "Validating...\n✓ CI/CD YAML is valid!\n",
		},
		{
			Name:   "with path specified as url",
			Args:   glTestHost + "/cli-automated-testing/test/-/raw/master/.gitlab-ci.yml",
			StdOut: "Validating...\n✓ CI/CD YAML is valid!\n",
		},
	}

	cmd := NewCmdLint(fac)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := cmdtest.RunCommand(cmd, test.Args)
			if err != nil {
				if test.WantErr == nil {
					t.Fatal(err)
				}
				assert.Equal(t, err, test.WantErr)
			}
			assert.Equal(t, test.StdOut, stdout.String())
			stdout.Reset()
		})
	}
}
