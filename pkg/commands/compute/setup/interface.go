package setup

type Interface interface {
	// Configure prompts the user (if necessary) for specific values related to
	// the service resource. It is expected to be called only after first calling
	// Missing() and finding `true` returned.
	Configure() error

	// Create calls the relevant API to create the service resource(s). It is
	// expected to be called only after first calling Missing() and finding `true`
	// returned.
	Create() error

	// Missing indicates if there are missing resources that need to be created.
	Missing() bool

	// Predefined indicates if the service resource has been specified within the
	// fastly.toml file using a [setup] configuration block.
	Predefined() bool

	// Validate checks if the service has the required resources. It should set
	// an internal `missing` field (boolean) accordingly so that the Missing()
	// method can report the state of the resource.
	Validate() error
}
