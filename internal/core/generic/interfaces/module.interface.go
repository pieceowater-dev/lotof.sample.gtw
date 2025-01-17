package interfaces

// IModule defines the interface for a module with initialization, version, and name retrieval methods.
type IModule interface {
	// Initialize initializes the module and returns an error if the initialization fails.
	Initialize() error

	// Version returns the version of the module.
	Version() string

	// Name returns the name of the module.
	Name() string
}
