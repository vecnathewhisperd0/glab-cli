package variableutils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"
)

func GetValue(value string, ios *iostreams.IOStreams, args []string) (string, error) {
	if value != "" {
		return value, nil
	} else if len(args) == 2 {
		return args[1], nil
	}

	if ios.IsInTTY {
		return "", &cmdutils.FlagError{Err: errors.New("no value specified but nothing on STDIN")}
	}

	defer func() {
		if err := ios.In.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	stdinValue, err := io.ReadAll(ios.In)
	if err != nil {
		return "", fmt.Errorf("failed to read value from STDIN: %w", err)
	}
	return strings.TrimSpace(string(stdinValue)), nil
}
