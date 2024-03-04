package api

import "github.com/xanzy/go-gitlab"

var CreateProjectDeployToken = func(client *gitlab.Client, projectID interface{}, opts *gitlab.CreateProjectDeployTokenOptions) (*gitlab.DeployToken, error) {
	if client == nil {
		client = apiClient.Lab()
	}

	environment, _, err := client.DeployTokens.CreateProjectDeployToken(projectID, opts)
	if err != nil {
		return nil, err
	}

	return environment, nil
}
