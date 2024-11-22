package main

import (
	"app/internal/core/cfg"
	"app/internal/core/graph"
	resolvers "app/internal/core/graph/resolvers"
	"app/internal/pkg/todo"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
)

func main() {

	srv := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: &resolvers.Resolver{
					Todo: todo.NewTodoModule(),
				},
			},
		),
	)

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Inst().AppPort)
	log.Fatal(http.ListenAndServe(":"+cfg.Inst().AppPort, nil))
}
