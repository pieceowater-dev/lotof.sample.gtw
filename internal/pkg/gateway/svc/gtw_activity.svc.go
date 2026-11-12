package svc

import (
	"context"
	"time"

	"app/internal/core/grpc/generated/generic/tenants"
)

// GetNamespaceActivity forwards the console's usage question to the backing
// core service (today always one client). Like GetNamespaceHealth it never
// errors out: a failed call comes back as Error so the console still shows
// the other products.
func (s *GtwService) GetNamespaceActivity(ctx context.Context, req *tenants.GetTenantActivityRequest) (*tenants.GetTenantActivityResponse, error) {
	s.mu.Lock()
	clients := make([]tenants.AppTenantsServiceClient, len(s.clients))
	copy(clients, s.clients)
	s.mu.Unlock()

	if len(clients) == 0 {
		return &tenants.GetTenantActivityResponse{Error: "no backend clients configured"}, nil
	}

	callCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	resp, err := clients[0].GetTenantActivity(callCtx, req)
	if err != nil {
		return &tenants.GetTenantActivityResponse{Error: err.Error()}, nil
	}
	return resp, nil
}
