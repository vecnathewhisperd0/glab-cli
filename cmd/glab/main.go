package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/pkg/iostreams"

	"github.com/mgutz/ansi"

	surveyCore "github.com/AlecAivazis/survey/v2/core"

	"gitlab.com/gitlab-org/cli/commands"
	"gitlab.com/gitlab-org/cli/commands/alias/expand"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/commands/help"
	"gitlab.com/gitlab-org/cli/commands/update"
	"gitlab.com/gitlab-org/cli/internal/config"
	"gitlab.com/gitlab-org/cli/internal/run"
	"gitlab.com/gitlab-org/cli/pkg/glinstance"
	"gitlab.com/gitlab-org/cli/pkg/tableprinter"
	"gitlab.com/gitlab-org/cli/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	// version is set dynamically at build
	version = "DEV"
	// buildDate is set dynamically at build
	buildDate string
	// platform is set dynamically at build
	platform = runtime.GOOS
)

// debug is set dynamically at build and can be overridden by
// the configuration file or environment variable
// sets to "true" or "false" or "1" or "0" as string
var debugMode = "false"

// debug is parsed boolean of debugMode
var debug bool

func main() {
	debug = debugMode == "true" || debugMode == "1"

	cmdFactory := cmdutils.NewFactory()

	cfg, err := cmdFactory.Config()
	if err != nil {
		cmdFactory.IO.Logf("failed to read configuration:  %s\n", err)
		os.Exit(2)
	}

	api.SetUserAgent(version, platform, runtime.GOARCH)
	maybeOverrideDefaultHost(cmdFactory, cfg)

	if !cmdFactory.IO.ColorEnabled() {
		surveyCore.DisableColor = true
	} else {
		// Override survey's choice of color for default values
		// For default values for e.g. `Input` prompts, Survey uses the literal "white" color,
		// which makes no sense on dark terminals and is literally invisible on light backgrounds.
		// This overrides Survey to output a gray color for 256-color terminals and "default" for basic terminals.
		surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
			switch style {
			case "white":
				if cmdFactory.IO.Is256ColorSupported() {
					return fmt.Sprintf("\x1b[%d;5;%dm", 38, 242)
				}
				return ansi.ColorCode("default")
			default:
				return ansi.ColorCode(style)
			}
		}
	}

	rootCmd := commands.NewCmdRoot(cmdFactory, version, buildDate)

	// Set Debug mode from config if not previously set by debugMode
	if !debug {
		debugModeCfg, _ := cfg.Get("", "debug")
		debug = debugModeCfg == "true" || debugModeCfg == "1"
	}

	if pager, _ := cfg.Get("", "glab_pager"); pager != "" {
		cmdFactory.IO.SetPager(pager)
	}

	if promptDisabled, _ := cfg.Get("", "no_prompt"); promptDisabled != "" {
		cmdFactory.IO.SetPrompt(promptDisabled)
	}

	if forceHyperlinks := os.Getenv("FORCE_HYPERLINKS"); forceHyperlinks != "" && forceHyperlinks != "0" {
		cmdFactory.IO.SetDisplayHyperlinks("always")
	} else if displayHyperlinks, _ := cfg.Get("", "display_hyperlinks"); displayHyperlinks == "true" {
		cmdFactory.IO.SetDisplayHyperlinks("auto")
	}

	var expandedArgs []string
	if len(os.Args) > 0 {
		expandedArgs = os.Args[1:]
	}

	cmd, _, err := rootCmd.Traverse(expandedArgs)
	// If a command was not found during traversal it will always return the rootCmd
	// which has no parent.
	if err != nil || !cmd.HasParent() {
		var (
			originalArgs = expandedArgs
			isShell      bool
		)

		if debug {
			fmt.Printf("%v -> %v\n", originalArgs, expandedArgs)
		}

		expandedArgs, isShell, err = expand.ExpandAlias(cfg, os.Args, nil)
		if err != nil {
			cmdFactory.IO.LogInfof("Failed to process alias: %s\n", err)
			os.Exit(2)
		}

		if isShell {
			runCmd(cmdFactory, expandedArgs)
		}

		if isPlugin(expandedArgs) {
			runPlugin(cmdFactory, expandedArgs)
		}
	}

	// Override the default column separator of tableprinter to double spaces
	tableprinter.SetTTYSeparator("  ")
	// Override the default terminal width of tableprinter
	tableprinter.SetTerminalWidth(cmdFactory.IO.TerminalWidth())
	// set whether terminal is a TTY or non-TTY
	tableprinter.SetIsTTY(cmdFactory.IO.IsOutputTTY())

	rootCmd.SetArgs(expandedArgs)

	if cmd, err := rootCmd.ExecuteC(); err != nil {
		printError(cmdFactory.IO, err, cmd, debug, true)
	}

	if help.HasFailed() {
		os.Exit(1)
	}

	checkUpdate, _ := cfg.Get("", "check_update")
	if checkUpdate, err := strconv.ParseBool(checkUpdate); err == nil && checkUpdate {

		var argCommand string

		if expandedArgs != nil {
			argCommand = expandedArgs[0]
		} else {
			argCommand = ""
		}

		err = update.CheckUpdate(cmdFactory, version, true, argCommand)
		if err != nil {
			printError(cmdFactory.IO, err, rootCmd, debug, false)
		}
	}

	api.GetClient().HTTPClient().CloseIdleConnections()
}

