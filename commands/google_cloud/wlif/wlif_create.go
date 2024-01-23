package wlif

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
)

const (
	flagGoogleCloudProjectID = "google-cloud-project-id"
)

type WLIFCreateOptions struct {
	GoogleCloudProjectID                               string `url:"google_cloud_project_id,omitempty"`
	GoogleCloudWorkloadIdentityPoolID                  string `url:"google_cloud_workload_identity_pool_id,omitempty"`
	GoogleCloudWorkloadIdentityPoolDisplayName         string `url:"google_cloud_workload_identity_pool_display_name,omitempty"`
	GoogleCloudWorkloadIdentityPoolProviderID          string `url:"google_cloud_workload_identity_pool_provider_id,omitempty"`
	GoogleCloudWorkloadIdentityPoolProviderDisplayName string `url:"google_cloud_workload_identity_pool_provider_display_name,omitempty"`
}

func NewCmdCreateWLIF(f *cmdutils.Factory) *cobra.Command {
	opts := &WLIFCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create-wlif [project] [flags]",
		Short: "EXPERIMENTAL: Create a new Workload Identity Federation on Google Cloud",
		Example: heredoc.Doc(`
		$ glab google-cloud create-wlif project-id \
		  --google-cloud-project-id=google-cloud-project-id \
		  --google-cloud-workload-identity-pool-id=pool-id \
		  --google-cloud-workload-identity-pool-display-name=pool-display-name \
		  --google-cloud-workload-identity-pool-provider-id=provider-id \
		  --google-cloud-workload-identity-pool-provider-display-name=provider-display-name
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			projectID := args[0]
			u := fmt.Sprintf("projects/%s/scripts/google_cloud/create_wlif", url.PathEscape(projectID))
			request, err := client.NewRequest(http.MethodGet, u, opts, nil)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			res, err := client.Do(request, &buf)
			if err != nil {
				return err
			}

			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status code: %d", res.StatusCode)
			}

			script := exec.Command("bash", "-c", buf.String())
			output, err := script.CombinedOutput()
			fmt.Println(string(output))
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.GoogleCloudProjectID, flagGoogleCloudProjectID, "", "Google Cloud Project ID for the Workload Identity Federation")
	cmd.Flags().StringVar(&opts.GoogleCloudWorkloadIdentityPoolID, "google-cloud-workload-identity-pool-id", "", "ID of the Google Cloud Workload Identity Pool to be created")
	cmd.Flags().StringVar(&opts.GoogleCloudWorkloadIdentityPoolDisplayName, "google-cloud-workload-identity-pool-display-name", "", "display name of the Google Cloud Workload Identity Pool to be created")
	cmd.Flags().StringVar(&opts.GoogleCloudWorkloadIdentityPoolProviderID, "google-cloud-workload-identity-pool-provider-id", "", "ID of the Google Cloud Workload Identity Pool Provider to be created")
	cmd.Flags().StringVar(&opts.GoogleCloudWorkloadIdentityPoolProviderDisplayName, "google-cloud-workload-identity-pool-provider-display-name", "", "display name of the Google Cloud Workload Identity Pool Provider to be created")
	cobra.CheckErr(cmd.MarkFlagRequired(flagGoogleCloudProjectID))

	return cmd
}
