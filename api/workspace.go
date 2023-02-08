package api

import (
	"context"
	"errors"

	"github.com/hasura/go-graphql-client"
)

var ErrWorkspaceNotFound = errors.New("Workspace not found")

type Workspace struct {
	ID          string
	Name        string
	Url         string
	Editor      string
	ActualState string
	Devfile     string
}

type WorkspaceCreateInput struct {
	GroupPath      string `json:"groupPath"`
	Editor         string `json:"editor"`
	ClusterAgentID string `json:"clusterAgentId"`
	Devfile        string `json:"devfile"`
	DesiredState   string `json:"desiredState"`
}

func ListWorkspaces(ctx context.Context, client *graphql.Client, group string) ([]Workspace, error) {
	var query struct {
		Group struct {
			Id         string
			Name       string
			Workspaces struct {
				Nodes []Workspace
			}
		} `graphql:"group(fullPath: $group)"`
	}

	err := client.Query(ctx, &query, map[string]interface{}{
		"group": graphql.ID(group),
	})
	if err != nil {
		return nil, err
	}

	return query.Group.Workspaces.Nodes, nil
}

type RemoteDevelopmentWorkspaceID string

func ViewWorkspace(ctx context.Context, client *graphql.Client, group string, workspaceID string) (*Workspace, error) {
	var query struct {
		Group struct {
			Id         string
			Name       string
			Workspaces struct {
				Nodes []Workspace
			} `graphql:"workspaces(id: $workspaceID)"`
		} `graphql:"group(fullPath: $group)"`
	}

	err := client.Query(ctx, &query, map[string]interface{}{
		"group":       graphql.ID(group),
		"workspaceID": RemoteDevelopmentWorkspaceID(workspaceID),
	})
	if err != nil {
		return nil, err
	}

	if len(query.Group.Workspaces.Nodes) == 0 {
		return nil, ErrWorkspaceNotFound
	}

	return &query.Group.Workspaces.Nodes[0], nil
}

func CreateWorkspace(ctx context.Context, client *graphql.Client, input WorkspaceCreateInput) (*Workspace, error) {
	var mutation struct {
		WorkspaceCreate struct {
			Workspace Workspace
		} `graphql:"workspaceCreate(input: $input)"`
	}

	err := client.Mutate(ctx, &mutation, map[string]interface{}{
		"input": input,
	})
	if err != nil {
		return nil, err
	}

	return &mutation.WorkspaceCreate.Workspace, nil
}
