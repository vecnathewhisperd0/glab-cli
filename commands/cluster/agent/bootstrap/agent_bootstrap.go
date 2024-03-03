package list

import (
	"fmt"
	"os/exec"
	"time"

	"gitlab.com/gitlab-org/cli/api"
	"gitlab.com/gitlab-org/cli/commands/cluster/agent/agentutils"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"

	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

var factory *cmdutils.Factory

func AgentBootstrapCmd(f *cmdutils.Factory) *cobra.Command {
	factory = f
	environmentCreateCmd := &cobra.Command{
		Use:   "bootstrap [flags]",
		Short: `Bootstrap a GitLab - cluster connection`,
		Long: `Bootstraps a cluster connection including:
		
		- uses External Secrets and masked environment variables to synchronize tokens from Gitlab to the cluster
		- uses Flux for GitOps application deployment
		- uses the agent for Kubernetes for bidirectional GitLab - cluster communication
		`,
		Args: cobra.MaximumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			factory = f
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			manifestDir, err := cmd.Flags().GetString("manifest-dir")
			if err != nil {
				return err
			}
			skipExternalSecrets, err := cmd.Flags().GetBool("skip-external-secrets")
			if err != nil {
				return err
			}
			return bootstrapAgent(name, manifestDir, skipExternalSecrets)
		},
	}
	environmentCreateCmd.Flags().StringP("name", "n", "", "Name of the new environment")
	environmentCreateCmd.Flags().StringP("manifest-dir", "", "manifests", "Base directory for manifests")
	environmentCreateCmd.Flags().BoolP("skip-external-secrets", "", false, "Skips creation of External Secrets")

	return environmentCreateCmd
}

