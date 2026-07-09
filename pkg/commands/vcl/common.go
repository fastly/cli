// Package vcl has the commands that test VCL on real Fastly edge nodes, without needing a service.
//
// They use Fiddle (https://fiddle.fastly.dev) under the hood, but only talk about VCL files, origins, and requests.
// That way, a local engine can replace Fiddle later without breaking anyone.
package vcl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/fastly/kingpin"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/fiddle"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
)

// defaultOrigin echoes requests back, which is handy when you're testing VCL and not an origin.
const defaultOrigin = "https://http-me.fastly.dev"

// specFlags are the flags that describe the VCL and its origins.
type specFlags struct {
	origins   []string
	slotFiles map[string]*string
	endpoint  string
}

// slotHelp is the --help text for each subroutine flag.
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

// vclSources maps each subroutine to the file its body comes from.
type vclSources map[string]string

// buildVCL reads the VCL files given as flags or arguments.
// Arguments are matched to a subroutine by file name, like recv.vcl or vcl_recv.vcl.
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
	vcl, err := readSources(sources)
	if err != nil {
		return nil, nil, err
	}
	return vcl, sources, nil
}

// readSources reads the subroutine bodies.
// It's separate from buildVCL so that serve --watch can reload the same files.
func readSources(sources vclSources) (map[string]string, error) {
	vcl := map[string]string{}
	for slot, path := range sources {
		data, err := os.ReadFile(path) // #nosec G304 (user-supplied path)
		if err != nil {
			return nil, err
		}
		body := string(data)
		if m := subWrapper.FindStringSubmatch(body); m != nil {
			return nil, fsterr.RemediationError{
				Inner:       fmt.Errorf("%s contains a subroutine declaration (sub vcl_%s { ... })", path, m[1]),
				Remediation: "Provide only the body of the subroutine; the sandbox adds the sub wrapper and the Fastly boilerplate itself.",
			}
		}
		vcl[slot] = body
	}
	return vcl, nil
}

// buildOrigins checks the --origin flags, and falls back to the test origin when there are none.
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

func (s *specFlags) client(g *global.Data) *fiddle.Client {
	return &fiddle.Client{
		Endpoint:   s.endpoint,
		HTTPClient: g.HTTPClient,
		UserAgent:  useragent.Name,
	}
}

// printOrigins shows the backend-to-origin mapping when it isn't obvious.
func printOrigins(out io.Writer, origins []string, defaulted bool) {
	if defaulted {
		text.Info(out, "No --origin given; requests go to the test origin %s.", defaultOrigin)
		return
	}
	if len(origins) > 1 {
		for i, o := range origins {
			text.Output(out, "Backend F_origin_%d = %s", i, o)
		}
	}
}

// lastCacheHop turns an X-Cache value like "MISS, HIT" into the part the client saw.
func lastCacheHop(v string) string {
	parts := strings.Split(v, ",")
	return strings.TrimSpace(parts[len(parts)-1])
}

// freshCacheID picks a new cache bucket, which is how we get a cold cache.
func freshCacheID() int {
	return rand.IntN(1_000_000_000) + 1 // #nosec G404 (bucket name, not security)
}

// diagnostic is a compiler message that doesn't depend on Fiddle's format.
type diagnostic struct {
	Subroutine string `json:"subroutine"`
	File       string `json:"file,omitempty"`
	Line       int    `json:"line"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Excerpt    string `json:"excerpt,omitempty"`
}

// flattenDiagnostics turns a lint report into a sorted list, with the file each message is about.
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

// printDiagnostics prints "file:line: level: message" lines that editors can jump to.
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

// sandboxState remembers the user's sandbox, so the next run can update it in seconds instead of waiting minutes for a new one to reach the edge.
// If the file gets lost, the next run just starts from scratch.
type sandboxState struct {
	ID string `json:"id"`
	// CacheID is shared across runs, so running the same request twice can show a MISS and then a HIT.
	// --fresh-cache picks a new one.
	CacheID int `json:"cache_id"`

	path string
}

// loadSandboxState reads the state file, which lives next to the CLI config file.
func loadSandboxState(g *global.Data) sandboxState {
	s := sandboxState{path: filepath.Join(filepath.Dir(g.ConfigPath), "vcl-sandbox.json")}
	if data, err := os.ReadFile(s.path); err == nil { // #nosec G304
		_ = json.Unmarshal(data, &s)
	}
	return s
}

func saveSandboxState(s sandboxState) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return
	}
	data, _ := json.Marshal(s)
	_ = os.WriteFile(s.path, data, 0o600)
}

// newSpec builds a sandbox spec.
// The title tells anyone who finds the fiddle in the web UI where it came from.
func newSpec(origins []string, vcl map[string]string, requests ...fiddle.Request) fiddle.Spec {
	return fiddle.Spec{
		Type:     "vcl",
		Title:    "fastly CLI test sandbox",
		Origins:  origins,
		VCL:      vcl,
		Requests: requests,
	}
}

// publish updates the user's sandbox, or creates a new one if there isn't one yet or it's gone.
// A new sandbox ID gets saved to the state file.
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
