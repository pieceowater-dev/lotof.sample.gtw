package interfaces

// IRouter defines the interface for a router with methods to initialize the router and retrieve modules.
type IRouter interface {
	// InitializeRouter initializes the router and returns the initialized router or an error if the initialization fails.
	InitializeRouter() (any, error)

	// GetModules returns a map of module names to their corresponding IModule instances.
	GetModules() map[string]IModule
}
