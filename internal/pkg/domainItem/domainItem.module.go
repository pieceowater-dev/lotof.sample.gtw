package domainItem

import (
	"app/internal/pkg/domainItem/ctrl"
	"app/internal/pkg/domainItem/svc"

	"google.golang.org/grpc"
)

// Module represents the domain item module, including its name, version, and API controller.
type Module struct {
	name    string
	version string
	API     *ctrl.DomainItemController
}

// New creates a new instance of the Module, initializing the service and
// controller against the shared pooled connection to the svc.
func New(conn grpc.ClientConnInterface) Module {
	service := svc.NewDomainItemService(conn)
	controller := ctrl.NewDomainItemController(service)

	return Module{
		name:    "DomainItem",
		version: "v1",
		API:     controller,
	}
}

func (m Module) Initialize() error { return nil }

func (m Module) Version() string { return m.version }

func (m Module) Name() string { return m.name }
