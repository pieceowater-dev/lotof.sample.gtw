package observability

import (
    "context"
    "log/slog"
    "os"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

// Config controls how observability is initialised.
type Config struct {
    ServiceName  string
    Environment  string
    OtlpEndpoint string
    SampleRatio  float64
    LogLevel     slog.Level
}

// ctxKey is a private type to avoid collisions.
type ctxKey string

const (
    ctxKeyRequestID ctxKey = "request-id"
    ctxKeyTenant    ctxKey = "tenant"
    ctxKeyUser      ctxKey = "user"
)

// Init sets up OTEL tracer provider and JSON logger.
func Init(ctx context.Context, cfg Config) (*slog.Logger, trace.Tracer, func(context.Context) error, error) {
    res, err := resource.Merge(
        resource.Default(),
        resource.NewWithAttributes(
            "http://opentelemetry.io/schema/1.28.0",
            attribute.String("service.name", cfg.ServiceName),
            attribute.String("environment", cfg.Environment),
        ),
    )
    if err != nil {
        return nil, nil, nil, err
    }

    exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(cfg.OtlpEndpoint), otlptracehttp.WithInsecure())
    if err != nil {
        return nil, nil, nil, err
    }

    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exporter),
        tracesdk.WithSampler(tracesdk.TraceIDRatioBased(cfg.SampleRatio)),
        tracesdk.WithResource(res),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel, AddSource: true})).
        With("service", cfg.ServiceName, "environment", cfg.Environment)

    shutdown := func(ctx context.Context) error { return tp.Shutdown(ctx) }
    return logger, tp.Tracer(cfg.ServiceName), shutdown, nil
}

// RequestID extracts request ID from context if set.
func RequestID(ctx context.Context) string {
    if v := ctx.Value(ctxKeyRequestID); v != nil {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return ""
}

// WithRequestData adds request-scoped fields to context.
func WithRequestData(ctx context.Context, requestID, tenant, user string) context.Context {
    ctx = context.WithValue(ctx, ctxKeyRequestID, requestID)
    if tenant != "" {
        ctx = context.WithValue(ctx, ctxKeyTenant, tenant)
    }
    if user != "" {
        ctx = context.WithValue(ctx, ctxKeyUser, user)
    }
    return ctx
}

// LoggerFromContext enriches base logger with request attributes when present.
func LoggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
    if base == nil {
        return slog.Default()
    }
    attrs := []any{}
    if id := RequestID(ctx); id != "" {
        attrs = append(attrs, "request_id", id)
    }
    if span := trace.SpanContextFromContext(ctx); span.IsValid() {
        attrs = append(attrs, "trace_id", span.TraceID().String(), "span_id", span.SpanID().String())
    }
    if tenant, _ := ctx.Value(ctxKeyTenant).(string); tenant != "" {
        attrs = append(attrs, "tenant", tenant)
    }
    if user, _ := ctx.Value(ctxKeyUser).(string); user != "" {
        attrs = append(attrs, "user", user)
    }
    if len(attrs) == 0 {
        return base
    }
    return base.With(attrs...)
}

// WithOutgoingMetadata attaches request-id and trace headers to outgoing context.
func WithOutgoingMetadata(ctx context.Context) context.Context {
    reqID := RequestID(ctx)
    if reqID == "" {
        reqID = uuid.NewString()
        ctx = WithRequestData(ctx, reqID, "", "")
    }

    md, _ := metadata.FromOutgoingContext(ctx)
    md = md.Copy()
    md.Set("x-request-id", reqID)

    carrier := metadataCarrier(md)
    otel.GetTextMapPropagator().Inject(ctx, carrier)

    return metadata.NewOutgoingContext(ctx, md)
}

