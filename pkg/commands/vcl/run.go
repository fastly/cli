package vcl

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/fiddle"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// clientFromCountries lists the country presets accepted by --client-from,
// mapped to their wire codes.
// The wire codes themselves are accepted too.
var clientFromCountries = map[string]string{
	"brazil":         "br",
	"china":          "cn",
	"germany":        "de",
	"japan":          "jp",
	"russia":         "ru",
	"south-africa":   "za",
	"united-kingdom": "uk",
	"united-states":  "us",
}

// clientFromValues holds the friendly --client-from values, for shell
// completion hints; clientFromHelp is the country subset for --help text.
var clientFromValues, clientFromHelp = func() ([]string, string) {
	countries := make([]string, 0, len(clientFromCountries))
	for name := range clientFromCountries {
		countries = append(countries, name)
	}
	sort.Strings(countries)
	return append([]string{"client", "server"}, countries...), strings.Join(countries, ", ")
}()

// requestFlags collects the flags describing the request to replay.
type requestFlags struct {
	path            string
	method          string
	headers         []string
	body            string
	clientFrom      string
	connection      string
	shield          bool
	noCluster       bool
	followRedirects bool
	freshCache      bool
}

func (r *requestFlags) register(cmd *kingpin.CmdClause) {
	cmd.Flag("path", "Request path, including any query string").Default("/").StringVar(&r.path)
	cmd.Flag("method", "HTTP method (GET, HEAD, POST, PUT, PATCH, DELETE, OPTIONS, PURGE, ...)").Default("GET").StringVar(&r.method)
	cmd.Flag("header", "Request header as 'Name: value' (repeatable)").Short('H').PlaceHolder("HEADER").StringsVar(&r.headers)
	cmd.Flag("body", "Request body; prefix with @ to read from a file").StringVar(&r.body)
	cmd.Flag("client-from", "Where the client appears to connect from, for geolocation-dependent VCL: 'client' (your real vantage point), 'server', or a country: "+clientFromHelp).Default("client").HintOptions(clientFromValues...).StringVar(&r.clientFrom)
	cmd.Flag("connection", "Protocol between client and edge: h2 (HTTP/2, default), h1 (HTTP/1.1 over TLS), or http (plain HTTP/1.1)").Default("h2").HintOptions(fiddle.ConnTypes...).StringVar(&r.connection)
	cmd.Flag("shield", "Enable shielding (a second cache layer, as when a shield POP is configured)").BoolVar(&r.shield)
	cmd.Flag("no-cluster", "Disable request clustering within the POP").BoolVar(&r.noCluster)
	cmd.Flag("follow-redirects", "Follow 3xx responses instead of reporting them").BoolVar(&r.followRedirects)
	cmd.Flag("fresh-cache", "Start from an empty cache instead of the cache state previous runs built up").BoolVar(&r.freshCache)
}

func (r *requestFlags) toRequest() (fiddle.Request, error) {
	req := fiddle.NewRequest()
	req.Path = r.path
	req.Method = strings.ToUpper(r.method)
	req.EnableCluster = !r.noCluster
	req.EnableShield = r.shield
	req.FollowRedirects = r.followRedirects
	req.UseFreshCache = r.freshCache

	var headers []string
	for _, h := range r.headers {
		if !strings.Contains(h, ":") {
			return req, fmt.Errorf("header %q is not in 'Name: value' form", h)
		}
		headers = append(headers, strings.TrimSpace(h))
	}
	req.Headers = strings.Join(headers, "\n")

	if strings.HasPrefix(r.body, "@") {
		data, err := os.ReadFile(r.body[1:]) // #nosec G304 (user-supplied path)
		if err != nil {
			return req, err
		}
		req.Body = string(data)
	} else {
		req.Body = r.body
	}

	from := strings.ToLower(r.clientFrom)
	if wire, ok := clientFromCountries[from]; ok {
		from = wire
	}
	if !slices.Contains(fiddle.SourceIPs, from) {
		return req, fsterr.RemediationError{
			Inner:       fmt.Errorf("unknown --client-from value %q", r.clientFrom),
			Remediation: "Valid values: client, server, " + clientFromHelp + ".",
		}
	}
	req.SourceIP = from

	if !slices.Contains(fiddle.ConnTypes, r.connection) {
		return req, fmt.Errorf("unknown --connection value %q (valid: %s)", r.connection, strings.Join(fiddle.ConnTypes, ", "))
	}
	req.ConnType = r.connection
	return req, nil
}

