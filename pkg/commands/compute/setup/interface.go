package setup

type Interface interface {
	// Configure prompts the user (if necessary) for specific values related to
	// the service resource.
	Configure() error

	// Create calls the relevant API to create the service resource(s).
	Create() error

	// Missing indicates if there are missing resources that need to be created.
	Missing() bool

	// Predefined indicates if the service resource has been specified within the
	// fastly.toml file using a [setup] configuration block.
	Predefined() bool
}
