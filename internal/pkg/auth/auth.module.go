package auth

import (
	"app/internal/pkg/auth/ctrl"
	"app/internal/pkg/auth/svc"
)

type Module struct {
	name    string
	version string
	API     *ctrl.AuthController
}

// New wires the auth module.
func New() Module {
	service := svc.NewAuthService()
	controller := ctrl.NewAuthController(service)

	return Module{
		name:    "Auth",
		version: "v1",
		API:     controller,
	}
}

func (m Module) Initialize() error { return nil }

func (m Module) Version() string { return m.version }

func (m Module) Name() string { return m.name }