// FiberMiddleware instruments incoming HTTP requests and sets context for downstream calls.
func FiberMiddleware(logger *slog.Logger, tracer trace.Tracer) func(c *fiber.Ctx) error {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        ctx := c.UserContext()
        if ctx == nil {
            ctx = context.Background()
        }

        headers := c.GetReqHeaders()
        carrier := make(mapCarrier, len(headers))
        for k, values := range headers {
            if len(values) > 0 {
                carrier[strings.ToLower(k)] = values[0]
            }
        }

        ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
        reqID := firstNonEmpty(
            carrier.Get("x-request-id"),
            carrier.Get("x-correlation-id"),
        )
        if reqID == "" {
            reqID = uuid.NewString()
        }

        ctx = WithRequestData(ctx, reqID, c.Get("Namespace"), c.Get("Authorization"))
        ctx = WithOutgoingMetadata(ctx)

        ctx, span := tracer.Start(ctx, c.Method()+" "+c.Route().Path, trace.WithSpanKind(trace.SpanKindServer))
        defer span.End()

        c.SetUserContext(ctx)
        c.Set("X-Request-ID", reqID)

        err := c.Next()

        loggerWithCtx := LoggerFromContext(ctx, logger)
        status := c.Response().StatusCode()

        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
            loggerWithCtx.Error("http request failed",
                slog.String("method", c.Method()),
                slog.String("route", c.Route().Path),
                slog.Int("status", status),
                slog.Duration("latency_ms", time.Since(start)),
                slog.String("remote_ip", c.IP()),
                slog.String("error", err.Error()),
            )
            return err
        }

        span.SetAttributes(
            attribute.String("http.method", c.Method()),
            attribute.String("http.route", c.Route().Path),
            attribute.Int("http.status_code", status),
        )
        loggerWithCtx.Info("http request completed",
            slog.String("method", c.Method()),
            slog.String("route", c.Route().Path),
            slog.Int("status", status),
            slog.Duration("latency_ms", time.Since(start)),
            slog.String("remote_ip", c.IP()),
        )
        return nil
    }
}

// GRPCServerInterceptor adds tracing, logging, and metadata propagation to gRPC servers.
func GRPCServerInterceptor(logger *slog.Logger, tracer trace.Tracer) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
        md, _ := metadata.FromIncomingContext(ctx)
        ctx = otel.GetTextMapPropagator().Extract(ctx, metadataCarrier(md))

        reqID := firstFromMD(md, "x-request-id", "x-correlation-id")
        if reqID == "" {
            reqID = uuid.NewString()
        }

        ctx = WithRequestData(ctx, reqID, "", "")
        ctx = WithOutgoingMetadata(ctx)

        ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
        defer span.End()

        resp, err := handler(ctx, req)

        loggerWithCtx := LoggerFromContext(ctx, logger)
        grpcCode := status.Code(err)

        if err != nil {
            span.RecordError(err)
            span.SetStatus(codes.Error, err.Error())
            loggerWithCtx.Error("grpc request failed",
                slog.String("method", info.FullMethod),
                slog.String("grpc_code", grpcCode.String()),
                slog.String("error", err.Error()),
            )
            return resp, err
        }

        loggerWithCtx.Info("grpc request completed",
            slog.String("method", info.FullMethod),
            slog.String("grpc_code", grpcCode.String()),
        )
        return resp, nil
    }
}

// GRPCClientInterceptor propagates request context and records client spans.
func GRPCClientInterceptor(logger *slog.Logger, tracer trace.Tracer) grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        ctx = WithOutgoingMetadata(ctx)
        ctx, span := tracer.Start(ctx, method, trace.WithSpanKind(trace.SpanKindClient))
        defer span.End()

        err := invoker(ctx, method, req, reply, cc, opts...)
        if err != nil {
            span.RecordError(err)
            LoggerFromContext(ctx, logger).Error("grpc client failed", slog.String("method", method), slog.String("error", err.Error()))
            return err
        }

        LoggerFromContext(ctx, logger).Info("grpc client completed", slog.String("method", method))
        return nil
    }
}

// mapCarrier adapts a plain map into a propagation carrier.
type mapCarrier map[string]string

func (c mapCarrier) Get(key string) string { return c[strings.ToLower(key)] }
func (c mapCarrier) Set(key, value string) { c[strings.ToLower(key)] = value }
func (c mapCarrier) Keys() []string {
    keys := make([]string, 0, len(c))
    for k := range c {
        keys = append(keys, k)
    }
    return keys
}

// metadataCarrier adapts gRPC metadata for propagation.
type metadataCarrier metadata.MD

func (m metadataCarrier) Get(key string) string {
    values := metadata.MD(m).Get(key)
    if len(values) == 0 {
        return ""
    }
    return values[0]
}

func (m metadataCarrier) Set(key, value string) {
    md := metadata.MD(m)
    md.Set(key, value)
}

func (m metadataCarrier) Keys() []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    return keys
}

func firstNonEmpty(values ...string) string {
    for _, v := range values {
        if strings.TrimSpace(v) != "" {
            return v
        }
    }
    return ""
}

func firstFromMD(md metadata.MD, keys ...string) string {
    for _, k := range keys {
        if v := md.Get(k); len(v) > 0 && v[0] != "" {
            return v[0]
        }
    }
    return ""
}
