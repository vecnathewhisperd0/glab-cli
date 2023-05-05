package connect

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"gitlab.com/gitlab-org/cli/pkg/utils"
)

func NewCmdConnect(f *cmdutils.Factory) *cobra.Command {
	jobConnectCmd := &cobra.Command{
		Use:     "connect <job-id>",
		Short:   "Connect to a CI/CD job",
		Aliases: []string{},
		Example: heredoc.Doc(`
	glab ci connect 123456
`),
		Long: "",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := utils.StringToInt(args[0])
			if jobID < 1 {
				fmt.Fprintln(f.IO.StdErr, "invalid job id:", args[0])
				return cmdutils.SilentError
			}
			data, err := ioutil.ReadFile("/home/josephburnett/.ssh/id_ed25519_work.pub")
			if err != nil {
				fmt.Fprintln(f.IO.StdErr, err)
				return cmdutils.SilentError
			}
			resp, err := http.Post("http://localhost:12345?jobID="+strconv.Itoa(jobID), "text", bytes.NewReader(data))
			if err != nil {
				fmt.Fprintln(f.IO.StdErr, err)
				return cmdutils.SilentError
			}
			if resp.StatusCode != 200 {
				msg, _ := ioutil.ReadAll(resp.Body)
				fmt.Fprintln(f.IO.StdErr, string(msg))
				return cmdutils.SilentError
			}
			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintln(f.IO.StdErr, err)
				return cmdutils.SilentError
			}
			ipAddress := strings.TrimSpace(string(data))
			if ipAddress == "" {
				fmt.Fprintln(f.IO.StdErr, "no public IP address")
				return cmdutils.SilentError
			}
			shell := exec.Command("ssh", "-i", "/home/josephburnett/.ssh/id_ed25519_work", "josephburnett@"+ipAddress)
			shell.Stdin = os.Stdin
			shell.Stdout = os.Stdout
			shell.Stderr = os.Stderr
			_ = shell.Run()
			return nil
		},
	}
	return jobConnectCmd
}
