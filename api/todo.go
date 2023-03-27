package api

import (
	"github.com/xanzy/go-gitlab"
)

var ListTodos = func(client *gitlab.Client, opts *gitlab.ListTodosOptions, all bool) ([]*gitlab.Todo, error) {
	var todos []*gitlab.Todo
	var err error
	if client == nil {
		client = apiClient.Lab()
	}
	if opts.PerPage == 0 {
		opts.PerPage = DefaultListLimit
	}
	if all {
		todos_part, resp, _ := client.Todos.ListTodos(opts)
		todos = append(todos, todos_part...)
		var myopt gitlab.ListTodosOptions = *opts
		for resp.NextPage != 0 {
			myopt.Page = resp.NextPage
			todos_part, resp, err = client.Todos.ListTodos(&myopt)
			todos = append(todos, todos_part...)
		}
	} else {
		todos, _, err = client.Todos.ListTodos(opts)
	}
	if err != nil {
		return nil, err
	}

	return todos, nil
}
