// Package fiddle provides a client for the Fastly Fiddle API
// (https://fiddle.fastly.dev), the sandbox service that compiles and runs
// VCL on real Fastly edge nodes.
//
// The API is undocumented and was designed for the Fiddle web UI, so it has
// a few wire-format quirks this package absorbs:
//
//   - VCL is sent under a "vcl" key but returned under "src".
//   - The "valid" flag on a create/update response reports compilation, but
//     on a GET it reports whether the fiddle has ever been executed.
//   - Execution results arrive over a server-sent-events stream whose
//     sessions expire quickly, and a fresh publish can take from ten seconds
//     to two minutes to sync to the edge.
package fiddle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint is the public Fiddle service.
const DefaultEndpoint = "https://fiddle.fastly.dev"

// Subroutines lists the VCL subroutine slots a fiddle accepts, in the order
// the request flows through them.
var Subroutines = []string{
	"init", "recv", "hash", "hit", "miss", "pass", "fetch", "error", "deliver", "log",
}

// MaxOrigins is the number of origins a fiddle can declare.
// They are exposed to the VCL as backends F_origin_0 through F_origin_4.
const MaxOrigins = 5

// SourceIPs lists the accepted values for Request.SourceIP: the two Fiddle
// vantage points plus the country presets offered by the web UI.
var SourceIPs = []string{"client", "server", "br", "cn", "de", "jp", "ru", "uk", "us", "za"}

// ConnTypes lists the accepted values for Request.ConnType.
var ConnTypes = []string{"http", "h1", "h2"}

// Spec is the shape POSTed to create or update a fiddle.
type Spec struct {
	ID    string `json:"id,omitempty"`
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
	// Origins and Requests are rejected by the server when present but
	// empty; omitted, they fall back to server defaults (the http-me test
	// origin, and a single GET request).
	Origins  []string          `json:"origins,omitempty"`
	VCL      map[string]string `json:"vcl"`
	Requests []Request         `json:"requests,omitempty"`
}

// Request is one request definition inside a fiddle.
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

// Diagnostic is one entry in the lint report for a subroutine.
type Diagnostic struct {
	Level   string `json:"level"`
	Str     string `json:"str"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// SavedFiddle is the envelope returned by create and update calls.
type SavedFiddle struct {
	ID         string
	SrcVersion int
	// Valid reports whether the VCL compiled.
	// Only trust it on create/update responses; on a GET the same wire
	// field means "has been executed".
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
		// Error bodies are HTML or plain text, not JSON.
		// Keep the first line so lint failures like schema errors stay
		// legible.
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

// Update overwrites an existing fiddle in place.
// The whole spec must be sent; omitted subroutines are cleared, not kept.
// Changing only the request list leaves the source version (and its edge
// sync state) untouched.
func (c *Client) Update(ctx context.Context, id string, spec Spec) (*SavedFiddle, error) {
	return c.save(ctx, http.MethodPut, "/fiddle/"+id, spec)
}

// Execute starts an execution of the fiddle's requests on the edge and
// returns the session ID whose result stream reports the outcome.
// Requests sharing a cacheID share cache state; vary it to start cold.
func (c *Client) Execute(ctx context.Context, id string, cacheID int) (string, error) {
	data, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/fiddle/%s/execute?cacheID=%d", id, cacheID), nil)
	if err != nil {
		return "", err
	}
	var out struct {
		SessionID string `json:"sessionID"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("unexpected response from sandbox: %w", err)
	}
	if out.SessionID == "" {
		return "", fmt.Errorf("sandbox did not return an execution session")
	}
	return out.SessionID, nil
}

// sleep waits without outliving the context, for retry loops.
func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
