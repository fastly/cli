package manifest

import (
	"fmt"
	"strings"
)

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
	if len(s.ObjectStores) > 0 {
		defined = true
	}
	if len(s.KVStores) > 0 {
		defined = true
	}
	if len(s.SecretStores) > 0 {
		defined = true
	}
	if s.Products != nil && s.Products.AnyEnabled() {
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
	ApiDiscovery        *SetupProductEnable `toml:"api_discovery,omitempty"`
	BotManagement       *SetupProductEnable `toml:"bot_management,omitempty"`
	BrotliCompression   *SetupProductEnable `toml:"brotli_compression,omitempty"`
	DdosProtection      *SetupProductEnable `toml:"ddos_protection,omitempty"`
	DomainInspector     *SetupProductEnable `toml:"domain_inspector,omitempty"`
	Fanout              *SetupProductEnable `toml:"fanout,omitempty"`
	ImageOptimizer      *SetupProductEnable `toml:"image_optimizer,omitempty"`
	LogExplorerInsights *SetupProductEnable `toml:"log_explorer_insights,omitempty"`
	Ngwaf               *SetupProductNgwaf  `toml:"ngwaf,omitempty"`
	OriginInspector     *SetupProductEnable `toml:"origin_inspector,omitempty"`
	WebSockets          *SetupProductEnable `toml:"websockets,omitempty"`
}

type SetupProduct interface {
	Enabled() bool
	Validate() error
}

type SetupProductEnable struct {
	Enable bool `toml:"enable,omitempty"`
}

var _ SetupProduct = (*SetupProductEnable)(nil)

func (p *SetupProductEnable) Enabled() bool {
	return p != nil && p.Enable
}
func (p *SetupProductEnable) Validate() error {
	return nil
}

type SetupProductNgwaf struct {
	SetupProductEnable
	WorkspaceID string `toml:"workspace_id,omitempty"`
}

var _ SetupProduct = (*SetupProductNgwaf)(nil)

func (p *SetupProductNgwaf) Enabled() bool {
	if p == nil {
		return false
	}
	return p.SetupProductEnable.Enabled()
}
func (p *SetupProductNgwaf) Validate() error {
	if p == nil || !p.Enable {
		return nil
	}
	if strings.TrimSpace(p.WorkspaceID) == "" {
		return fmt.Errorf("workspace_id is required when enable = true")
	}
	return nil
}

func (p *SetupProducts) AllSettings() []SetupProduct {
	if p == nil {
		return nil
	}

	return []SetupProduct{
		p.ApiDiscovery,
		p.BotManagement,
		p.BrotliCompression,
		p.DdosProtection,
		p.DomainInspector,
		p.Fanout,
		p.ImageOptimizer,
		p.LogExplorerInsights,
		p.Ngwaf,
		p.OriginInspector,
		p.WebSockets,
	}
}

func (p *SetupProducts) AnyEnabled() bool {
	for _, s := range p.AllSettings() {
		if s != nil && s.Enabled() {
			return true
		}
	}
	return false
}
