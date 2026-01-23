package manifest

import "reflect"

// Setup represents a set of service configuration that works with the code in
// the package. See https://www.fastly.com/documentation/reference/compute/fastly-toml.
type Setup struct {
	Backends     map[string]*SetupBackend     `toml:"backends,omitempty"`
	ConfigStores map[string]*SetupConfigStore `toml:"config_stores,omitempty"`
	Loggers      map[string]*SetupLogger      `toml:"log_endpoints,omitempty"`
	ObjectStores map[string]*SetupKVStore     `toml:"object_stores,omitempty"`
	KVStores     map[string]*SetupKVStore     `toml:"kv_stores,omitempty"`
	SecretStores map[string]*SetupSecretStore `toml:"secret_stores,omitempty"`
	Products     *SetupProducts               `toml:"products,omitempty"`
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
	if s.Products != nil && s.Products.AnyDefined() {
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

type SetupProducts struct {
	APIDiscovery        *SetupProduct      `toml:"api_discovery,omitempty"`
	BotManagement       *SetupProduct      `toml:"bot_management,omitempty"`
	BrotliCompression   *SetupProduct      `toml:"brotli_compression,omitempty"`
	DdosProtection      *SetupProduct      `toml:"ddos_protection,omitempty"`
	DomainInspector     *SetupProduct      `toml:"domain_inspector,omitempty"`
	Fanout              *SetupProduct      `toml:"fanout,omitempty"`
	ImageOptimizer      *SetupProduct      `toml:"image_optimizer,omitempty"`
	LogExplorerInsights *SetupProduct      `toml:"log_explorer_insights,omitempty"`
	Ngwaf               *SetupProductNgwaf `toml:"ngwaf,omitempty"`
	OriginInspector     *SetupProduct      `toml:"origin_inspector,omitempty"`
	WebSockets          *SetupProduct      `toml:"websockets,omitempty"`
}

func (p *SetupProducts) AnyDefined() bool {
	if p == nil {
		return false
	}

	rv := reflect.ValueOf(p).Elem() // SetupProducts
	settingsT := reflect.TypeOf((*SetupProductSettings)(nil)).Elem()

	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		if fv.Kind() != reflect.Ptr || fv.IsNil() {
			continue
		}

		if fv.Type().Implements(settingsT) {
			return true
		}
	}

	return false
}

type SetupProductSettings interface {
	Enabled() bool
}

type SetupProduct struct {
	Enable bool `toml:"enable,omitempty"`
}

var _ SetupProductSettings = (*SetupProduct)(nil)

func (p *SetupProduct) Enabled() bool {
	return p != nil && p.Enable
}

type SetupProductNgwaf struct {
	SetupProduct
	WorkspaceID string `toml:"workspace_id,omitempty"`
}

var _ SetupProductSettings = (*SetupProductNgwaf)(nil)
