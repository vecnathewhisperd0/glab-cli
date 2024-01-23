package create_wlif

import (
	"bytes"
	"fmt"
	"strings"

	"net/http"
	"net/url"

	"os/exec"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

type WlifCreateOptions struct {
	GcpProjectId                               string `url:"gcp_project_id,omitempty"`
	GcpWorkloadIdentityPoolId                  string `url:"gcp_workload_identity_pool_id,omitempty"`
	GcpWorkloadIdentityPoolDisplayName         string `url:"gcp_workload_identity_pool_display_name,omitempty"`
	GcpWorkloadIdentityPoolProviderId          string `url:"gcp_workload_identity_pool_provider_id,omitempty"`
	GcpWorkloadIdentityPoolProviderDisplayName string `url:"gcp_workload_identity_pool_provider_display_name,omitempty"`
}

func NewCmdCreateWlif(f *cmdutils.Factory) *cobra.Command {
	opts := &WlifCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create-wlif [project] [flags]",
		Short: "Create a new Workload Identity Federation on GCP",
		Example: heredoc.Doc(`
		glab gcp create-wlif project-id \
			--gcp-project-id=project-id \
			--gcp-workload-identity-pool-id=pool-id \
			--gcp-workload-identity-pool-display-name=pool-display-name \
			--gcp-workload-identity-pool-provider-id=provider-id \
			--gcp-workload-identity-pool-provider-display-name=provider-display-name
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			projectId := args[0]
			u := fmt.Sprintf("projects/%s/gcp_integration/wlif/create_script", url.PathEscape(projectId))
			request, err := client.NewRequest(http.MethodGet, u, opts, nil)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			_, err = client.Do(request, &buf)
			if err != nil {
				fmt.Println(err)

				return nil
			}

			script := exec.Command("sh")
			script.Stdin = strings.NewReader(fmt.Sprintf("echo '%s' | sh", buf.String()))
			output, err := script.CombinedOutput()
			if err != nil {
				fmt.Println("Error running script")
				fmt.Println(err)
			}
			fmt.Println(string(output))

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.GcpProjectId, "gcp-project-id", "", "GCP Project ID for the Workload Identity Federation")
	cmd.Flags().StringVar(&opts.GcpWorkloadIdentityPoolId, "gcp-workload-identity-pool-id", "", "GCP Workload Identity Pool ID")
	cmd.Flags().StringVar(&opts.GcpWorkloadIdentityPoolDisplayName, "gcp_workload_identity_pool_display_name", "", "GCP Workload Identity Pool display name")
	cmd.Flags().StringVar(&opts.GcpWorkloadIdentityPoolProviderId, "gcp_workload_identity_pool_provider_id", "", "GCP Workload Identity Pool Provider ID")
	cmd.Flags().StringVar(&opts.GcpWorkloadIdentityPoolProviderDisplayName, "gcp_workload_identity_pool_provider_display_name", "", "GCP Workload Identity Pool Provider display name")

	return cmd
}