func printError(streams *iostreams.IOStreams, err error, cmd *cobra.Command, debug, shouldExit bool) {
	if errors.Is(err, cmdutils.SilentError) {
		return
	}
	color := streams.Color()
	printMore := true
	exitCode := 1

	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		streams.Logf("%s error connecting to %s\n", color.FailedIcon(), dnsError.Name)
		if debug {
			streams.Log(color.FailedIcon(), dnsError)
		}
		streams.Logf("%s Check your internet connection and status.gitlab.com. If on a self-managed instance, run 'sudo gitlab-ctl status' on your server.\n", color.DotWarnIcon())
		printMore = false
	}
	if printMore {
		var exitError *cmdutils.ExitError
		if errors.As(err, &exitError) {
			streams.Logf("%s %s %s=%s\n", color.FailedIcon(), color.Bold(exitError.Details), color.Red("error"), exitError.Err)
			exitCode = exitError.Code
			printMore = false
		}

		if printMore {
			streams.Log(err)

			var flagError *cmdutils.FlagError
			if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
				if !strings.HasSuffix(err.Error(), "\n") {
					streams.Log()
				}
				streams.Log(cmd.UsageString())
			}
		}
	}

	if cmd != nil {
		cmd.Print("\n")
	}
	if shouldExit {
		os.Exit(exitCode)
	}
}

func maybeOverrideDefaultHost(f *cmdutils.Factory, cfg config.Config) {
	baseRepo, err := f.BaseRepo()
	if err == nil {
		glinstance.OverrideDefault(baseRepo.RepoHost())
	}

	// Fetch the custom host config from env vars, then local config.yml, then global config,yml.
	customGLHost, _ := cfg.Get("", "host")
	if customGLHost != "" {
		if utils.IsValidURL(customGLHost) {
			var protocol string
			customGLHost, protocol = glinstance.StripHostProtocol(customGLHost)
			glinstance.OverrideDefaultProtocol(protocol)
		}
		glinstance.OverrideDefault(customGLHost)
	}
}

func isPlugin(args []string) bool {
	if len(args) == 0 {
		return false
	}
	searchTerm := "glab-" + args[0]
	_, err := exec.LookPath(searchTerm)
	return err == nil
}

func runCmd(utils *cmdutils.Factory, args []string) {
	externalCmd := exec.Command(args[0], args[1:]...)
	externalCmd.Stderr = os.Stderr
	externalCmd.Stdout = os.Stdout
	externalCmd.Stdin = os.Stdin

	preparedCmd := run.PrepareCmd(externalCmd)
	if err := preparedCmd.Run(); err != nil {
		handleCmdError(utils, err)
	}

	os.Exit(0)
}

func runPlugin(utils *cmdutils.Factory, args []string) {
	args[0] = "glab-" + args[0]
	runCmd(utils, args)
}

func handleCmdError(utils *cmdutils.Factory, err error) {
	utils.IO.LogInfof("failed to run external command: %s", err)
	if ee, ok := err.(*exec.ExitError); ok {
		os.Exit(ee.ExitCode())
	}
	os.Exit(3)
}
