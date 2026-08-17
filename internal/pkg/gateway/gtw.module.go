package gateway

import (
	"app/internal/core/cfg"
	"app/internal/pkg/gateway/ctrl"
	"app/internal/pkg/gateway/svc"
)

// Module is the inbound tenants.GatewayService implementation (Hub → this
// service, for tenant provisioning — see generic/tenants/tenants.proto).
// Add more backend addresses to the NewGtwService call below if this domain
// ever splits into several backing microservices.
type Module struct {
	name    string
	version string
	API     *ctrl.GtwController
}

func New() Module {
	service := svc.NewGtwService([]string{
		cfg.Inst().LotofSampleSvcGrpcAddress,
	})
	controller := ctrl.NewGtwController(service)

	return Module{
		name:    "Gateway",
		version: "v1",
		API:     controller,
	}
}

func (m Module) Initialize() error { return nil }

func (m Module) Version() string { return m.version }

func (m Module) Name() string { return m.name }
