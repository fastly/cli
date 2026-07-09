// Package vcl implements sandbox-backed VCL testing commands.
// They compile VCL on real Fastly edge nodes before any service exists,
// using the Fiddle service (https://fiddle.fastly.dev) as infrastructure.
// No Fiddle concept surfaces in the interface: the commands speak in terms
// of VCL files, origins, and requests, so a local engine can replace the
// sandbox later without breaking anyone.
package vcl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/fiddle"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
)

// defaultOrigin is Fastly's public test origin, which echoes request
// details back, a useful default when the point is testing VCL, not an
// origin.
const defaultOrigin = "https://http-me.fastly.dev"

// specFlags collects the flags that describe the VCL and its origins.
type specFlags struct {
	origins   []string
	slotFiles map[string]*string
	endpoint  string
}

// slotHelp describes each subroutine slot in --help output.
var slotHelp = map[string]string{
	"init":    "one-time setup: backends, ACLs, tables",
	"recv":    "request inspection and routing",
	"hash":    "cache key composition",
	"hit":     "on cache hit",
	"miss":    "on cache miss, before the origin fetch",
	"pass":    "on explicit pass",
	"fetch":   "origin response handling (beresp)",
	"error":   "synthetic and error responses",
	"deliver": "final response mutations (resp)",
	"log":     "logging",
}

func (s *specFlags) register(cmd *kingpin.CmdClause) {
	cmd.Flag("origin", fmt.Sprintf("Origin server URL, including scheme (repeatable, up to %d). Origins become VCL backends F_origin_0, F_origin_1, ... in flag order, with the first as the default backend. Defaults to %s, a test origin that echoes requests back", fiddle.MaxOrigins, defaultOrigin)).PlaceHolder("URL").StringsVar(&s.origins)
	s.slotFiles = map[string]*string{}
	for _, slot := range fiddle.Subroutines {
		v := new(string)
		s.slotFiles[slot] = v
		cmd.Flag(slot, fmt.Sprintf("VCL file with the body of vcl_%s (%s)", slot, slotHelp[slot])).PlaceHolder("FILE").StringVar(v)
	}
	cmd.Flag("sandbox-endpoint", "Base URL of the sandbox service").Default(fiddle.DefaultEndpoint).Hidden().StringVar(&s.endpoint)
}

var subWrapper = regexp.MustCompile(`(?m)^\s*sub\s+vcl_(\w+)\s*\{`)

// vclSources maps subroutine slot to the file providing its body.
type vclSources map[string]string

// buildVCL reads the slot files into subroutine bodies.
// Positional files, if any, have their slot inferred from the file name
// (recv.vcl, vcl_recv.vcl).
func (s *specFlags) buildVCL(positional []string) (map[string]string, vclSources, error) {
	sources := vclSources{}
	for slot, f := range s.slotFiles {
		if *f != "" {
			sources[slot] = *f
		}
	}
	for _, path := range positional {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		slot := strings.TrimPrefix(base, "vcl_")
		if !slices.Contains(fiddle.Subroutines, slot) {
			return nil, nil, fsterr.RemediationError{
				Inner:       fmt.Errorf("cannot tell which subroutine %q holds", path),
				Remediation: fmt.Sprintf("Name the file after its subroutine (e.g. recv.vcl, fetch.vcl) or pass it with an explicit flag (e.g. --recv=%s). Subroutines: %s.", path, strings.Join(fiddle.Subroutines, ", ")),
			}
		}
		if prev, dup := sources[slot]; dup {
			return nil, nil, fmt.Errorf("both %s and %s provide vcl_%s", prev, path, slot)
		}
		sources[slot] = path
	}
	if len(sources) == 0 {
		return nil, nil, fsterr.RemediationError{
			Inner:       fmt.Errorf("no VCL given"),
			Remediation: "Pass at least one VCL file, either positionally (recv.vcl) or with a subroutine flag (--recv=FILE).",
		}
	}
	vcl := map[string]string{}
	for slot, path := range sources {
		data, err := os.ReadFile(path) // #nosec G304 (user-supplied path)
		if err != nil {
			return nil, nil, err
		}
		body := string(data)
		if m := subWrapper.FindStringSubmatch(body); m != nil {
			return nil, nil, fsterr.RemediationError{
				Inner:       fmt.Errorf("%s contains a subroutine declaration (sub vcl_%s { ... })", path, m[1]),
				Remediation: "Provide only the body of the subroutine; the sandbox adds the sub wrapper and the Fastly boilerplate itself.",
			}
		}
		vcl[slot] = body
	}
	return vcl, sources, nil
}

