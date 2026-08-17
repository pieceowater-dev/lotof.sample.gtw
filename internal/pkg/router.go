package pkg

import (
	"reflect"

	"google.golang.org/grpc"

	"app/internal/core/generic/interfaces"
	"app/internal/core/grpc/generated/generic/tenants"
	resolvers "app/internal/pkg/_resolvers"
	"app/internal/pkg/auth"
	"app/internal/pkg/domainItem"
	"app/internal/pkg/gateway"
)

// Router manages the modules and initializes the routes for the application.
type Router struct {
	modules map[string]interfaces.IModule
	conn    grpc.ClientConnInterface
}

// NewRouter creates a new Router instance and initializes every module.
// conn is the shared pooled connection to this domain's svc (see
// grpcpool.NewPooledClient in app.go) -- every module's client is
// constructed from it the same way.
func NewRouter(conn grpc.ClientConnInterface) *Router {
	gatewayModule := gateway.New()
	authModule := auth.New()
	domainItemModule := domainItem.New(conn)

	return &Router{
		conn: conn,
		modules: map[string]interfaces.IModule{
			gatewayModule.Name():    gatewayModule,
			authModule.Name():       authModule,
			domainItemModule.Name(): domainItemModule,
		},
	}
}

// InitializeRouter initializes the router and its gRPC/GraphQL routes.
func (r *Router) InitializeRouter() (any, error) {
	resolver := r.initializeGQLResolvers()
	return resolver, nil
}

// InitializeGRPCRoutes registers the inbound gRPC services this gateway
// exposes to the rest of the platform -- tenants.GatewayService, which Hub
// calls to provision a new namespace's schema in this domain's svc.
func (r *Router) InitializeGRPCRoutes(server *grpc.Server) {
	if m, ok := r.modules["Gateway"]; ok {
		tenants.RegisterGatewayServiceServer(server, m.(gateway.Module).API)
	}
}

// initializeGQLResolvers initializes the GraphQL resolvers for all modules.
func (r *Router) initializeGQLResolvers() *resolvers.Resolver {
	resolver := &resolvers.Resolver{}
	resolverValue := reflect.ValueOf(resolver).Elem()

	for i := 0; i < resolverValue.NumField(); i++ {
		field := resolverValue.Type().Field(i)
		moduleName := field.Tag.Get("module")
		if moduleName == "" {
			moduleName = field.Name
		}
		if module, ok := r.modules[moduleName]; ok {
			resolverValue.Field(i).Set(reflect.ValueOf(module))
		}
	}

	return resolver
}

// GetModules returns the map of modules.
func (r *Router) GetModules() map[string]interfaces.IModule {
	return r.modules
}
