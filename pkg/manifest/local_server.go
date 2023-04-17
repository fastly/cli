package manifest

// LocalServer represents a list of mocked Viceroy resources.
type LocalServer struct {
	Backends     map[string]LocalBackend       `toml:"backends"`
	ConfigStores map[string]LocalConfigStore   `toml:"config_stores,omitempty"`
	KVStores     map[string][]LocalKVStore     `toml:"kv_stores,omitempty"`
	SecretStores map[string][]LocalSecretStore `toml:"secret_stores,omitempty"`
}

// LocalBackend represents a backend to be mocked by the local testing server.
type LocalBackend struct {
	URL          string `toml:"url"`
	OverrideHost string `toml:"override_host,omitempty"`
	CertHost     string `toml:"cert_host,omitempty"`
	UseSNI       bool   `toml:"use_sni,omitempty"`
}

// LocalConfigStore represents a config store to be mocked by the local testing server.
type LocalConfigStore struct {
	File     string            `toml:"file,omitempty"`
	Format   string            `toml:"format"`
	Contents map[string]string `toml:"contents,omitempty"`
}

// LocalKVStore represents an kv_store to be mocked by the local testing server.
type LocalKVStore struct {
	Key  string `toml:"key"`
	File string `toml:"file,omitempty"`
	Data string `toml:"data,omitempty"`
}

// LocalSecretStore represents a secret_store to be mocked by the local testing server.
type LocalSecretStore struct {
	Key  string `toml:"key"`
	File string `toml:"file,omitempty"`
	Data string `toml:"data,omitempty"`
}
