package manifest

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/pelletier/go-toml"
)

type KVWrapper struct {
	KVStores LocalKVStoreMap `toml:"kv_stores"`
}

func (w KVWrapper) MarshalTOML() ([]byte, error) {
	obj := make(map[string]interface{})
	kv := make(map[string]interface{})

	for key, entry := range w.KVStores {
		if entry.External != nil {
			kv[key] = map[string]interface{}{
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
	obj := make(map[string]interface{})
	kv := make(map[string]interface{})

	for key, entry := range w.SecretStores {
		if entry.External != nil {
			kv[key] = map[string]interface{}{
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
		name         string
		inputTOML    string
		expectError  bool
		expectArray  bool
		expectLength int
		wantFile     string
	}{
		{
			name: "legacy array format",
			inputTOML: `
[[kv_stores.my-kv]]
key = "my-kv"
file = "kv.json"
`,
			expectArray:  true,
			expectLength: 1,
			wantFile:     "kv.json",
		},
		{
			name: "external file format",
			inputTOML: `
[kv_stores]
my-kv = { file = "kv.json", format = "json" }
`,
			expectArray:  false,
			expectLength: 0,
			wantFile:     "kv.json",
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

			buf := new(bytes.Buffer)
			encoder := toml.NewEncoder(buf)

			encodeErr := encoder.Encode(m)
			if encodeErr != nil {
				log.Fatalf("TOML encode failed: %v", encodeErr)
			}

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

			if got.IsArray != tt.expectArray {
				t.Fatalf("Expected IsArray=%v, got %v", tt.expectArray, got.IsArray)
			}

			if tt.expectArray {
				if len(got.Array) != tt.expectLength {
					t.Fatalf("Expected %d inline entries, got %d", tt.expectLength, len(got.Array))
				}
				if got.Array[0].File != tt.wantFile {
					t.Errorf("Expected file %q, got %q", tt.wantFile, got.Array[0].File)
				}
			} else {
				if got.External == nil {
					t.Fatal("Expected KVStoreExternalFile but got nil")
				}
				if got.External.File != tt.wantFile {
					t.Errorf("Expected file %q, got %q", tt.wantFile, got.External.File)
				}
			}
		})
	}
}

func TestLocalSecretStores_UnmarshalTOML(t *testing.T) {
	tests := []struct {
		name         string
		inputTOML    string
		expectError  bool
		expectArray  bool
		expectLength int
		wantFile     string
	}{
		{
			name: "legacy array format",
			inputTOML: `
[[secret_stores.my-secret-store]]
key = "secret"
file = "secret.json"
`,
			expectArray:  true,
			expectLength: 1,
			wantFile:     "secret.json",
		},
		{
			name: "external file format",
			inputTOML: `
[secret_stores]
my-secret-store = { file = "kv.json", format = "json" }
`,
			expectArray:  false,
			expectLength: 0,
			wantFile:     "kv.json",
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

			buf := new(bytes.Buffer)
			encoder := toml.NewEncoder(buf)

			encodeErr := encoder.Encode(m)
			if encodeErr != nil {
				log.Fatalf("TOML encode failed: %v", encodeErr)
			}

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

			if got.IsArray != tt.expectArray {
				t.Fatalf("Expected IsArray=%v, got %v", tt.expectArray, got.IsArray)
			}

			if tt.expectArray {
				if len(got.Array) != tt.expectLength {
					t.Fatalf("Expected %d inline entries, got %d", tt.expectLength, len(got.Array))
				}
				if got.Array[0].File != tt.wantFile {
					t.Errorf("Expected file %q, got %q", tt.wantFile, got.Array[0].File)
				}
			} else {
				if got.External == nil {
					t.Fatal("Expected SecretStoreExternalFile but got nil")
				}
				if got.External.File != tt.wantFile {
					t.Errorf("Expected file %q, got %q", tt.wantFile, got.External.File)
				}
			}
		})
	}
}
