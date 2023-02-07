package api

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Workspace struct {
	ID          string
	Name        string
	Url         string
	Editor      string
	ActualState string
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