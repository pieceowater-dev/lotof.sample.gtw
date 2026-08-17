package graph

import (
	"app/internal/pkg/auth"
	"app/internal/pkg/domainItem"
	"app/internal/pkg/gateway"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthModule       auth.Module       `module:"Auth"`
	GatewayModule    gateway.Module    `module:"Gateway"`
	DomainItemModule domainItem.Module `module:"DomainItem"`
}
