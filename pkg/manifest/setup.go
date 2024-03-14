package manifest

// Setup represents a set of service configuration that works with the code in
// the package. See https://www.fastly.com/documentation/reference/compute/fastly-toml.
type Setup struct {
	Backends     map[string]*SetupBackend     `toml:"backends,omitempty"`
	ConfigStores map[string]*SetupConfigStore `toml:"config_stores,omitempty"`
	Loggers      map[string]*SetupLogger      `toml:"log_endpoints,omitempty"`
	ObjectStores map[string]*SetupKVStore     `toml:"object_stores,omitempty"`
	KVStores     map[string]*SetupKVStore     `toml:"kv_stores,omitempty"`
	SecretStores map[string]*SetupSecretStore `toml:"secret_stores,omitempty"`
}

// Defined indicates if there is any [setup] configuration in the manifest.
func (s Setup) Defined() bool {
	var defined bool

	if len(s.Backends) > 0 {
		defined = true
	}
	if len(s.ConfigStores) > 0 {
		defined = true
	}
	if len(s.Loggers) > 0 {
		defined = true
	}
	if len(s.KVStores) > 0 {
		defined = true
	}

	return defined
}

// SetupBackend represents a '[setup.backends.<T>]' instance.
type SetupBackend struct {
	Address     string `toml:"address,omitempty"`
	Port        int    `toml:"port,omitempty"`
	Description string `toml:"description,omitempty"`
}

// SetupConfigStore represents a '[setup.dictionaries.<T>]' instance.
type SetupConfigStore struct {
	Items       map[string]SetupConfigStoreItems `toml:"items,omitempty"`
	Description string                           `toml:"description,omitempty"`
}

// SetupConfigStoreItems represents a '[setup.dictionaries.<T>.items]' instance.
type SetupConfigStoreItems struct {
	Value       string `toml:"value,omitempty"`
	Description string `toml:"description,omitempty"`
}

// SetupLogger represents a '[setup.log_endpoints.<T>]' instance.
type SetupLogger struct {
	Provider string `toml:"provider,omitempty"`
}

// SetupKVStore represents a '[setup.kv_stores.<T>]' instance.
type SetupKVStore struct {
	Items       map[string]SetupKVStoreItems `toml:"items,omitempty"`
	Description string                       `toml:"description,omitempty"`
}

// SetupKVStoreItems represents a '[setup.kv_stores.<T>.items]' instance.
type SetupKVStoreItems struct {
	File        string `toml:"file,omitempty"`
	Value       string `toml:"value,omitempty"`
	Description string `toml:"description,omitempty"`
}

// SetupSecretStore represents a '[setup.secret_stores.<T>]' instance.
type SetupSecretStore struct {
	Entries     map[string]SetupSecretStoreEntry `toml:"entries,omitempty"`
	Description string                           `toml:"description,omitempty"`
}

// SetupSecretStoreEntry represents a '[setup.secret_stores.<T>.entries]' instance.
type SetupSecretStoreEntry struct {
	// The secret value is intentionally omitted to avoid secrets
	// from being included in the manifest. Instead, secret
	// values are input during setup.
	Description string `toml:"description,omitempty"`
}
