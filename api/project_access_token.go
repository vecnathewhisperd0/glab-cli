package api

import "github.com/xanzy/go-gitlab"

var CreateProjectAccessToken = func(client *gitlab.Client, projectID interface{}, opts *gitlab.CreateProjectAccessTokenOptions) (*gitlab.ProjectAccessToken, error) {
	if client == nil {
		client = apiClient.Lab()
	}

	pat, _, err := client.ProjectAccessTokens.CreateProjectAccessToken(projectID, opts)
	if err != nil {
		return nil, err
	}

	return pat, nil
}
