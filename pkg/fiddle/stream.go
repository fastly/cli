package fiddle

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ClientFetch is what the fiddle's test client saw for one request.
type ClientFetch struct {
	// Req and Resp are the raw request and response head (status line
	// followed by header lines, newline separated).
	Req  string `json:"req"`
	Resp string `json:"resp"`
	// Status is the HTTP status the client received.
	Status int `json:"status"`
	// BodyPreview holds roughly the first kilobyte of the response body.
	BodyPreview       string `json:"bodyPreview"`
	BodyBytesReceived int    `json:"bodyBytesReceived"`
	IsText            bool   `json:"isText"`
	Complete          bool   `json:"complete"`
}

// OriginFetch is one request the edge made to an origin.
type OriginFetch struct {
	Req    string `json:"req"`
	Resp   string `json:"resp"`
	Status int    `json:"status"`
}

// Event is one VCL subroutine invocation reported by the edge.
// The wire object carries more (timestamps, request IDs, log lines); only
// what the CLI presents is decoded.
type Event struct {
	Type   string `json:"type"`
	FnName string `json:"fnName"`
	Server struct {
		POP    string `json:"pop"`
		NodeID string `json:"nodeID"`
	} `json:"server"`
	Attribs map[string]any `json:"attribs"`
}

// Result is a snapshot of an execution, as reported by the result stream.
// The server may send several, each more complete than the last.
type Result struct {
	ID           string                 `json:"id"`
	RequestCount int                    `json:"requestCount"`
	ExecHost     string                 `json:"execHost"`
	ExecVersion  int                    `json:"execVersion"`
	ClientFetch  map[string]ClientFetch `json:"clientFetches"`
	OriginFetch  map[string]OriginFetch `json:"originFetches"`
	Events       []Event                `json:"events"`
}

// CompleteFetches returns how many client fetches have finished.
func (r *Result) CompleteFetches() int {
	n := 0
	for _, f := range r.ClientFetch {
		if f.Complete {
			n++
		}
	}
	return n
}

// defaultMaxWait bounds a wait when the caller doesn't: long enough for a
// cold edge sync, short enough to never hang a terminal indefinitely.
const defaultMaxWait = 2 * time.Minute

// StreamOptions controls how long WaitForResult keeps the stream open.
type StreamOptions struct {
	// WantFetches is the number of completed client fetches that counts as
	// done.
	// Zero means the first result snapshot is enough (useful when the goal
	// is confirming the fiddle synced to the edge).
	WantFetches int
	// MaxWait bounds the whole wait, including the edge-sync phase.
	MaxWait time.Duration
	// Syncing, if set, is called once when the fiddle has been reported as
	// still syncing to the edge for a few seconds, so callers can explain
	// the pause without flashing it on runs that are only briefly delayed.
	Syncing func()
}

// WaitForResult opens the result stream for a session and blocks until the
// wanted number of fetches completed, the server stops early, or MaxWait
// passes.
// It returns the latest snapshot, which on timeout may be partial (or nil,
// if nothing arrived at all).
func (c *Client) WaitForResult(ctx context.Context, sessionID string, opts StreamOptions) (*Result, error) {
	if opts.MaxWait <= 0 {
		opts.MaxWait = defaultMaxWait
	}
	ctx, cancel := context.WithTimeout(ctx, opts.MaxWait)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(c.Endpoint, "/")+"/results/"+sessionID+"/stream", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // #nosec G307
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("result stream returned %s", resp.Status)
	}

	var (
		latest    *Result
		event     string
		data      strings.Builder
		syncTicks int
	)
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 4<<20)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event:"):
			event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			if data.Len() > 0 {
				data.WriteByte('\n')
			}
			data.WriteString(strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		case line == "":
			// Blank line dispatches the accumulated event.
			payload := data.String()
			data.Reset()
			switch event {
			case "waitingForSync":
				// The server emits one of these per second while the
				// config syncs; a couple can appear even when the fiddle
				// is already live, so only a sustained wait counts.
				syncTicks++
				if opts.Syncing != nil && syncTicks == 3 {
					opts.Syncing()
					opts.Syncing = nil
				}
			case "updateResult":
				var r Result
				if err := json.Unmarshal([]byte(payload), &r); err == nil {
					latest = &r
					if r.CompleteFetches() >= opts.WantFetches {
						return latest, nil
					}
				}
			}
			event = ""
		}
	}
	if ctx.Err() != nil {
		// The deadline passed without the wanted fetches completing.
		// A partial snapshot must not masquerade as an answer.
		return latest, fmt.Errorf("timed out after %s waiting for the sandbox: %w", opts.MaxWait, ctx.Err())
	}
	if err := scanner.Err(); err != nil && latest == nil {
		return nil, err
	}
	// Stream ended server-side.
	// Whatever we have is the answer; a nil result means the session
	// produced nothing (the caller may retry with a fresh execution).
	return latest, nil
}

// Run executes the fiddle and waits for the result, retrying with fresh
// sessions when the server closes a stream before anything completed,
// which happens routinely right after a publish.
func (c *Client) Run(ctx context.Context, id string, cacheID int, opts StreamOptions) (*Result, error) {
	if opts.MaxWait <= 0 {
		opts.MaxWait = defaultMaxWait
	}
	deadline := time.Now().Add(opts.MaxWait)
	var lastErr error
	for attempt := 0; ; attempt++ {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("timed out after %s waiting for the sandbox", opts.MaxWait)
		}
		sessionID, err := c.Execute(ctx, id, cacheID)
		if err != nil {
			return nil, err
		}
		attemptOpts := opts
		attemptOpts.MaxWait = remaining
		result, err := c.WaitForResult(ctx, sessionID, attemptOpts)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return nil, err
			}
		} else if result != nil && result.CompleteFetches() >= opts.WantFetches {
			return result, nil
		}
		// Only the first syncing notification should reach the caller.
		if opts.Syncing != nil {
			opts.Syncing = nil
		}
		if err := sleep(ctx, time.Duration(attempt+1)*time.Second); err != nil {
			return nil, err
		}
	}
}
