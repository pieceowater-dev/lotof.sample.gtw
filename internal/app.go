package internal

import (
	"app/internal/core/cfg"
	"app/internal/core/graph"
	"app/internal/pkg"
	"context"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	gossiper "github.com/pieceowater-dev/lotof.lib.gossiper/v2"
)

type Application interface {
	Start()
	Stop()
}

type App struct {
	cfg     *cfg.Config
	ctx     context.Context
	servers *gossiper.ServerManager
}

func NewApp() *App {
	return &App{
		// Initialize context
		// This context can be used to manage the lifecycle of the application
		// and pass it to various components as needed
		ctx: context.Background(),
		// Load configuration
		// This configuration can be used to set up the application
		cfg: cfg.Inst(),
		// Initialize server manager
		// This server manager can be used to manage multiple servers
		// and their lifecycle
		// It can also be used to add new servers dynamically
		// and manage their lifecycle
		servers: gossiper.NewServerManager(),
	}
}

func (a *App) Start() {
	// Initialize the application router.
	appRouter := pkg.NewRouter()

	// If this gateway serves as grpc server somehow uncomment below
	// serverManager := gossiper.NewServerManager()
	// serverManager.AddServer(gossiper.NewGRPCServ(appCfg.GrpcPort, grpc.NewServer(), appRouter.InitGRPC))
	// var wg sync.WaitGroup
	// wg.Add(1)
	// // Start gRPC servers in a goroutine
	// go func() {
	//  defer wg.Done()
	//  serverManager.StartAll()
	// }()

	// Initialize resolvers.
	resolvers, err := appRouter.InitializeRouter()
	if err != nil {
		log.Fatalf("Error initializing router: %v", err)
	}

	// Create GraphQL server.
	srv := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: resolvers.(graph.ResolverRoot),
			},
		),
	)

	// Set up the HTTP routes.
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	// Start the HTTP server.
	log.Printf("connect to http://localhost:%s/ for GraphQL playground", a.cfg.AppPort)
	log.Fatal(http.ListenAndServe(":"+a.cfg.AppPort, nil))
	a.servers.StartAll()
	defer a.servers.StopAll()
}

func (a *App) Stop() {
	a.servers.StopAll()
}
