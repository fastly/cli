package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadKVStoreFile(t *testing.T) {
	// Create a temporary JSON file for testing
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")

	jsonContent := `{
		"key1": "value1",
		"key2": "value2",
		"key3": {"nested": "object"}
	}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		file      string
		wantErr   bool
		wantCount int
	}{
		{
			name:      "valid json file",
			file:      jsonFile,
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:    "non-existent file",
			file:    "/nonexistent/path/file.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := loadKVStoreFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadKVStoreFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(items) != tt.wantCount {
				t.Errorf("loadKVStoreFile() got %d items, want %d", len(items), tt.wantCount)
			}

			// Verify that nested objects are marshaled to JSON strings
			if !tt.wantErr {
				foundNested := false
				for _, item := range items {
					if item.Key == "key3" {
						foundNested = true
						if item.Value != `{"nested":"object"}` {
							t.Errorf("nested object not properly marshaled: got %q", item.Value)
						}
					}
				}
				if !foundNested {
					t.Error("expected to find key3 with nested object")
				}
			}
		})
	}
}
