package manifest

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
)

type KVWrapper struct {
	KVStores LocalKVStoreMap `toml:"kv_stores"`
}

func (w KVWrapper) MarshalTOML() ([]byte, error) {
	obj := make(map[string]any)
	kv := make(map[string]any)

	for key, entry := range w.KVStores {
		if entry.External != nil {
			kv[key] = map[string]any{
				"file":   entry.External.File,
				"format": entry.External.Format,
			}
		} else {
			kv[key] = entry.Array
		}
	}

	obj["kv_stores"] = kv

	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(obj)
	return buf.Bytes(), err
}

type SecretStoreWrapper struct {
	SecretStores LocalSecretStoreMap `toml:"secret_stores"`
}

func (w SecretStoreWrapper) MarshalTOML() ([]byte, error) {
	obj := make(map[string]any)
	kv := make(map[string]any)

	for key, entry := range w.SecretStores {
		if entry.External != nil {
			kv[key] = map[string]any{
				"file":   entry.External.File,
				"format": entry.External.Format,
			}
		} else {
			kv[key] = entry.Array
		}
	}

	obj["kv_stores"] = kv

	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(obj)
	return buf.Bytes(), err
}

func TestLocalKVStores_UnmarshalTOML(t *testing.T) {
	tests := []struct {
		name        string
		inputTOML   string
		expectError bool
		expected    LocalKVStore
	}{
		{
			name: "legacy array format",
			inputTOML: `
[[kv_stores.my-kv]]
key = "kv"
file = "kv.json"
`,
			expected: LocalKVStore{
				IsArray: true,
				Array: []KVStoreArrayEntry{
					{
						Key:  "kv",
						File: "kv.json",
					},
				},
			},
		},
		{
			name: "external file format",
			inputTOML: `
[kv_stores]
my-kv = { file = "kv.json", format = "json" }
`,
			expected: LocalKVStore{
				IsArray: false,
				External: &KVStoreExternalFile{
					File:   "kv.json",
					Format: "json",
				},
			},
		},
		{
			name: "invalid format",
			inputTOML: `
[kv_stores]
my-kv = "not-a-valid-entry"
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m KVWrapper

			decoder := toml.NewDecoder(strings.NewReader(tt.inputTOML))
			err := decoder.Decode(&m)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error for invalid format, but got none")
				}
				return
			} else if err != nil {
				t.Fatalf("Failed to parse TOML: %v", err)
			}

			got, ok := m.KVStores["my-kv"]
			if !ok {
				t.Fatalf("Expected key 'my-kv' not found")
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Mismatch!\nGot:  %+v\nWant: %+v", got, tt.expected)
			}
		})
	}
}

func TestLocalSecretStores_UnmarshalTOML(t *testing.T) {
	tests := []struct {
		name        string
		inputTOML   string
		expectError bool
		expected    LocalSecretStore
	}{
		{
			name: "legacy array format",
			inputTOML: `
[[secret_stores.my-secret-store]]
key = "secret"
file = "secret.json"
`,
			expected: LocalSecretStore{
				IsArray: true,
				Array: []SecretStoreArrayEntry{
					{
						Key:  "secret",
						File: "secret.json",
					},
				},
			},
		},
		{
			name: "external file format",
			inputTOML: `
[secret_stores]
my-secret-store = { file = "secret.json", format = "json" }
`,
			expected: LocalSecretStore{
				IsArray: false,
				External: &SecretStoreExternalFile{
					File:   "secret.json",
					Format: "json",
				},
			},
		},
		{
			name: "invalid format",
			inputTOML: `
[secret_stores]
my-secret-store = "not-a-valid-entry"
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m SecretStoreWrapper

			decoder := toml.NewDecoder(strings.NewReader(tt.inputTOML))
			err := decoder.Decode(&m)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error for invalid format, but got none")
				}
				return
			} else if err != nil {
				t.Fatalf("Failed to parse TOML: %v", err)
			}

			got, ok := m.SecretStores["my-secret-store"]
			if !ok {
				t.Fatalf("Expected key 'my-secret-store' not found")
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Mismatch!\nGot:  %+v\nWant: %+v", got, tt.expected)
			}
		})
	}
}
