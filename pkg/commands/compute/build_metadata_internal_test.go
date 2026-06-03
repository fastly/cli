package compute

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// mockWasmToolsScript generates a mock executable shell script for wasm-tools.
//
// Because the production code runs exec.Command under the hood, we mock it by writing
// a temporary executable bash script to disk that outputs the mock JSON we expect.
// We use a bash heredoc (cat << 'EOF') so that the JSON structure and inner quotes
// are written exactly as-is, avoiding platform-specific shell escape/echo issues.
func mockWasmToolsScript(staticOutput string) string {
	return "#!/usr/bin/env bash\ncat << 'EOF'\n" + staticOutput + "\nEOF"
}

func TestReadExistingFastlyData(t *testing.T) {
	// Create a temporary directory for our mock environment
	rootdir, err := os.MkdirTemp("", "fastly-metadata-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootdir)

	// Save original PWD and return to it later
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(pwd)
	}()

	// Ensure the bin directory and main.wasm file exist
	// (binWasmPath points to "./bin/main.wasm")
	if err := os.MkdirAll("bin", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binWasmPath, []byte("mock-wasm-binary"), 0o600); err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		name         string
		jsonOutput   string
		expectedData *DataCollection
	}{
		{
			name:       "extracts from component-based metadata structure",
			jsonOutput: `{"component":{"metadata":{"producers":[["processed-by",{"fastly_data":"{\"package_info\":{\"packages\":{\"foo\":\"1.0.0\"}},\"script_info\":{\"build_script\":\"echo component\"}}"}]]}}}`,
			expectedData: &DataCollection{
				PackageInfo: DataCollectionPackageInfo{
					Packages: map[string]string{"foo": "1.0.0"},
				},
				ScriptInfo: DataCollectionScriptInfo{
					BuildScript: "echo component",
				},
			},
		},
		{
			name:       "extracts from module-based metadata structure",
			jsonOutput: `{"module":{"producers":[["processed-by",{"fastly_data":"{\"package_info\":{\"packages\":{\"bar\":\"2.0.0\"}},\"script_info\":{\"build_script\":\"echo module\"}}"}]]}}`,
			expectedData: &DataCollection{
				PackageInfo: DataCollectionPackageInfo{
					Packages: map[string]string{"bar": "2.0.0"},
				},
				ScriptInfo: DataCollectionScriptInfo{
					BuildScript: "echo module",
				},
			},
		},
		{
			name:         "handles missing fastly_data gracefully",
			jsonOutput:   `{"component":{"metadata":{"producers":[["processed-by",{"other_tool":"1.0.0"}]]}}}`,
			expectedData: nil,
		},
		{
			name:         "handles invalid JSON from wasm-tools gracefully",
			jsonOutput:   `invalid-json`,
			expectedData: nil,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			wasmtoolsBin := filepath.Join(rootdir, "mock-wasm-tools")
			scriptContent := mockWasmToolsScript(tc.jsonOutput)
			// #nosec G306 -- mock binary must be executable
			if err := os.WriteFile(wasmtoolsBin, []byte(scriptContent), 0o700); err != nil {
				t.Fatal(err)
			}

			cmd := &BuildCommand{}
			actualData := cmd.readExistingFastlyData(wasmtoolsBin)

			if tc.expectedData == nil {
				if actualData != nil {
					t.Fatalf("expected nil, got: %+v", actualData)
				}
				return
			}

			if actualData == nil {
				t.Fatal("expected non-nil DataCollection, got nil")
			}

			// Validate values
			expectedBytes, _ := json.Marshal(tc.expectedData)
			actualBytes, _ := json.Marshal(actualData)
			if string(expectedBytes) != string(actualBytes) {
				t.Errorf("\nwant: %s\ngot:  %s", expectedBytes, actualBytes)
			}
		})
	}
}
