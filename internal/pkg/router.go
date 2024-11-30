package pkg

import (
	resolvers "app/internal/pkg/resolvers"
	"app/internal/pkg/todo"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Init() *resolvers.Resolver {
	return &resolvers.Resolver{
		Todo: todo.NewTodoModule(),
	}
}
