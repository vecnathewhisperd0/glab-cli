package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coder/websocket"
	"github.com/spf13/cobra"
	"gitlab.com/gitlab-org/cli/commands/cmdutils"
	"net/http"
	"os"
)

const (
	graphAPISubprotocolName = "gitlab-agent-graph-api"
)

func NewCmdGraph(f *cmdutils.Factory) *cobra.Command {
	graphCmd := &cobra.Command{
		Use:   "graph [flags]",
		Short: `Query Kubernetes object graph using GitLab Agent for Kubernetes.`,
		Long:  ``,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID, err := cmd.Flags().GetInt64("agent")
			if err != nil {
				return err
			}

			// TODO figure out how to get kas URL.
			//apiClient, err := f.HttpClient()
			//if err != nil {
			//	return err
			//}
			//url.JoinPath(apiClient.BaseURL().Path,"/")

			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			conn, _, err := websocket.Dial(ctx, "https://gdk.test:8080/-/k8s-proxy/graph", &websocket.DialOptions{
				//HTTPClient:           nil,
				HTTPHeader: http.Header{
					"Authorization": []string{fmt.Sprintf("Bearer pat:%d:%s", agentID, os.Getenv("MY_TOKEN"))},
				},
				//Host:                 "",
				Subprotocols: []string{graphAPISubprotocolName},
			})
			if err != nil {
				return fmt.Errorf("websocket dial: %w", err)
			}
			req, err := json.Marshal(&watchGraphWebSocketRequest{
				Queries: []query{
					{
						Include: &queryInclude{
							ResourceSelectorExpression: "group == '' && version == 'v1' && resource == 'pods'",
							Object:                     nil,
						},
					},
				},
				Namespaces: &namespaces{
					//LabelSelector:            "",
					//FieldSelector:            "",
					ObjectSelectorExpression: "group == '' && version == 'v1' && name == 'ns'",
				},
			})
			if err != nil {
				return err
			}

			err = conn.Write(ctx, websocket.MessageText, req)
			if err != nil {
				return err
			}

			for {
				mt, data, err := conn.Read(ctx)
				if err != nil {
					return err
				}
				if mt != websocket.MessageText {
					return errors.New("unexpected message type") // shouldn't ever happen
				}
				_, err = f.IO.StdOut.Write(data)
				if err != nil {
					return err
				}
			}
		},
	}
	graphCmd.Flags().Int64P("agent", "a", 0, "The numerical Agent ID to connect to.")
	cobra.CheckErr(graphCmd.MarkFlagRequired("agent"))

	return graphCmd
}

type watchGraphWebSocketRequest struct {
	Queries    []query     `json:"queries,omitempty"`
	Namespaces *namespaces `json:"namespaces,omitempty"`
	Roots      *roots      `json:"roots,omitempty"`
}

type query struct {
	Include *queryInclude `json:"include,omitempty"`
	Exclude *queryExclude `json:"exclude,omitempty"`
}

type queryInclude struct {
	ResourceSelectorExpression string              `json:"resource_selector_expression,omitempty"`
	Object                     *queryIncludeObject `json:"object,omitempty"`
}

type queryIncludeObject struct {
	LabelSelector            string `json:"label_selector,omitempty"`
	FieldSelector            string `json:"field_selector,omitempty"`
	ObjectSelectorExpression string `json:"object_selector_expression,omitempty"`
	JsonPath                 string `json:"json_path,omitempty"`
}

type queryExclude struct {
	ResourceSelectorExpression string `json:"resource_selector_expression,omitempty"`
}

type namespaces struct {
	Names                    []string `json:"names,omitempty"`
	LabelSelector            string   `json:"label_selector,omitempty"`
	FieldSelector            string   `json:"field_selector,omitempty"`
	ObjectSelectorExpression string   `json:"object_selector_expression,omitempty"`
}

type roots struct {
	Individual []rootsIndividual `json:"individual,omitempty"`
	Selector   []rootsSelector   `json:"selector,omitempty"`
}

type rootsIndividual struct {
	Group     string `json:"group,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

type rootsSelector struct {
	Group         string `json:"group,omitempty"`
	Resource      string `json:"resource,omitempty"`
	LabelSelector string `json:"label_selector,omitempty"`
	FieldSelector string `json:"field_selector,omitempty"`
}
