package api

import "github.com/xanzy/go-gitlab"

var CreateEnvironment = func(client *gitlab.Client, projectID interface{}, opts *gitlab.CreateEnvironmentOptions) (*gitlab.Environment, error) {
	if client == nil {
		client = apiClient.Lab()
	}

	environment, _, err := client.Environments.CreateEnvironment(projectID, opts)
	if err != nil {
		return nil, err
	}

	return environment, nil
}

var GetEnvironment = func(client *gitlab.Client, projectID interface{}, environmentID int) (*gitlab.Environment, error) {
	if client == nil {
		client = apiClient.Lab()
	}

	agent, _, err := client.Environments.GetEnvironment(projectID, environmentID)
	if err != nil {
		return nil, err
	}

	return agent, nil
}
