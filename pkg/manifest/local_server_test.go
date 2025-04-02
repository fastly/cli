package manifest

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
)

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
			var m struct {
				KVStores LocalKVStoreMap `toml:"kv_stores"`
			}

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
			var m struct {
				SecretStores LocalSecretStoreMap `toml:"secret_stores"`
			}

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
