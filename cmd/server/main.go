package main

import (
	"app/internal/core/graph"
	resolvers "app/internal/core/graph/resolvers"
	"app/internal/pkg/todo"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

//todo:
// 1 - abstract interface for transport communication
// 2 - config for env + remote microservices hosts
// 3 - .proto files sharing from microservices and unpacking into ./internal/core/grpc/protos/ somehow

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

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

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