// buildOrigins validates the --origin flags, falling back to the default
// test origin.
func (s *specFlags) buildOrigins() ([]string, error) {
	if len(s.origins) == 0 {
		return []string{defaultOrigin}, nil
	}
	if len(s.origins) > fiddle.MaxOrigins {
		return nil, fmt.Errorf("at most %d origins are supported", fiddle.MaxOrigins)
	}
	for _, o := range s.origins {
		u, err := url.Parse(o)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return nil, fmt.Errorf("origin %q must be a URL with an http or https scheme", o)
		}
	}
	return s.origins, nil
}

// client returns a fiddle client for the configured endpoint.
func (s *specFlags) client(g *global.Data) *fiddle.Client {
	return &fiddle.Client{
		Endpoint:   s.endpoint,
		HTTPClient: g.HTTPClient,
		UserAgent:  useragent.Name,
	}
}

// diagnostic is the provider-neutral form of a compile finding, used by
// the text and JSON output of check.
type diagnostic struct {
	Subroutine string `json:"subroutine"`
	File       string `json:"file,omitempty"`
	Line       int    `json:"line"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Excerpt    string `json:"excerpt,omitempty"`
}

// flattenDiagnostics converts a lint report into printable diagnostics,
// attributing each to the file that provided the subroutine.
func flattenDiagnostics(lint map[string][]fiddle.Diagnostic, sources vclSources) []diagnostic {
	var out []diagnostic
	slots := make([]string, 0, len(lint))
	for slot := range lint {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	for _, slot := range slots {
		for _, d := range lint[slot] {
			out = append(out, diagnostic{
				Subroutine: slot,
				File:       sources[slot],
				Line:       d.Line + 1,
				Level:      d.Level,
				Message:    d.Message,
				Excerpt:    d.Str,
			})
		}
	}
	return out
}

// printDiagnostics writes compiler-style, editor-clickable lines.
func printDiagnostics(out io.Writer, diags []diagnostic) {
	for _, d := range diags {
		where := d.File
		if where == "" {
			where = "vcl_" + d.Subroutine
		}
		msg := d.Message
		if d.Excerpt != "" {
			msg = fmt.Sprintf("%s: %s", msg, d.Excerpt)
		}
		text.Output(out, "%s:%d: %s: %s", where, d.Line, d.Level, msg)
	}
}

// errCompileFailed is returned after compile diagnostics have been printed.
var errCompileFailed = fsterr.RemediationError{
	Inner:       fmt.Errorf("the VCL failed to compile"),
	Remediation: "Fix the problems reported above and re-run the command.",
}

// sandboxState remembers the sandbox created for this user so repeat runs
// update it in place (seconds) instead of publishing fresh (up to minutes
// of edge sync).
// Purely an optimization: losing the file just means the next run starts
// cold.
type sandboxState struct {
	ID string `json:"id"`
}

// sandboxStatePath keeps the state file next to the CLI's config file.
func sandboxStatePath() string {
	return filepath.Join(filepath.Dir(config.FilePath), "vcl-sandbox.json")
}

func loadSandboxState() (s sandboxState) {
	data, err := os.ReadFile(sandboxStatePath()) // #nosec G304
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	return s
}

func saveSandboxState(s sandboxState) {
	path := sandboxStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	data, _ := json.Marshal(s)
	_ = os.WriteFile(path, data, 0o600)
}

// newSpec assembles a sandbox spec with the CLI's title, so anyone who
// stumbles on the fiddle in the web UI knows where it came from.
func newSpec(origins []string, vcl map[string]string, requests ...fiddle.Request) fiddle.Spec {
	return fiddle.Spec{
		Type:     "vcl",
		Title:    "fastly CLI test sandbox",
		Origins:  origins,
		VCL:      vcl,
		Requests: requests,
	}
}

// publish updates the remembered sandbox with the spec, creating a fresh
// one when there is none or the old one is gone.
// The state is updated (and persisted) with the resulting sandbox ID.
func publish(ctx context.Context, client *fiddle.Client, spec fiddle.Spec, state *sandboxState) (*fiddle.SavedFiddle, error) {
	if state.ID != "" {
		if saved, err := client.Update(ctx, state.ID, spec); err == nil {
			return saved, nil
		}
	}
	saved, err := client.Create(ctx, spec)
	if err != nil {
		return nil, err
	}
	if saved.ID != "" && saved.ID != state.ID {
		state.ID = saved.ID
		saveSandboxState(*state)
	}
	return saved, nil
}
