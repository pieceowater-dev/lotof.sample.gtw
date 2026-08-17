package svc

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"app/internal/core/grpc/generated/generic/tenants"

	gossiper "github.com/pieceowater-dev/lotof.lib.gossiper/v2"
)

// GtwService fans out tenant provisioning requests from Hub to every backing
// microservice behind this gateway (AppTenantsService.NewTenant). Today
// that's just this domain's own svc, but the client list is built to hold
// more than one if a domain ever splits into several backing services.
type GtwService struct {
	clients []tenants.AppTenantsServiceClient
	mu      sync.Mutex
}

func NewGtwService(serverAddresses []string) *GtwService {
	factory := gossiper.NewTransportFactory()
	clients := make([]tenants.AppTenantsServiceClient, 0, len(serverAddresses))

	for _, address := range serverAddresses {
		grpcTransport := factory.CreateTransport(gossiper.GRPC, address)

		client, err := grpcTransport.CreateClient(tenants.NewAppTenantsServiceClient)
		if err != nil {
			log.Printf("Error creating tenants client for %s: %v", address, err)
			continue
		}

		clients = append(clients, client.(tenants.AppTenantsServiceClient))
	}

	return &GtwService{clients: clients}
}

func (s *GtwService) AddNamespaceTenant(ctx context.Context, req *tenants.AddNamespaceTenantRequest) (*tenants.AddNamespaceTenantResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var wg sync.WaitGroup
	results := make([]bool, len(s.clients))
	errors := make([]error, len(s.clients))

	for i, client := range s.clients {
		wg.Add(1)
		go func(i int, client tenants.AppTenantsServiceClient) {
			defer wg.Done()
			resp, err := client.NewTenant(ctx, &tenants.AddNamespaceTenantRequest{
				Namespace:   req.Namespace,
				Credentials: req.Credentials,
			})
			if err != nil {
				errors[i] = err
				log.Printf("Error calling NewTenant on client %d: %v", i, err)
				return
			}
			results[i] = resp.Success
		}(i, client)
	}

	wg.Wait()

	for i, success := range results {
		if !success {
			return nil, errors[i]
		}
	}

	return &tenants.AddNamespaceTenantResponse{Success: true}, nil
}

// GetNamespaceHealth fans out to every configured backend client and
// reports whether the given tenant's schema is reachable and migrated.
// Unlike AddNamespaceTenant above, a single client's failure must never
// blank out the whole response: this exists specifically to surface "this
// product is broken for this tenant," so it always returns a populated
// result instead of erroring out on the first bad client.
func (s *GtwService) GetNamespaceHealth(ctx context.Context, req *tenants.GetTenantStatusRequest) (*tenants.GetNamespaceHealthResponse, error) {
	s.mu.Lock()
	clients := make([]tenants.AppTenantsServiceClient, len(s.clients))
	copy(clients, s.clients)
	s.mu.Unlock()

	if len(clients) == 0 {
		return &tenants.GetNamespaceHealthResponse{Reachable: false, Error: "no backend clients configured"}, nil
	}

	type callResult struct {
		resp *tenants.GetTenantStatusResponse
		err  error
	}
	results := make([]callResult, len(clients))

	var wg sync.WaitGroup
	for i, client := range clients {
		wg.Add(1)
		go func(i int, client tenants.AppTenantsServiceClient) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("recovered from panic calling GetTenantStatus on client %d: %v", i, r)
					results[i].err = fmt.Errorf("panic: %v", r)
				}
			}()
			callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			resp, err := client.GetTenantStatus(callCtx, &tenants.GetTenantStatusRequest{Namespace: req.Namespace})
			if err != nil {
				log.Printf("Error calling GetTenantStatus on client %d: %v", i, err)
			}
			results[i] = callResult{resp: resp, err: err}
		}(i, client)
	}
	wg.Wait()

	reachable := true
	schemaReady := true
	var appliedVersion, targetVersion, errMsg string
	for _, r := range results {
		if r.err != nil {
			reachable = false
			schemaReady = false
			if errMsg == "" {
				errMsg = r.err.Error()
			}
			continue
		}
		if !r.resp.GetSchemaReady() {
			schemaReady = false
		}
		if appliedVersion == "" {
			appliedVersion = r.resp.GetAppliedVersion()
		}
		if targetVersion == "" {
			targetVersion = r.resp.GetTargetVersion()
		}
	}

	return &tenants.GetNamespaceHealthResponse{
		Reachable:      reachable,
		SchemaReady:    schemaReady,
		AppliedVersion: appliedVersion,
		TargetVersion:  targetVersion,
		Error:          errMsg,
	}, nil
}
