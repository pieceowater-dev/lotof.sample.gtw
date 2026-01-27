package internal

import (
	"app/internal/core/cfg"
	"app/internal/core/graph"
	"app/internal/pkg"
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	gossiper "github.com/pieceowater-dev/lotof.lib.gossiper/v2"
	"go.opentelemetry.io/otel/trace"
)

type Application interface {
	Start()
	Stop()
}

type App struct {
	cfg      *cfg.Config
	ctx      context.Context
	servers  *gossiper.ServerManager
	logger   *slog.Logger
	tracer   trace.Tracer
	shutdown func(context.Context) error
}

func NewApp() *App {
	baseCtx := context.Background()
	obsLogger, tracer, shutdown, err := observability.Init(baseCtx, observability.Config{
		ServiceName:  "lotof.hub.gtw",
		Environment:  cfg.Inst().Environment,
		OtlpEndpoint: cfg.Inst().OtlpEndpoint,
		SampleRatio:  cfg.Inst().TraceSampleRatio,
		LogLevel:     parseLevel(cfg.Inst().LogLevel),
	})
	if err != nil {
		fallback := slog.Default()
		fallback.Error("observability init failed", slog.Any("error", err))
		obsLogger = fallback
		tracer = trace.NewNoopTracerProvider().Tracer("noop")
		shutdown = func(context.Context) error { return nil }
	}

	return &App{
		// Initialize context
		// This context can be used to manage the lifecycle of the application
		// and pass it to various components as needed
		ctx: baseCtx,
		// Load configuration
		// This configuration can be used to set up the application
		cfg: cfg.Inst(),
		// Initialize server manager
		// This server manager can be used to manage multiple servers
		// and their lifecycle
		// It can also be used to add new servers dynamically
		// and manage their lifecycle
		servers:  gossiper.NewServerManager(),
		logger:   obsLogger,
		tracer:   tracer,
		shutdown: shutdown,
	}
}

func (a *App) Start() {
	// Initialize the application router.
	appRouter, err := pkg.NewRouter()
	if err != nil {
		a.logger.Error("create router failed", slog.String("error", err.Error()))
		return
	}

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

	// Initialize router in goroutine but wait for it before starting HTTP
	resolversChan := make(chan any, 1)
	go func() {
		resolversInit, err := appRouter.InitializeRouter()
		if err != nil {
			a.logger.Error("initialize router failed", slog.String("error", err.Error()))
			resolversChan <- nil
			return
		}
		resolversChan <- resolversInit
	}()

	// Wait for resolvers before starting HTTP server
	resolvers := <-resolversChan

	// Create GraphQL server.
	srv := handler.NewDefaultServer(
		graph.NewExecutableSchema(
			graph.Config{
				Resolvers: resolvers.(graph.ResolverRoot),
			},
		),
	)

	fiberApp := fiber.New(
		fiber.Config{
			DisableStartupMessage: true,
		},
	)
	fiberApp.Use(observability.FiberMiddleware(a.logger, a.tracer))
	fiberApp.Use(cors.New())
	_ = pkg.NewHttpRouter(fiberApp, resolvers)
	a.servers.AddServer(gossiper.NewRESTServ(a.cfg.AppPort, fiberApp, func(app *fiber.App) {}))

	a.servers.StartAll()
	defer a.servers.StopAll()
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if a.shutdown != nil {
		_ = a.shutdown(ctx)
	}
	a.servers.StopAll()
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
