package api

import "github.com/xanzy/go-gitlab"

var ListTodos = func(client *gitlab.Client, opts *gitlab.ListTodosOptions) ([]*gitlab.Todo, *gitlab.Response, error) {
	if client == nil {
		client = apiClient.Lab()
	}

	if opts.PerPage == 0 {
		opts.PerPage = DefaultListLimit
	}

	todos, resp, err := client.Todos.ListTodos(opts)
	if err != nil {
		return nil, nil, err
	}
	return todos, resp, nil
}
