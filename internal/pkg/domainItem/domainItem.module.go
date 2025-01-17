package domainItem

import (
	"app/internal/pkg/domainItem/ctrl"
	"app/internal/pkg/domainItem/svc"
)

// Module represents the domain item module, including its name, version, and API controller.
type Module struct {
	name    string
	version string
	API     *ctrl.DomainItemController
}

// New creates a new instance of the Module, initializing the service and controller.
func New() Module {
	service := svc.NewDomainItemService()
	controller := ctrl.NewDomainItemController(service)

	return Module{
		name:    "DomainItem",
		version: "v1",
		API:     controller,
	}
}

// Initialize initializes the module. Currently not implemented.
func (m Module) Initialize() error {
	panic("Not implemented")
}

// Version returns the version of the module.
func (m Module) Version() string {
	return m.version
}

// Name returns the name of the module.
func (m Module) Name() string {
	return m.name
}