func bootstrapAgent(name, manifestDir string, skipExternalSecrets bool) error {
	apiClient, err := factory.HttpClient()
	if err != nil {
		return err
	}

	repo, err := factory.BaseRepo()
	if err != nil {
		return err
	}
	project, err := repo.Project(apiClient)
	if err != nil {
		return err
	}

	// TODO: Validation - https://gitlab.com/groups/gitlab-org/-/epics/12594#validations

	// Create an environment to store related secrets
	_, err = api.CreateEnvironment(apiClient, repo.FullName(), &gitlab.CreateEnvironmentOptions{
		Name: &name,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(factory.IO.StdOut, "Created environment: %s\n", name)
	agentManifestDir := fmt.Sprintf("%s/%s", manifestDir, name)

	if !skipExternalSecrets {
		// Create a project access token for the External Secrets controller to retrieve secrets
		pat_name := fmt.Sprintf("external_secrets_pat_%s", name)
		now := time.Now()
		expires_at := gitlab.ISOTime(now.AddDate(0, 0, 90))
		accessLevel := gitlab.MaintainerPermissions
		pat, err := api.CreateProjectAccessToken(apiClient, repo.FullName(), &gitlab.CreateProjectAccessTokenOptions{
			Name:        &pat_name,
			Scopes:      &[]string{"api"},
			AccessLevel: &accessLevel,
			ExpiresAt:   &expires_at,
		})
		if err != nil {
			return err
		}
		_, err = api.CreateProjectVariable(apiClient, repo.FullName(), &gitlab.CreateProjectVariableOptions{
			Key:              &pat.Name,
			Value:            &pat.Token,
			Masked:           gitlab.Bool(true),
			Protected:        gitlab.Bool(true),
			EnvironmentScope: &name,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Created and saved project access token in environment variable: %s\n", pat.Name)

		// Create the external-secrets es_namespace
		es_namespace := "external-secrets"
		es_secret_ns_cmd := exec.Command("kubectl", "create", "namespace", es_namespace)
		es_secret_ns_cmd.Stdout = factory.IO.StdOut
		es_secret_ns_cmd.Stderr = factory.IO.StdErr
		err = es_secret_ns_cmd.Run()
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Created "+es_namespace+" namespace\n")

		// Applies the External Secret token to Kubernetes
		es_gitlab_secret_name := "external-secrets-token"
		es_secret_cmd := exec.Command("kubectl", "create", "secret", "generic", "-n", es_namespace, "--from-literal=token="+pat.Token, es_gitlab_secret_name)
		es_secret_cmd.Stdout = factory.IO.StdOut
		es_secret_cmd.Stderr = factory.IO.StdErr
		err = es_secret_cmd.Run()
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Applied External Secrets token to Kubernetes\n")

		// Generates the YAML to install the External Secrets controller to /manifests/demo-agent/external-secrets.yaml
		es_helm_add_cmd := exec.Command("helm", "repo", "add", "external-secrets", "https://charts.external-secrets.io")
		es_helm_add_cmd.Stdout = factory.IO.StdOut
		es_helm_add_cmd.Stderr = factory.IO.StdErr
		err = es_helm_add_cmd.Run()
		if err != nil {
			return err
		}

		es_template_cmd := exec.Command("helm", "template", "external-secrets", "external-secrets/external-secrets", "-n", es_namespace)
		agentutils.RunCommandToFile(es_template_cmd, agentManifestDir+"/external-secrets.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Generated external secrets manifests to %s\n", agentManifestDir+"/external-secrets.yaml")

		// Applies /manifests/demo-agent/external-secrets.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"/external-secrets.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Applied External Secrets manifests to Kubernetes\n")

		// TODO: Wait for external secrets to be ready

		// Generates the YAML to configure the External Secrets controller to retrieve its own token from GitLab (to allow the rotation of the token) under /manifests/demo-agent/external-secrets-gitlab.yaml
		externalSecretStoreYAML := fmt.Sprintf(`apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: gitlab-secret-store
  namespace: %s
spec:
  provider:
    # provider type: gitlab
    gitlab:
      # url: https://gitlab.mydomain.com/
      auth:
        SecretRef:
          accessToken:
            name: %s
            key: token
      projectID: "%d"
      environment: "%s"`, es_namespace, es_gitlab_secret_name, project.ID, name)
		agentutils.WriteToFile(agentManifestDir+"/external-secrets-gitlab.yaml", externalSecretStoreYAML)
		fmt.Fprintf(factory.IO.StdOut, "Generated external secrets gitlab store manifests to %s\n", agentManifestDir+"/external-secrets-gitlab.yaml")

		// Applies /manifests/demo-agent/external-secrets-gitlab.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"/external-secrets-gitlab.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Applied External Secrets gitlab store manifests to Kubernetes\n")

		// Retrieve the external secrets own token by external secrets - YAML
		externalSecretOwnYAML := fmt.Sprintf(`apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: %s
  namespace: %s
spec:
  refreshInterval: 1h

  secretStoreRef:
    kind: SecretStore
    name: gitlab-secret-store # Must match SecretStore on the cluster

  target:
    name: %s # Name for the secret to be created on the cluster
    creationPolicy: Owner

  data:
    - secretKey: token # Key given to the secret to be created on the cluster
      remoteRef: 
        key: external_secrets_pat_%s # Key of the variable on Gitlab`, es_gitlab_secret_name, es_namespace, es_gitlab_secret_name, name)
		agentutils.WriteToFile(agentManifestDir+"/external-secrets-token.yaml", externalSecretOwnYAML)
		fmt.Fprintf(factory.IO.StdOut, "Generated external secrets manifests to %s\n", agentManifestDir+"/external-secrets-token.yaml")

		// Applies /manifests/demo-agent/external-secrets-token.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"/external-secrets-token.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Configured External Secrets to update its own token\n")
	}

	// Creates a deployment token with read_registry rights for Flux and store it as a masked variable for the environment

	// Applies /manifests/demo-agent/external-secrets-flux.yaml in the cluster

	// Generates the YAML for Flux under /manifests/demo-agent/flux-system/ (equivalent to flux install --export > /manifests/demo-agent/flux-system/gotk-components.yaml

	// Generates the YAML for Flux to watch the OCI image registry.gitlab.com/<project-slug>/demo-agent/flux-manifests:latest

	// Generates the YAML for Flux to watch /manifests/demo-agent/ under /manifests/demo-agent/flux-system/gotk-sync.yaml

	// Builds and pushes the registry.gitlab.com/<project-slug>/demo-agent/flux-manifests:latest image including /manifests/demo-agent

	// Generates a kustomization to use both files in /manifests/demo-agent/flux-system/kustomization.yaml

	// Adds a CI component to /.gitlab-ci.yml to build the OCI image when any file changes under /manifests/demo-agent on the default branch

	// Applies /manifests/demo-agent/flux-system/kustomization.yaml in the cluster (with kubectl -k)

	// Commits and pushes all the above to the Gitlab project (from this point on everything is deployed by Flux through building the OCI image)

	// Registers the demo-agent with GitLab and stores the registration token as an environment variable under the demo-agent environment

	// Generates the YAML to configure the External Secrets controller to retrieve the agent registration token from GitLab to /manifests/demo-agent/external-secrets-agentk.yaml

	// Generates the YAML to deploy the most recent agentk with a Flux HelmRelease

	// Generates an agent config under /.gitlab/agents/demo-agent/config.yaml that enables user access for the current project

	// generate /manifests/demo-agent/auth/project-owner-is-namespace-admin.yaml that sets up the current project's GitLab owners as namespace admins with the necessary RoleBindings

	// Commits and pushes all the above to the Gitlab project

	// Prints the URL to the terminal to access the configured environment and check the status of the cluster

	// Prints instructions to configure local access with glab (glab cluster agent update-kubeconfig --repo '<group>/<project>' --agent '<agent-id>' --use-context)

	fmt.Fprintf(factory.IO.StdOut, "Bootstrap done\n")
	return nil
}