// RunCommand replays a request through VCL running on a real Fastly edge
// node and reports what happened.
type RunCommand struct {
	argparser.Base
	argparser.JSONOutput

	spec    specFlags
	request requestFlags
	files   []string
	timeout time.Duration
}

// NewRunCommand returns a usable command registered under the parent.
func NewRunCommand(parent argparser.Registerer, g *global.Data) *RunCommand {
	c := RunCommand{
		Base: argparser.Base{
			Globals:         g,
			SuppressVerbose: true,
		},
	}
	c.CmdClause = parent.Command("run", "Run VCL on a real Fastly edge node by replaying a request through it (no service or API token needed)")
	c.CmdClause.Arg("file", "VCL files holding subroutine bodies, named after their subroutine (recv.vcl, fetch.vcl, ...)").StringsVar(&c.files)
	c.spec.register(c.CmdClause)
	c.request.register(c.CmdClause)
	c.CmdClause.Flag("timeout", "How long to wait for the edge before giving up").Default("2m").DurationVar(&c.timeout)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// Exec invokes the application logic for the command.
func (c *RunCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	vcl, sources, err := c.spec.buildVCL(c.files)
	if err != nil {
		return err
	}
	origins, err := c.spec.buildOrigins()
	if err != nil {
		return err
	}
	request, err := c.request.toRequest()
	if err != nil {
		return err
	}

	ctx := context.Background()
	client := c.spec.client(c.Globals)
	state := loadSandboxState()
	saved, err := publish(ctx, client, newSpec(origins, vcl, request), &state)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	if !saved.Valid {
		printDiagnostics(out, flattenDiagnostics(saved.LintStatus, sources))
		return errCompileFailed
	}

	if !c.JSONOutput.Enabled {
		printOrigins(out, origins, len(c.spec.origins) == 0)
	}

	if state.CacheID == 0 || c.request.freshCache {
		state.CacheID = freshCacheID()
		saveSandboxState(state)
	}

	result, err := client.Run(ctx, saved.ID, state.CacheID, fiddle.StreamOptions{
		WantFetches: 1,
		MaxWait:     c.timeout,
		Syncing: func() {
			if !c.JSONOutput.Enabled {
				text.Info(out, "Deploying VCL to a test sandbox... (a first run can take a minute to reach the edge)")
			}
		},
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fetch := pickFetch(result)
	if fetch == nil {
		return fmt.Errorf("the edge did not report a result in time; try again or raise --timeout")
	}

	if c.JSONOutput.Enabled {
		_, err := c.WriteJSON(out, newRunResult(request, result, fetch))
		return err
	}
	renderFetch(out, request, fetch)
	if c.Globals.Verbose() {
		renderTrace(out, result, fetch)
	}
	return nil
}

// pickFetch selects the fetch to report.
// The result keys fetches by internal ID, so iterate deterministically
// and, when redirects produced several, prefer the final (non-3xx)
// response.
func pickFetch(r *fiddle.Result) *fiddle.ClientFetch {
	keys := make([]string, 0, len(r.ClientFetch))
	for k, f := range r.ClientFetch {
		if f.Complete {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)
	var fallback *fiddle.ClientFetch
	for _, k := range keys {
		f := r.ClientFetch[k]
		if f.Status < 300 || f.Status >= 400 {
			return &f
		}
		if fallback == nil {
			fallback = &f
		}
	}
	return fallback
}

// parseRespHead splits a raw response head ("HTTP/2 200 OK\nname: value\n...")
// into header pairs.
func parseRespHead(resp string) (headers [][2]string) {
	lines := strings.Split(strings.ReplaceAll(resp, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return nil
	}
	for _, line := range lines[1:] {
		name, value, ok := strings.Cut(line, ":")
		if !ok || name == "" {
			continue
		}
		headers = append(headers, [2]string{strings.ToLower(strings.TrimSpace(name)), strings.TrimSpace(value)})
	}
	return headers
}

func headerValue(headers [][2]string, name string) string {
	for _, h := range headers {
		if h[0] == name {
			return h[1]
		}
	}
	return ""
}

// renderFetch prints the response summary and body:
//
//	GET /robots.txt → 200 (text/plain, 34 bytes, MISS)
func renderFetch(out io.Writer, req fiddle.Request, fetch *fiddle.ClientFetch) {
	headers := parseRespHead(fetch.Resp)
	var details []string
	if ct := headerValue(headers, "content-type"); ct != "" {
		if mt, _, ok := strings.Cut(ct, ";"); ok {
			ct = strings.TrimSpace(mt)
		}
		details = append(details, ct)
	}
	details = append(details, fmt.Sprintf("%d bytes", fetch.BodyBytesReceived))
	if cache := headerValue(headers, "x-cache"); cache != "" {
		details = append(details, lastCacheHop(cache))
	}
	text.Output(out, "%s %s → %d (%s)", req.Method, req.Path, fetch.Status, strings.Join(details, ", "))

	if fetch.BodyBytesReceived == 0 {
		return
	}
	text.Break(out)
	if !fetch.IsText {
		text.Output(out, "(binary body, %d bytes)", fetch.BodyBytesReceived)
		return
	}
	text.Output(out, "%s", text.SanitizeTerminalOutput(fetch.BodyPreview))
	if len(fetch.BodyPreview) < fetch.BodyBytesReceived {
		text.Break(out)
		text.Info(out, "Body truncated: showing the first %d of %d bytes.", len(fetch.BodyPreview), fetch.BodyBytesReceived)
	}
}

// renderTrace prints the full response head, the VCL subroutine trace, and
// any origin fetches.
func renderTrace(out io.Writer, result *fiddle.Result, fetch *fiddle.ClientFetch) {
	if fetch != nil {
		text.Break(out)
		text.Output(out, "Response:")
		for _, line := range strings.Split(strings.TrimSpace(fetch.Resp), "\n") {
			text.Indent(out, 2, "%s", text.SanitizeTerminalOutput(line))
		}
	}
	if len(result.Events) > 0 {
		text.Break(out)
		text.Output(out, "VCL trace:")
		for _, e := range result.Events {
			if e.Type != "vcl-sub" {
				continue
			}
			line := "vcl_" + e.FnName
			if ret, ok := e.Attribs["return"].(string); ok && ret != "" {
				line += " → return(" + ret + ")"
			}
			if e.Server.POP != "" {
				line += "  [POP " + e.Server.POP + "]"
			}
			text.Indent(out, 2, "%s", line)
		}
	}
	if len(result.OriginFetch) > 0 {
		text.Break(out)
		text.Output(out, "Origin fetches:")
		for _, f := range result.OriginFetch {
			reqLine, _ := firstLine(f.Req)
			text.Indent(out, 2, "%s → %d", reqLine, f.Status)
		}
	}
}

func firstLine(s string) (string, bool) {
	line, _, found := strings.Cut(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	return line, found
}

// runResult is the machine-readable shape of a run.
// Provider-specific details (like the executing POP inside events) are
// optional: they may be absent under a future local engine.
type runResult struct {
	Provider      string           `json:"provider"`
	Request       runRequest       `json:"request"`
	Response      runResponse      `json:"response"`
	Events        []runEvent       `json:"events"`
	OriginFetches []runOriginFetch `json:"origin_fetches"`
}

type runRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type runResponse struct {
	Status       int                 `json:"status"`
	Headers      map[string][]string `json:"headers"`
	BodyPreview  string              `json:"body_preview"`
	BodyBytes    int                 `json:"body_bytes"`
	BodyComplete bool                `json:"body_complete"`
}

type runEvent struct {
	Subroutine string         `json:"subroutine"`
	Attributes map[string]any `json:"attributes,omitempty"`
	POP        string         `json:"pop,omitempty"`
}

type runOriginFetch struct {
	Request string `json:"request"`
	Status  int    `json:"status"`
}

func newRunResult(req fiddle.Request, result *fiddle.Result, fetch *fiddle.ClientFetch) runResult {
	headerMap := map[string][]string{}
	for _, h := range parseRespHead(fetch.Resp) {
		headerMap[h[0]] = append(headerMap[h[0]], h[1])
	}
	var origins []runOriginFetch
	for _, f := range result.OriginFetch {
		reqLine, _ := firstLine(f.Req)
		origins = append(origins, runOriginFetch{Request: reqLine, Status: f.Status})
	}
	var events []runEvent
	for _, e := range result.Events {
		if e.Type != "vcl-sub" {
			continue
		}
		events = append(events, runEvent{Subroutine: "vcl_" + e.FnName, Attributes: e.Attribs, POP: e.Server.POP})
	}
	return runResult{
		Provider: "fiddle",
		Request:  runRequest{Method: req.Method, Path: req.Path},
		Response: runResponse{
			Status:       fetch.Status,
			Headers:      headerMap,
			BodyPreview:  fetch.BodyPreview,
			BodyBytes:    fetch.BodyBytesReceived,
			BodyComplete: len(fetch.BodyPreview) >= fetch.BodyBytesReceived,
		},
		Events:        events,
		OriginFetches: origins,
	}
}
