package members

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

var tests = []struct {
	expectedUse   string
	expectedShort string
	subcommands   []string
}{
	{
		expectedUse:   "members [command] [flags]",
		expectedShort: "Manage project members.",
		subcommands:   []string{"add", "remove"},
	},
}

func TestNewCmdMembers(t *testing.T) {
	cmd := NewCmdMembers(&cmdutils.Factory{})
	assert.Equal(t, tests[0].expectedUse, cmd.Use)
	assert.Equal(t, tests[0].expectedShort, cmd.Short)

	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewCmdMembersSubcommands(t *testing.T) {
	cmd := NewCmdMembers(&cmdutils.Factory{})
	subcommands := cmd.Commands()
	assert.NotNilf(t, subcommands, "expected subcommands not to be nil")
	assert.True(t, cmd.HasAvailableSubCommands())
	assert.Len(t, subcommands, len(tests[0].subcommands))
}
