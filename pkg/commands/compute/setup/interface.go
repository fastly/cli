package setup

// Interface represents the behaviour of a [setup] resource.
type Interface interface {
	// Configure prompts the user for specific values related to the service resource.
	Configure() error

	// Create calls the relevant API to create the service resource(s).
	Create() error

	// Missing indicates if there are missing resources that need to be
	// configured and/or created.
	Missing() bool

	// Predefined indicates if the service resource has been specified within the
	// fastly.toml file using a [setup] configuration block.
	Predefined() bool

	// Validate checks if the service has the required resources.
	Validate() error
}
