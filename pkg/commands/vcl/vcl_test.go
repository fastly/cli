package vcl_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/testutil"
)

// fakeSandbox is a minimal in-memory stand-in for the Fiddle API.
func fakeSandbox(t *testing.T, valid bool) *httptest.Server {
	t.Helper()
	envelope := func(id string) string {
		if valid {
			return fmt.Sprintf(`{"fiddle":{"id":%q,"srcVersion":0},"valid":true,"lintStatus":{}}`, id)
		}
		return fmt.Sprintf(`{"fiddle":{"id":%q,"srcVersion":0},"valid":false,"lintStatus":{"recv":[{"level":"error","line":0,"startPos":4,"endPos":10,"str":"req.bogus","message":"Unknown variable"}]}}`, id)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /fiddle", func(w http.ResponseWriter, r *http.Request) {
		var spec map[string]any
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, envelope("cafe0123"))
	})
	mux.HandleFunc("PUT /fiddle/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, envelope(r.PathValue("id")))
	})
	mux.HandleFunc("POST /fiddle/{id}/execute", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"sessionID":"feedbeef00","streamHost":""}`)
	})
	mux.HandleFunc("GET /results/{sid}/stream", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, `event: updateResult
data: {"execHost":"s-r-cafe0123v0-1-100.exec9.fiddle.fastly.dev","clientFetches":{"a":{"status":200,"complete":true,"resp":"HTTP/2 200 OK\ncontent-type: text/plain\nx-cache: MISS","bodyPreview":"hello","bodyBytesReceived":5,"isText":true}},"events":[{"type":"vcl-sub","fnName":"recv","attribs":{"return":"lookup"}}]}

`)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// writeVCL writes a valid subroutine body to <dir>/<name> and returns its path.
func writeVCL(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// isolatedHome keeps the sandbox state file out of the real user config dir.
func isolatedHome(t *testing.T) map[string]string {
	t.Helper()
	home := t.TempDir()
	return map[string]string{"HOME": home, "XDG_CONFIG_HOME": home, "AppData": home}
}

func TestVCLCheck(t *testing.T) {
	dir := t.TempDir()
	recv := writeVCL(t, dir, "recv.vcl", "set req.http.X = \"1\";\n")
	wrapped := writeVCL(t, dir, "wrapped.recv.vcl", "sub vcl_recv {\n}\n")
	mystery := writeVCL(t, dir, "mystery.vcl", "")

	valid := fakeSandbox(t, true)
	invalid := fakeSandbox(t, false)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate no VCL given",
			Args:      "",
			WantError: "no VCL given",
			EnvVars:   isolatedHome(t),
		},
		{
			Name:      "validate slot inference failure",
			Args:      mystery,
			WantError: "cannot tell which subroutine",
			EnvVars:   isolatedHome(t),
		},
		{
			Name:      "reject files with sub wrappers",
			Args:      "--recv " + wrapped + " --sandbox-endpoint " + valid.URL,
			WantError: "contains a subroutine declaration",
			EnvVars:   isolatedHome(t),
		},
		{
			Name:      "reject too many origins",
			Args:      recv + " --origin https://a.example --origin https://b.example --origin https://c.example --origin https://d.example --origin https://e.example --origin https://f.example --sandbox-endpoint " + valid.URL,
			WantError: "at most 5 origins",
			EnvVars:   isolatedHome(t),
		},
		{
			Name:      "reject bad origin",
			Args:      recv + " --origin ftp://a.example --sandbox-endpoint " + valid.URL,
			WantError: "must be a URL with an http or https scheme",
			EnvVars:   isolatedHome(t),
		},
		{
			Name:       "compile success via positional file",
			Args:       recv + " --sandbox-endpoint " + valid.URL,
			WantOutput: "VCL compiled cleanly.",
			EnvVars:    isolatedHome(t),
		},
		{
			Name:       "compile success via slot flag",
			Args:       "--recv " + recv + " --sandbox-endpoint " + valid.URL,
			WantOutput: "VCL compiled cleanly.",
			EnvVars:    isolatedHome(t),
		},
		{
			Name:       "compile failure prints diagnostics",
			Args:       recv + " --sandbox-endpoint " + invalid.URL,
			WantError:  "the VCL failed to compile",
			WantOutput: recv + ":1: error: Unknown variable",
			EnvVars:    isolatedHome(t),
		},
		{
			Name:       "compile failure in json",
			Args:       recv + " --json --sandbox-endpoint " + invalid.URL,
			WantError:  "the VCL failed to compile",
			WantOutput: `"valid": false`,
			EnvVars:    isolatedHome(t),
		},
	}
	testutil.RunCLIScenarios(t, []string{"vcl", "check"}, scenarios)
}

func TestVCLRun(t *testing.T) {
	dir := t.TempDir()
	recv := writeVCL(t, dir, "recv.vcl", "set req.http.X = \"1\";\n")
	valid := fakeSandbox(t, true)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "reject unknown client-from",
			Args:      recv + " --client-from atlantis --sandbox-endpoint " + valid.URL,
			WantError: `unknown --client-from value "atlantis"`,
			EnvVars:   isolatedHome(t),
		},
		{
			Name:      "reject unknown connection",
			Args:      recv + " --connection h3 --sandbox-endpoint " + valid.URL,
			WantError: `unknown --connection value "h3"`,
			EnvVars:   isolatedHome(t),
		},
		{
			Name:      "reject malformed header",
			Args:      recv + " --header nocolon --sandbox-endpoint " + valid.URL,
			WantError: `header "nocolon" is not in 'Name: value' form`,
			EnvVars:   isolatedHome(t),
		},
		{
			Name:       "replay a request",
			Args:       recv + " --path /hello --sandbox-endpoint " + valid.URL,
			WantOutput: "GET /hello → 200 (text/plain, 5 bytes, MISS)",
			EnvVars:    isolatedHome(t),
		},
		{
			Name:       "replay body is shown",
			Args:       recv + " --path /hello --sandbox-endpoint " + valid.URL,
			WantOutput: "hello",
			EnvVars:    isolatedHome(t),
		},
		{
			Name:       "json output",
			Args:       recv + " --json --sandbox-endpoint " + valid.URL,
			WantOutput: `"provider": "fiddle"`,
			EnvVars:    isolatedHome(t),
		},
	}
	testutil.RunCLIScenarios(t, []string{"vcl", "run"}, scenarios)
}

func TestVCLServeValidation(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate no VCL given",
			Args:      "",
			WantError: "no VCL given",
			EnvVars:   isolatedHome(t),
		},
	}
	testutil.RunCLIScenarios(t, []string{"vcl", "serve"}, scenarios)
}
