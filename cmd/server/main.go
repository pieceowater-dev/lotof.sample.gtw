package main

import (
	"app/internal/core/cfg"
	"app/internal/core/graph"
	"app/internal/pkg"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"log"
	"net/http"
)

func main() {
	appCfg := cfg.Inst()
	appRouter := pkg.NewRouter()

	// if this gateway serves as grpc server somehow uncomment below
	//serverManager := gossiper.NewServerManager()
	//serverManager.AddServer(gossiper.NewGRPCServ(appCfg.GrpcPort, grpc.NewServer(), appRouter.InitGRPC))
	//var wg sync.WaitGroup
	//wg.Add(1)
	//// Start gRPC servers in a goroutine
	//go func() {
	//	defer wg.Done()
	//	serverManager.StartAll()
	//}()

	// Initialize resolvers
	resolvers := appRouter.Init()

	// Create GraphQL server
	srv := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: resolvers,
			},
		),
	)

	// Set up the HTTP routes
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", cfg.Inst().AppPort)
	log.Fatal(http.ListenAndServe(":"+appCfg.AppPort, nil))
}
