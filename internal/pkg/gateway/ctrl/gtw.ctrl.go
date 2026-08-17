package ctrl

import (
	"context"

	"app/internal/core/grpc/generated/generic/tenants"
	"app/internal/pkg/gateway/svc"
)

// GtwController implements the generic tenants.GatewayService gRPC server
// that Hub calls into to provision a new namespace's schema across every
// backing microservice behind this gateway.
type GtwController struct {
	service *svc.GtwService
	tenants.UnimplementedGatewayServiceServer
}

func NewGtwController(service *svc.GtwService) *GtwController {
	return &GtwController{service: service}
}

func (c *GtwController) AddNamespaceTenant(ctx context.Context, req *tenants.AddNamespaceTenantRequest) (*tenants.AddNamespaceTenantResponse, error) {
	return c.service.AddNamespaceTenant(ctx, req)
}

func (c *GtwController) GetNamespaceHealth(ctx context.Context, req *tenants.GetTenantStatusRequest) (*tenants.GetNamespaceHealthResponse, error) {
	return c.service.GetNamespaceHealth(ctx, req)
}
