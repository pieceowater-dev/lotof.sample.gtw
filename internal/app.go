package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gofiber/fiber/v2"
	recovermw "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gorilla/websocket"
	gossiper "github.com/pieceowater-dev/lotof.lib.gossiper/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"app/internal/core/cfg"
	"app/internal/core/generic/middleware"
	"app/internal/core/generic/observability"
	"app/internal/core/graph"
	"app/internal/core/grpcpool"
	"app/internal/pkg"
	resolverspkg "app/internal/pkg/_resolvers"
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
		ServiceName:  "lotof.sample.gtw",
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

	gossiper.RegisterTransportContextMiddleware(observability.WithOutgoingMetadata)

	return &App{
		ctx:      baseCtx,
		cfg:      cfg.Inst(),
		servers:  gossiper.NewServerManager(),
		logger:   obsLogger,
		tracer:   tracer,
		shutdown: shutdown,
	}
}

func (a *App) Start() {
	// Inbound gRPC server: exposes tenants.GatewayService so Hub can provision
	// a new namespace's schema in this domain's svc (see internal/pkg/gateway).
	grpcServ := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			gossiper.RecoveryUnaryServerInterceptor(),
			observability.GRPCServerInterceptor(a.logger, a.tracer),
		),
	)
	grpcHealth := health.NewServer()
	healthgrpc.RegisterHealthServer(grpcServ, grpcHealth)
	grpcHealth.SetServingStatus("", healthgrpc.HealthCheckResponse_NOT_SERVING)
	reflection.Register(grpcServ)

	// Outbound gRPC client: talks to this domain's svc for all domain APIs.
	// Pooled across several independent connections rather than one -- see
	// grpcpool.NewPooledClient for why.
	conn, err := grpcpool.NewPooledClient(
		a.cfg.LotofSampleSvcGrpcAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			observability.GRPCClientInterceptor(a.logger, a.tracer),
			middleware.NamespaceClientInterceptor(),
			middleware.UserIDClientInterceptor(),
		),
	)
	if err != nil {
		a.logger.Error("failed to connect to svc", slog.String("address", a.cfg.LotofSampleSvcGrpcAddress), slog.String("error", err.Error()))
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			a.logger.Warn("failed to close svc connection", slog.String("error", closeErr.Error()))
		}
	}()

	appRouter := pkg.NewRouter(conn)
	a.servers.AddServer(gossiper.NewGRPCServ(a.cfg.GrpcPort, grpcServ, appRouter.InitializeGRPCRoutes))

	resolversChan := make(chan any, 1)
	go func() {
		resolversInit, resolveErr := appRouter.InitializeRouter()
		if resolveErr != nil {
			a.logger.Error("initialize router failed", slog.String("error", resolveErr.Error()))
			resolversChan <- nil
			return
		}
		resolversChan <- resolversInit
	}()

	resolvers := <-resolversChan

	authValidate := resolvers.(*resolverspkg.Resolver).AuthModule.API.ValidateToken

	srv := handler.New(
		graph.NewExecutableSchema(graph.Config{
			Resolvers: resolvers.(graph.ResolverRoot),
			Directives: graph.DirectiveRoot{
				HubAuth: middleware.HubAuthDirective,
				SampleAuth: func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
					return middleware.SampleAuthDirective(ctx, next, authValidate)
				},
			},
		}),
	)
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 20 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				parsedOrigin, err := url.Parse(origin)
				if err != nil || parsedOrigin.Hostname() == "" {
					return false
				}
				hostToCompare := r.Host
				if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); forwarded != "" {
					hostToCompare = forwarded
				}
				parsedHost, err := url.Parse("//" + hostToCompare)
				if err != nil || parsedHost.Hostname() == "" {
					return false
				}
				return parsedOrigin.Hostname() == parsedHost.Hostname()
			},
		},
		// graphql-ws sends auth via the connection_init payload, not real
		// HTTP headers, so SampleAuthDirective (which normally reads
		// SampleAuthorization/Authorization off opCtx.Headers) has nothing
		// to read for a subscription op. Stash the raw token here instead --
		// the directive falls back to middleware.WSTokenContextKey and
		// validates it through the exact same code path as any other
		// request.
		InitFunc: func(ctx context.Context, payload transport.InitPayload) (context.Context, *transport.InitPayload, error) {
			raw := ""
			if v, ok := payload["SampleAuthorization"]; ok {
				raw = fmt.Sprint(v)
			}
			if raw == "" {
				if v, ok := payload["Authorization"]; ok {
					raw = fmt.Sprint(v)
				}
			}
			if raw != "" {
				ctx = context.WithValue(ctx, middleware.WSTokenContextKey, raw)
			}
			return ctx, &payload, nil
		},
	})

	fiberApp := fiber.New(fiber.Config{DisableStartupMessage: true})
	// An unrecovered panic in any handler otherwise takes down the whole
	// gateway process for every tenant.
	fiberApp.Use(recovermw.New())
	fiberApp.Use(func(c *fiber.Ctx) error {
		middleware.NormalizeAPIPathPrefix(c)
		return c.Next()
	})
	fiberApp.Use(func(c *fiber.Ctx) error {
		middleware.FiberCORSMiddleware(c)
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.Next()
	})
	fiberApp.Use(observability.FiberMiddleware(a.logger, a.tracer))

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok", "service": "lotof.sample.gtw"})
	})
	fiberApp.Get("/health/live", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok", "service": "lotof.sample.gtw", "check": "liveness"})
	})
	fiberApp.Get("/health/ready", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok", "service": "lotof.sample.gtw", "check": "readiness"})
	})

	fiberApp.Post("/query", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srv.ServeHTTP(w, r)
		}))(c.Context())
		return nil
	})

	fiberApp.Get("/query", func(c *fiber.Ctx) error {
		fasthttpadaptor.NewFastHTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			srv.ServeHTTP(w, r)
		}))(c.Context())
		return nil
	})

	a.servers.AddServer(gossiper.NewRESTServ(a.cfg.AppPort, fiberApp, func(app *fiber.App) {}))
	grpcHealth.SetServingStatus("", healthgrpc.HealthCheckResponse_SERVING)

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
