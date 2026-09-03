package compute

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// mockWasmToolsSource is a tiny Go program that stands in for `wasm-tools
// metadata show --json`: it prints whatever JSON is given via the
// MOCK_WASM_TOOLS_JSON environment variable to stdout. Compiling a real
// executable (rather than a bash script) keeps this mock runnable on
// Windows, where exec.Command cannot interpret a shebang line.
const mockWasmToolsSource = `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Print(os.Getenv("MOCK_WASM_TOOLS_JSON"))
}
`

// buildMockWasmTools compiles mockWasmToolsSource into an executable in dir
// and returns its path.
func buildMockWasmTools(t *testing.T, dir string) string {
	t.Helper()

	src := filepath.Join(dir, "mock_wasm_tools.go")
	if err := os.WriteFile(src, []byte(mockWasmToolsSource), 0o600); err != nil {
		t.Fatal(err)
	}

	binName := "mock-wasm-tools"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	bin := filepath.Join(dir, binName)

	// #nosec G204 -- fixed arguments, no user input
	if out, err := exec.Command("go", "build", "-o", bin, src).CombinedOutput(); err != nil {
		t.Fatalf("failed to build mock wasm-tools binary: %v\n%s", err, out)
	}

	return bin
}

func TestReadExistingPackageInfo(t *testing.T) {
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

	wasmtoolsBin := buildMockWasmTools(t, rootdir)

	scenarios := []struct {
		name         string
		jsonOutput   string
		expectedData *DataCollectionPackageInfo
	}{
		{
			name:       "extracts from component-based metadata structure",
			jsonOutput: `{"component":{"metadata":{"producers":[["processed-by",{"fastly_data":"{\"package_info\":{\"packages\":{\"foo\":\"1.0.0\"}},\"script_info\":{\"build_script\":\"echo component\"}}"}]]}}}`,
			expectedData: &DataCollectionPackageInfo{
				Packages: map[string]string{"foo": "1.0.0"},
			},
		},
		{
			name:       "extracts from module-based metadata structure",
			jsonOutput: `{"module":{"producers":[["processed-by",{"fastly_data":"{\"package_info\":{\"packages\":{\"bar\":\"2.0.0\"}},\"script_info\":{\"build_script\":\"echo module\"}}"}]]}}`,
			expectedData: &DataCollectionPackageInfo{
				Packages: map[string]string{"bar": "2.0.0"},
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
			t.Setenv("MOCK_WASM_TOOLS_JSON", tc.jsonOutput)

			cmd := &BuildCommand{}
			actualData := cmd.readExistingPackageInfo(wasmtoolsBin)

			if tc.expectedData == nil {
				if actualData != nil {
					t.Fatalf("expected nil, got: %+v", actualData)
				}
				return
			}

			if actualData == nil {
				t.Fatal("expected non-nil DataCollectionPackageInfo, got nil")
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
