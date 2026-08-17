package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NamespaceClientInterceptor injects the tenant namespace into the outgoing
// gRPC metadata so the downstream service can switch to the correct tenant
// schema.
func NamespaceClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if ns := namespaceFromContext(ctx); ns != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "namespace", ns)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NamespaceFromContext resolves the tenant namespace the same way the
// outbound gRPC interceptor does.
func NamespaceFromContext(ctx context.Context) string {
	return namespaceFromContext(ctx)
}

// UserIDClientInterceptor forwards the calling user's id (set into context by
// the @sampleAuth directive, see UserIDContextKey) as outgoing gRPC metadata.
// Without this, the downstream service's TenantReadiness middleware never
// sees a user id and can't fall back to EnsureAppTenant for a namespace's
// very first request -- it would silently stay stuck on "tenant migration in
// progress" forever instead of self-healing.
func UserIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if userID, ok := ctx.Value(UserIDContextKey).(string); ok && userID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "user-id", userID)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func namespaceFromContext(ctx context.Context) (ns string) {
	if v, ok := ctx.Value(NamespaceContextKey).(string); ok && v != "" {
		return v
	}
	// GetOperationContext panics when called outside a GraphQL request
	// lifecycle -- recover and fall through to "" instead.
	defer func() {
		if recover() != nil {
			ns = ""
		}
	}()
	if opCtx := graphql.GetOperationContext(ctx); opCtx != nil {
		if v := opCtx.Headers.Get("Namespace"); v != "" {
			return v
		}
		if v := opCtx.Headers.Get("namespace-slug"); v != "" {
			return v
		}
	}
	return ""
}
