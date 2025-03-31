package manifest

import (
	"bytes"
	"fmt"

	"github.com/pelletier/go-toml"
)

// LocalServer represents a list of mocked Viceroy resources.
type LocalServer struct {
	Backends       map[string]LocalBackend     `toml:"backends"`
	ConfigStores   map[string]LocalConfigStore `toml:"config_stores,omitempty"`
	KVStores       LocalKVStoreMap             `toml:"kv_stores,omitempty"`
	SecretStores   LocalSecretStoreMap         `toml:"secret_stores,omitempty"`
	ViceroyVersion string                      `toml:"viceroy_version,omitempty"`
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

// KVStoreArrayEntry represents an array-based key/value store entries.
// It expects a key plus either a data or file field.
type KVStoreArrayEntry struct {
	Key  string `toml:"key"`
	File string `toml:"file,omitempty"`
	Data string `toml:"data,omitempty"`
}

// KVStoreExternalFile represents the external key/value store,
// which must have both a file and a format.
type KVStoreExternalFile struct {
	File   string `toml:"file"`
	Format string `toml:"format"`
}

// LocalKVStore represents a kv_store to be mocked by the local testing server.
// It is a union type and can either be an array of KVStoreArrayEntry or a single KVStoreExternalFile.
// The IsArray flag is used to preserve the original input style.
type LocalKVStore struct {
	IsArray  bool                 `toml:"-"`
	Array    []KVStoreArrayEntry  `toml:"-"`
	External *KVStoreExternalFile `toml:"-"`
}

// LocalKVStoreMap is a map of kv_store names to the local kv_store representation.
type LocalKVStoreMap map[string]LocalKVStore

// UnmarshalTOML performs custom unmarshalling of TOML data for LocalKVStoreMap.
func (m *LocalKVStoreMap) UnmarshalTOML(v interface{}) error {
	raw, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected kv_stores to be a TOML table")
	}

	result := make(LocalKVStoreMap)

	for key, val := range raw {
		switch typed := val.(type) {
		case []interface{}:
			var entries []KVStoreArrayEntry
			for _, item := range typed {
				obj, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid item in array for key %q", key)
				}
				var arrayEntry KVStoreArrayEntry
				if err := decodeTOMLMap(obj, &arrayEntry); err != nil {
					return fmt.Errorf("decode failed for array item in key %q: %w", key, err)
				}
				entries = append(entries, arrayEntry)
			}
			result[key] = LocalKVStore{
				IsArray: true,
				Array:   entries,
			}

		case map[string]interface{}:
			file, hasFile := typed["file"].(string)
			format, hasFormat := typed["format"].(string)

			if !hasFile || !hasFormat {
				return fmt.Errorf("key %q must have both file and format", key)
			}
			result[key] = LocalKVStore{
				IsArray: false,
				External: &KVStoreExternalFile{
					File:   file,
					Format: format,
				},
			}

		default:
			return fmt.Errorf("unsupported value type for key %q: %T", key, typed)
		}
	}

	*m = result
	return nil
}

// SecretStoreArrayEntry represents an array-based key/value store entries.
// It expects a key plus either a data or file field.
type SecretStoreArrayEntry struct {
	Key  string `toml:"key"`
	File string `toml:"file,omitempty"`
	Data string `toml:"data,omitempty"`
}

// SecretStoreExternalFile represents the external key/value store,
// which must have both a file and a format.
type SecretStoreExternalFile struct {
	File   string `toml:"file"`
	Format string `toml:"format"`
}

// LocalSecretStore represents a secret_store to be mocked by the local testing server.
// It is a union type and can either be an array of SecretStoreArrayEntry or a single SecretStoreExternalFile.
// The IsArray flag is used to preserve the original input style.
type LocalSecretStore struct {
	IsArray  bool                     `toml:"-"`
	Array    []SecretStoreArrayEntry  `toml:"-"`
	External *SecretStoreExternalFile `toml:"-"`
}

// LocalSecretStoreMap is a map of secret_store names to the local secret_store representation.
type LocalSecretStoreMap map[string]LocalSecretStore

// UnmarshalTOML performs custom unmarshalling of TOML data for LocalSecretStoreMap.
func (m *LocalSecretStoreMap) UnmarshalTOML(v interface{}) error {
	raw, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected secret_stores to be a TOML table")
	}

	result := make(LocalSecretStoreMap)

	for key, val := range raw {
		switch typed := val.(type) {
		case []interface{}:
			var entries []SecretStoreArrayEntry
			for _, item := range typed {
				obj, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid item in array for key %q", key)
				}
				var arrayEntry SecretStoreArrayEntry
				if err := decodeTOMLMap(obj, &arrayEntry); err != nil {
					return fmt.Errorf("decode failed for array item in key %q: %w", key, err)
				}
				entries = append(entries, arrayEntry)
			}
			result[key] = LocalSecretStore{
				IsArray: true,
				Array:   entries,
			}

		case map[string]interface{}:
			file, hasFile := typed["file"].(string)
			format, hasFormat := typed["format"].(string)

			if !hasFile || !hasFormat {
				return fmt.Errorf("key %q must have both file and format", key)
			}
			result[key] = LocalSecretStore{
				IsArray: false,
				External: &SecretStoreExternalFile{
					File:   file,
					Format: format,
				},
			}

		default:
			return fmt.Errorf("unsupported value type for key %q: %T", key, typed)
		}
	}

	*m = result
	return nil
}

func decodeTOMLMap(m map[string]interface{}, out interface{}) error {
	buf := new(bytes.Buffer)
	enc := toml.NewEncoder(buf)
	if err := enc.Encode(m); err != nil {
		return err
	}
	return toml.NewDecoder(buf).Decode(out)
}
