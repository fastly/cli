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

func TestSetBackendsWithDefaultOverrideHostIfMissingWritesManifest(t *testing.T) {
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
	if err := c.setBackendsWithDefaultOverrideHostIfMissing(&out, manifestPath); err != nil {
		t.Fatal(err)
	}

	written, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	body := string(written)
	if !strings.Contains(body, "override_host") || !strings.Contains(body, "http-me.glitch.me") {
		t.Fatalf("expected override_host persisted in manifest, got:\n%s", body)
	}
	if !strings.Contains(out.String(), "http-me.glitch.me") {
		t.Fatalf("expected override_host info message, got: %q", out.String())
	}
}
