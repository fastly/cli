package mock

// ConfigFile is a mock implementation of the toml.ReadWriter interface that's
// used for testing.
type ConfigFile struct {
	PathFn   func() string
	ExistsFn func() bool
	ReadFn   func(c any) error
	WriteFn  func(c any) error
}

// Path satisfies the toml.ReadWriter interface for testing purposes.
func (c *ConfigFile) Path() string {
	return c.PathFn()
}

// Exists satisfies the toml.ReadWriter interface for testing purposes.
func (c *ConfigFile) Exists() bool {
	return c.ExistsFn()
}

// Read satisfies the toml.ReadWriter interface for testing purposes.
func (c *ConfigFile) Read(config any) error {
	return c.ReadFn(config)
}

// Write satisfies the toml.ReadWriter interface for testing purposes.
func (c *ConfigFile) Write(config any) error {
	return c.WriteFn(config)
}

// NewNonExistentConfigFile is a test helper function which constructs a new
// non-existent config file interface.
func NewNonExistentConfigFile() *ConfigFile {
	return &ConfigFile{
		PathFn:   func() string { return "" },
		ExistsFn: func() bool { return false },
	}
}
