package compute

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsnotify/fsnotify"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
)

// TestIsContentChange checks that content/existence events trigger a rebuild and
// attribute-only (Chmod) events do not.
func TestIsContentChange(t *testing.T) {
	scenarios := []struct {
		name string
		op   fsnotify.Op
		want bool
	}{
		{name: "create", op: fsnotify.Create, want: true},
		{name: "write", op: fsnotify.Write, want: true},
		{name: "remove", op: fsnotify.Remove, want: true},
		{name: "rename", op: fsnotify.Rename, want: true},
		{name: "chmod only", op: fsnotify.Chmod, want: false},
		{name: "write combined with chmod", op: fsnotify.Write | fsnotify.Chmod, want: true},
		{name: "no op", op: 0, want: false},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			if got := isContentChange(testcase.op); got != testcase.want {
				t.Errorf("isContentChange(%s) = %v, want %v", testcase.op, got, testcase.want)
			}
		})
	}
}

func TestSetBackendsWithDefaultOverrideHostIfMissing_TemporaryManifest(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "fastly.toml")
	const initial = `manifest_version = "3.0.0"
name = "test"

[local_server.backends.http-me]
url = "https://http-me.glitch.me"
`
	if err := os.WriteFile(manifestPath, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	var file manifest.File
	if err := file.Read(manifestPath); err != nil {
		t.Fatal(err)
	}

	c := &ServeCommand{
		Base: argparser.Base{
			Globals: &global.Data{
				Manifest: &manifest.Data{File: file},
			},
		},
	}

	var out bytes.Buffer
	tmpManifest, cleanup, err := c.setBackendsWithDefaultOverrideHostIfMissing(&out, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if tmpManifest == manifestPath {
		t.Fatalf("expected temporary manifest path, got original %s", manifestPath)
	}

	// Verify temporary manifest exists and contains override_host.
	tmpData, err := os.ReadFile(tmpManifest)
	if err != nil {
		t.Fatalf("failed to read temporary manifest: %v", err)
	}
	tmpBody := string(tmpData)
	if !strings.Contains(tmpBody, "override_host") || !strings.Contains(tmpBody, "http-me.glitch.me") {
		t.Fatalf("expected override_host in temporary manifest, got:\n%s", tmpBody)
	}
	if !strings.Contains(out.String(), "http-me.glitch.me") {
		t.Fatalf("expected override_host info message, got: %q", out.String())
	}

	// Verify original manifest was not mutated on disk.
	origWritten, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(origWritten) != initial {
		t.Fatalf("original manifest was mutated on disk:\n%s", string(origWritten))
	}

	// Verify cleanup removes the temporary file.
	cleanup()
	if _, err := os.Stat(tmpManifest); !os.IsNotExist(err) {
		t.Fatalf("expected temporary manifest to be removed after cleanup, stat err: %v", err)
	}
}

func TestSetBackendsWithDefaultOverrideHostIfMissing_BackendMatrix(t *testing.T) {
	tests := []struct {
		name                string
		backendURL          string
		initialOverrideHost string
		wantOverrideHost    string
		wantTemporaryFile   bool
		wantInfoMessage     bool
	}{
		{
			name:                "hostname without port",
			backendURL:          "https://example.com",
			initialOverrideHost: "",
			wantOverrideHost:    "example.com",
			wantTemporaryFile:   true,
			wantInfoMessage:     true,
		},
		{
			name:                "hostname with port",
			backendURL:          "http://example.com:8080",
			initialOverrideHost: "",
			wantOverrideHost:    "example.com:8080",
			wantTemporaryFile:   true,
			wantInfoMessage:     true,
		},
		{
			name:                "ipv4 without port",
			backendURL:          "http://127.0.0.1",
			initialOverrideHost: "",
			wantOverrideHost:    "",
			wantTemporaryFile:   false,
			wantInfoMessage:     false,
		},
		{
			name:                "ipv4 with port",
			backendURL:          "http://127.0.0.1:8080",
			initialOverrideHost: "",
			wantOverrideHost:    "",
			wantTemporaryFile:   false,
			wantInfoMessage:     false,
		},
		{
			name:                "ipv6 without port",
			backendURL:          "http://[::1]",
			initialOverrideHost: "",
			wantOverrideHost:    "",
			wantTemporaryFile:   false,
			wantInfoMessage:     false,
		},
		{
			name:                "ipv6 with port",
			backendURL:          "http://[::1]:8080",
			initialOverrideHost: "",
			wantOverrideHost:    "",
			wantTemporaryFile:   false,
			wantInfoMessage:     false,
		},
		{
			name:                "ipv6 full address with port",
			backendURL:          "http://[2001:db8::1]:8080",
			initialOverrideHost: "",
			wantOverrideHost:    "",
			wantTemporaryFile:   false,
			wantInfoMessage:     false,
		},
		{
			name:                "hostname with pre-existing override_host",
			backendURL:          "https://example.com",
			initialOverrideHost: "custom.domain.test",
			wantOverrideHost:    "custom.domain.test",
			wantTemporaryFile:   false,
			wantInfoMessage:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			manifestPath := filepath.Join(dir, "fastly.toml")
			if err := os.WriteFile(manifestPath, []byte("manifest_version = \"3.0.0\"\nname = \"test\"\n"), 0o600); err != nil {
				t.Fatal(err)
			}

			file := manifest.File{
				LocalServer: manifest.LocalServer{
					Backends: map[string]manifest.LocalBackend{
						"test-backend": {
							URL:          tc.backendURL,
							OverrideHost: tc.initialOverrideHost,
						},
					},
				},
			}

			c := &ServeCommand{
				Base: argparser.Base{
					Globals: &global.Data{
						Manifest: &manifest.Data{File: file},
					},
				},
			}

			var out bytes.Buffer
			manifestToUse, cleanup, err := c.setBackendsWithDefaultOverrideHostIfMissing(&out, manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			gotBackend := c.Globals.Manifest.File.LocalServer.Backends["test-backend"]
			if gotBackend.OverrideHost != tc.wantOverrideHost {
				t.Errorf("override_host = %q, want %q", gotBackend.OverrideHost, tc.wantOverrideHost)
			}

			if tc.wantTemporaryFile {
				if manifestToUse == manifestPath {
					t.Errorf("expected temporary manifest file, got original path %s", manifestPath)
				}
				if _, err := os.Stat(manifestToUse); err != nil {
					t.Errorf("expected temporary manifest to exist on disk: %v", err)
				}
			} else if manifestToUse != manifestPath {
				t.Errorf("expected original manifest path %s, got %s", manifestPath, manifestToUse)
			}

			if tc.wantInfoMessage && !strings.Contains(out.String(), "is configured without an `override_host`") {
				t.Errorf("expected info message in output, got: %q", out.String())
			} else if !tc.wantInfoMessage && strings.Contains(out.String(), "is configured without an `override_host`") {
				t.Errorf("unexpected info message in output: %q", out.String())
			}
		})
	}
}
