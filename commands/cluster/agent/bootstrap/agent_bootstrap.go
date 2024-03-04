package list

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"
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
			skipFlux, err := cmd.Flags().GetBool("skip-flux")
			if err != nil {
				return err
			}
			return bootstrapAgent(name, manifestDir, skipExternalSecrets, skipFlux)
		},
	}
	environmentCreateCmd.Flags().StringP("name", "n", "", "Name of the new environment")
	environmentCreateCmd.Flags().StringP("manifest-dir", "", "manifests", "Base directory for manifests")
	environmentCreateCmd.Flags().BoolP("skip-external-secrets", "", false, "Skips creation of External Secrets")
	environmentCreateCmd.Flags().BoolP("skip-flux", "", false, "Skips the Flux setup")

	return environmentCreateCmd
}

type SecretStore struct {
	SecretStoreName string
	Namespace       string
	ProjectID       int
	SecretName      string
}

type ExternalSecretData struct {
	SecretKey          string
	GitLabVariableName string
}

type ExternalSecret struct {
	ExternalSecretName string
	Namespace          string
	SecretStoreName    string
	TargetSecretName   string
	Data               []ExternalSecretData
}

func bootstrapAgent(name, manifestDir string, skipExternalSecrets, skipFlux bool) error {
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
	var snippetsBase = "https://gitlab.com/-/snippets/3682432/raw/main/"
	now := time.Now()
	// Sometimes we need time.Time
	expires_at_time := now.AddDate(0, 0, 90)
	// Sometimes we need gitlab.ISOTime
	expires_at := gitlab.ISOTime(expires_at_time)

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
	bootstrap_namespace := "gitlab-tools"

	// Initialize some templates
	externalSecretYAMLTemplateFile, err := agentutils.DownloadFile(snippetsBase+"/external-secret.gotmpl", "external_secret_gotmpl-", true)
	if err != nil {
		return err
	}
	defer os.Remove(externalSecretYAMLTemplateFile)
	externalSecretYAMLTemplate, err := template.New(path.Base(externalSecretYAMLTemplateFile)).ParseFiles(externalSecretYAMLTemplateFile)
	if err != nil {
		return err
	}

	// Create the bootstrap_namespace
	es_secret_ns_cmd := exec.Command("kubectl", "create", "namespace", bootstrap_namespace)
	es_secret_ns_cmd.Stdout = factory.IO.StdOut
	es_secret_ns_cmd.Stderr = factory.IO.StdErr
	err = es_secret_ns_cmd.Run()
	if err != nil {
		return err
	}

	type DockerRegistryAuth struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Auth     string `json:"auth"`
	}

	type DockerRegistryConfig struct {
		Auths map[string]DockerRegistryAuth `json:"auths"`
	}

	if !skipExternalSecrets {
		// Create a project access token for the External Secrets controller to retrieve secrets
		pat_name := fmt.Sprintf("external_secrets_pat_%s", name)
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

		fmt.Fprintf(factory.IO.StdOut, "Created "+bootstrap_namespace+" namespace\n")

		// Applies the External Secret token to Kubernetes
		es_gitlab_secret_name := "external-secrets-token"
		es_secret_cmd := exec.Command("kubectl", "create", "secret", "generic", "-n", bootstrap_namespace, "--from-literal=token="+pat.Token, es_gitlab_secret_name)
		es_secret_cmd.Stdout = factory.IO.StdOut
		es_secret_cmd.Stderr = factory.IO.StdErr
		err = es_secret_cmd.Run()
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Applied External Secrets token to Kubernetes\n")

		// Generates the YAML to install the External Secrets controller to /manifests/demo-agent/external-secrets.yaml
		externalSecretControllerTemplateFile, err := agentutils.DownloadFile(snippetsBase+"/external-secrets-controller.gotmpl", "external_secrets_controller_gotmpl-", true)
		if err != nil {
			return err
		}
		defer os.Remove(externalSecretControllerTemplateFile)

		externalSecretControllerYAMLTemplate, err := template.New(path.Base(externalSecretControllerTemplateFile)).ParseFiles(externalSecretControllerTemplateFile)
		if err != nil {
			return err
		}
		err = agentutils.WriteTemplateToFile(externalSecretControllerYAMLTemplate, agentManifestDir+"/external-secrets.yaml", SecretStore{
			Namespace: bootstrap_namespace,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Saved external secrets manifests to %s\n", agentManifestDir+"/external-secrets.yaml")

		// Applies /manifests/demo-agent/external-secrets.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"/external-secrets.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Applied External Secrets manifests to Kubernetes\n")

		fmt.Fprintf(factory.IO.StdOut, "Waiting for External Secrets controller to start\n")
		// Wait for external secrets to be ready
		es_apply_wait := exec.Command("kubectl", "wait", "--for=condition=ready", "pod", "-n", bootstrap_namespace, "-l", "app.kubernetes.io/instance=external-secrets", "--timeout=3m")
		es_apply_wait.Stdout = factory.IO.StdOut
		es_apply_wait.Stderr = factory.IO.StdErr
		err = es_apply_wait.Run()
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "External Secrets controller started\n")

		// Generates the YAML to configure the External Secrets controller to retrieve its own token from GitLab (to allow the rotation of the token) under /manifests/demo-agent/external-secrets-gitlab.yaml
		externalSecretStoreYAMLTemplateFile, err := agentutils.DownloadFile(snippetsBase+"/secret-store.gotmpl", "secret_store_gotmpl-", true)
		if err != nil {
			return err
		}
		defer os.Remove(externalSecretStoreYAMLTemplateFile)

		externalSecretStoreYAMLTemplate, err := template.New(path.Base(externalSecretStoreYAMLTemplateFile)).ParseFiles(externalSecretStoreYAMLTemplateFile)
		if err != nil {
			return err
		}
		err = agentutils.WriteTemplateToFile(externalSecretStoreYAMLTemplate, agentManifestDir+"/external-secrets-gitlab.yaml", SecretStore{
			SecretStoreName: "gitlab-secret-store",
			Namespace:       bootstrap_namespace,
			ProjectID:       project.ID,
			SecretName:      es_gitlab_secret_name,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Generated external secrets gitlab store manifests to %s\n", agentManifestDir+"/external-secrets-gitlab.yaml")

		// Applies /manifests/demo-agent/external-secrets-gitlab.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"/external-secrets-gitlab.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Applied External Secrets gitlab store manifests to Kubernetes\n")

		// Retrieve the external secrets own token with external secrets - YAML
		err = agentutils.WriteTemplateToFile(externalSecretYAMLTemplate, agentManifestDir+"/external-secrets-token.yaml", ExternalSecret{
			ExternalSecretName: es_gitlab_secret_name,
			Namespace:          bootstrap_namespace,
			SecretStoreName:    "gitlab-secret-store",
			TargetSecretName:   es_gitlab_secret_name,
			Data: []ExternalSecretData{
				{
					SecretKey:          "token",
					GitLabVariableName: "external_secrets_pat_" + name,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Generated external secrets manifests to %s\n", agentManifestDir+"/external-secrets-token.yaml")

		// Applies /manifests/demo-agent/external-secrets-token.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"/external-secrets-token.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Configured External Secrets to update its own token\n")
	}

	if !skipFlux {

		// Creates a deployment token with read_registry rights for Flux and store it as a masked variable for the environment
		fluxToken, err := api.CreateProjectDeployToken(apiClient, repo.FullName(), &gitlab.CreateProjectDeployTokenOptions{
			Name:      &name,
			ExpiresAt: &expires_at_time,
			Scopes:    &[]string{"read_registry"},
		})
		if err != nil {
			return err
		}
		var fluxVarName = "flux_deploy_token_" + name
		var dockerConfig = DockerRegistryConfig{
			Auths: map[string]DockerRegistryAuth{
				"registry.gitlab.com": {
					Username: fluxToken.Username,
					Password: fluxToken.Token,
					Auth:     base64.StdEncoding.EncodeToString([]byte(fluxToken.Username + ":" + fluxToken.Token)),
				},
			},
		}
		dockerConfigJSON, err := json.Marshal(dockerConfig)
		if err != nil {
			return err
		}
		var dockerConfigJSONString = string(dockerConfigJSON)
		_, err = api.CreateProjectVariable(apiClient, repo.FullName(), &gitlab.CreateProjectVariableOptions{
			Key:              &fluxVarName,
			Value:            &dockerConfigJSONString,
			Masked:           gitlab.Bool(true),
			Protected:        gitlab.Bool(true),
			EnvironmentScope: &name,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Created and saved deploy token for Flux in environment variable: %s\n", fluxVarName)

		// Retrieve the flux token with external secrets - YAML
		externalSecretDockerConfigYAMLTemplateFile, err := agentutils.DownloadFile(snippetsBase+"/external-secret-dockerconfig", "external_secret_gotmpl-", true)
		if err != nil {
			return err
		}
		defer os.Remove(externalSecretDockerConfigYAMLTemplateFile)
		externalSecretDockerConfigYAMLTemplate, err := template.New(path.Base(externalSecretDockerConfigYAMLTemplateFile)).ParseFiles(externalSecretDockerConfigYAMLTemplateFile)
		if err != nil {
			return err
		}
		err = agentutils.WriteTemplateToFile(externalSecretDockerConfigYAMLTemplate, agentManifestDir+"/flux-token.yaml", ExternalSecret{
			ExternalSecretName: "flux-deploy-token",
			Namespace:          bootstrap_namespace,
			SecretStoreName:    "gitlab-secret-store",
			TargetSecretName:   "flux-gitlab-access-token",
			Data: []ExternalSecretData{
				{
					GitLabVariableName: fluxVarName,
				},
			},
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(factory.IO.StdOut, "Generated external secrets manifests to %s\n", agentManifestDir+"/flux-token.yaml")

		// Applies /manifests/demo-agent/flux-token.yaml in the cluster
		agentutils.KubectlApply(factory.IO, agentManifestDir+"flux-token.yaml")
		fmt.Fprintf(factory.IO.StdOut, "Configured External Secrets to retrieve the flux token\n")

		// Generates the YAML for Flux under /manifests/demo-agent/flux-system/ (equivalent to flux install --export > /manifests/demo-agent/flux-system/gotk-components.yaml

		// Generates the YAML for Flux to watch the OCI image registry.gitlab.com/<project-slug>/demo-agent/flux-manifests:latest

		// Generates a kustomization to use both files in /manifests/demo-agent/flux-system/kustomization.yaml

		// Builds and pushes /manifests/demo-agent to registry.gitlab.com/<project-slug>/demo-agent/flux-manifests:latest OCI image

		// TODO: Adds a CI component to /.gitlab-ci.yml to build the OCI image when any file changes under /manifests/demo-agent on the default branch

		// Applies /manifests/demo-agent/flux-system/kustomization.yaml in the cluster (with kubectl -k) to start Flux

		// TODO: Ask the user to continue with commit and push

		// Commits and pushes all the above to the Gitlab project (from this point on everything is deployed by Flux through building the OCI image)

	}
	// Registers the demo-agent with GitLab and stores the registration token as an environment variable under the demo-agent environment

	// Generates the YAML to configure the External Secrets controller to retrieve the agent registration token from GitLab to /manifests/demo-agent/external-secrets-agentk.yaml

	// Generates the YAML to deploy the most recent agentk with a Flux HelmRelease

	// Generates an agent config under /.gitlab/agents/demo-agent/config.yaml that enables user access for the current project

	// Generate /manifests/demo-agent/auth/project-owner-is-namespace-admin.yaml that sets up the current project's GitLab owners as namespace admins with the necessary RoleBindings

	// TODO: Ask the user to continue with commit and push

	// Commits and pushes all the above to the Gitlab project

	// Prints the URL to the terminal to access the configured environment and check the status of the cluster

	// Prints instructions to configure local access with glab (glab cluster agent update-kubeconfig --repo '<group>/<project>' --agent '<agent-id>' --use-context)

	fmt.Fprintf(factory.IO.StdOut, "Bootstrap done\n")
	return nil
}
