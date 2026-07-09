// Package fiddle is a client for Fastly Fiddle (https://fiddle.fastly.dev), the sandbox that compiles and runs VCL on real Fastly edge nodes.
//
// The API isn't documented and was built for the Fiddle web UI, so it has a few traps this package hides:
//
//   - VCL is sent as "vcl" but comes back as "src".
//   - "valid" means "it compiles" after a create or update, but "it ran at least once" on a GET.
package fiddle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultEndpoint is the public Fiddle service.
const DefaultEndpoint = "https://fiddle.fastly.dev"

// Subroutines lists the VCL subroutines a fiddle accepts, in the order a request goes through them.
var Subroutines = []string{
	"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log",
}

// MaxOrigins is how many origins a fiddle can have.
// The VCL sees them as backends F_origin_0 to F_origin_4.
const MaxOrigins = 5

// SourceIPs lists the values Request.SourceIP accepts: Fiddle's own "client" and "server" locations, or a country.
var SourceIPs = []string{"client", "server", "br", "cn", "de", "jp", "ru", "uk", "us", "za"}

// ConnTypes lists the values Request.ConnType accepts.
var ConnTypes = []string{"http", "h1", "h2"}

// Spec describes a fiddle to create or update.
type Spec struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
	// The server rejects empty Origins or Requests, but fills in defaults when they're left out.
	Origins  []string          `json:"origins,omitempty"`
	VCL      map[string]string `json:"vcl"`
	Requests []Request         `json:"requests,omitempty"`
}

// Request is one request a fiddle sends through the VCL.
type Request struct {
	Method          string `json:"method"`
	Path            string `json:"path"`
	Headers         string `json:"headers"`
	Body            string `json:"body"`
	EnableCluster   bool   `json:"enableCluster"`
	EnableShield    bool   `json:"enableShield"`
	UseFreshCache   bool   `json:"useFreshCache"`
	FollowRedirects bool   `json:"followRedirects"`
	ConnType        string `json:"connType"`
	SourceIP        string `json:"sourceIP"`
	Delay           int    `json:"delay"`
}

// NewRequest returns a Request with the same defaults as the Fiddle web UI.
func NewRequest() Request {
	return Request{
		Method:        "GET",
		Path:          "/",
		EnableCluster: true,
		ConnType:      "h2",
		SourceIP:      "client",
	}
}

// Diagnostic is one compiler message about a subroutine.
type Diagnostic struct {
	Level   string `json:"level"`
	Str     string `json:"str"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// SavedFiddle is what Create and Update return.
type SavedFiddle struct {
	ID         string
	SrcVersion int
	// Valid reports whether the VCL compiled.
	Valid bool
	// LintStatus is keyed by subroutine name.
	LintStatus map[string][]Diagnostic
}

type envelope struct {
	Fiddle struct {
		ID         string `json:"id"`
		SrcVersion int    `json:"srcVersion"`
	} `json:"fiddle"`
	Valid      bool                    `json:"valid"`
	LintStatus map[string][]Diagnostic `json:"lintStatus"`
}

// Client talks to a Fiddle service.
type Client struct {
	Endpoint   string
	HTTPClient interface {
		Do(*http.Request) (*http.Response, error)
	}
	UserAgent string
}

func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSuffix(c.Endpoint, "/")+path, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // #nosec G307
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		// Error bodies are HTML or plain text, and only the first line is worth showing.
		msg := strings.TrimSpace(string(data))
		if i := strings.IndexByte(msg, '\n'); i > 0 {
			msg = msg[:i]
		}
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return nil, fmt.Errorf("sandbox returned %s: %s", resp.Status, msg)
	}
	return data, nil
}

func (c *Client) save(ctx context.Context, method, path string, spec Spec) (*SavedFiddle, error) {
	if spec.VCL == nil {
		spec.VCL = map[string]string{}
	}
	data, err := c.do(ctx, method, path, spec)
	if err != nil {
		return nil, err
	}
	var env envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("unexpected response from sandbox: %w", err)
	}
	return &SavedFiddle{
		ID:         env.Fiddle.ID,
		SrcVersion: env.Fiddle.SrcVersion,
		Valid:      env.Valid,
		LintStatus: env.LintStatus,
	}, nil
}

// Create publishes a new fiddle and returns its compile status.
func (c *Client) Create(ctx context.Context, spec Spec) (*SavedFiddle, error) {
	return c.save(ctx, http.MethodPost, "/fiddle", spec)
}

// Update replaces an existing fiddle.
// Send the whole spec, since missing subroutines get cleared.
// When only the requests change, the edge doesn't need to sync again, which makes this much faster than Create.
func (c *Client) Update(ctx context.Context, id string, spec Spec) (*SavedFiddle, error) {
	return c.save(ctx, http.MethodPut, "/fiddle/"+id, spec)
}
