package vcl_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/testutil"
)

// fakeSandbox is a minimal in-memory stand-in for the Fiddle API.
func fakeSandbox(t *testing.T, valid bool) *httptest.Server {
	t.Helper()
	saved := `{"fiddle":{"id":"cafe0123"},"valid":true}`
	if !valid {
		saved = `{"fiddle":{"id":"cafe0123"},"valid":false,"lintStatus":{"recv":[{"level":"error","line":0,"str":"req.bogus","message":"Unknown variable"}]}}`
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /fiddle", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, saved)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

func writeVCL(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// runVCL runs the scenarios for a vcl subcommand, each with its own sandbox state file in a temp dir.
func runVCL(t *testing.T, command string, scenarios []testutil.CLIScenario) {
	t.Helper()
	for i := range scenarios {
		scenarios[i].ConfigPath = filepath.Join(t.TempDir(), config.FileName)
	}
	testutil.RunCLIScenarios(t, []string{"vcl", command}, scenarios)
}

func TestVCLCheck(t *testing.T) {
	dir := t.TempDir()
	recv := writeVCL(t, dir, "recv.vcl", "set req.http.X = \"1\";\n")
	wrapped := writeVCL(t, dir, "wrapped.recv.vcl", "sub vcl_recv {\n}\n")
	mystery := writeVCL(t, dir, "mystery.vcl", "")

	valid := fakeSandbox(t, true)
	invalid := fakeSandbox(t, false)

	runVCL(t, "check", []testutil.CLIScenario{
		{
			Name:      "validate no VCL given",
			WantError: "no VCL given",
		},
		{
			Name:      "validate slot inference failure",
			Args:      mystery,
			WantError: "cannot tell which subroutine",
		},
		{
			Name:      "reject files with sub wrappers",
			Args:      "--recv " + wrapped + " --sandbox-endpoint " + valid.URL,
			WantError: "contains a subroutine declaration",
		},
		{
			Name:      "reject too many origins",
			Args:      recv + " --origin https://a.example --origin https://b.example --origin https://c.example --origin https://d.example --origin https://e.example --origin https://f.example --sandbox-endpoint " + valid.URL,
			WantError: "at most 5 origins",
		},
		{
			Name:      "reject bad origin",
			Args:      recv + " --origin ftp://a.example --sandbox-endpoint " + valid.URL,
			WantError: "must be a URL with an http or https scheme",
		},
		{
			Name:       "compile success",
			Args:       recv + " --sandbox-endpoint " + valid.URL,
			WantOutput: "VCL compiled cleanly.",
		},
		{
			Name:       "compile failure prints diagnostics",
			Args:       recv + " --sandbox-endpoint " + invalid.URL,
			WantError:  "the VCL failed to compile",
			WantOutput: recv + ":1: error: Unknown variable",
		},
		{
			Name:       "compile failure in json",
			Args:       recv + " --json --sandbox-endpoint " + invalid.URL,
			WantError:  "the VCL failed to compile",
			WantOutput: `"valid": false`,
		},
	})
}
