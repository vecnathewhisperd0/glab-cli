package create_iam_policy

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

type IamPolicyCreateOptions struct {
	GcpProjectId              string `url:"gcp_project_id,omitempty"`
	GcpWorkloadIdentityPoolId string `url:"gcp_workload_identity_pool_id,omitempty"`
	OidcClaimName             string `url:"oidc_claim_name,omitempty"`
	OidcClaimValue            string `url:"oidc_claim_value,omitempty"`
	GcpIamRole                string `url:"gcp_iam_role,omitempty"`
}

func NewCmdCreateIamPolicy(f *cmdutils.Factory) *cobra.Command {
	opts := &IamPolicyCreateOptions{}

	cmd := &cobra.Command{
		Use:   "create-iam-policy [project] [flags]",
		Short: "Create a new IAM policy on GCP for the WLIF principalSet with a OIDC claim",
		Example: heredoc.Doc(`
		glab gcp create-iam-policy project-id \
			--gcp-project-id=project-id \
			--gcp-workload-identity-pool-id=pool-id \
			--oidc-claim-name=user_email \
			--oidc-claim-value="user@example.com" \
			--gcp-iam-role="compute.admin"
	`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HttpClient()
			if err != nil {
				return err
			}

			projectId := args[0]
			u := fmt.Sprintf("projects/%s/gcp_integration/iam_policies/create_script", url.PathEscape(projectId))
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

			script := exec.Command("bash", "-c", buf.String())
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
	cmd.Flags().StringVar(&opts.OidcClaimName, "oidc-claim-name", "", "OIDC claim name for the GCP Workload Identity Federation principalSet")
	cmd.Flags().StringVar(&opts.OidcClaimValue, "oidc-claim-value", "", "OIDC claim value for the GCP Workload Identity Federation principalSet")
	cmd.Flags().StringVar(&opts.GcpIamRole, "gcp-iam-role", "", "GCP IAM role")

	return cmd
}
