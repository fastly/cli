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
	// Req and Resp are the raw request and response heads: the first line, then one header per line.
	Req    string `json:"req"`
	Resp   string `json:"resp"`
	Status int    `json:"status"`
	// BodyPreview is roughly the first kilobyte of the body.
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

// Event is one VCL subroutine call reported by the edge.
// The edge also sends timestamps, request IDs, and log lines, which aren't decoded.
type Event struct {
	Type   string `json:"type"`
	FnName string `json:"fnName"`
	Server struct {
		POP    string `json:"pop"`
		NodeID string `json:"nodeID"`
	} `json:"server"`
	Attribs map[string]any `json:"attribs"`
}

// Result is a snapshot of an execution.
// The stream can send several, each more complete than the last.
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

// defaultMaxWait is long enough for a cold edge sync, without leaving a terminal hanging forever.
const defaultMaxWait = 2 * time.Minute

// StreamOptions controls how long WaitForResult keeps the stream open.
type StreamOptions struct {
	// WantFetches is how many finished client fetches count as done.
	// Zero means the first snapshot is enough.
	WantFetches int
	// MaxWait caps the whole wait, edge sync included.
	MaxWait time.Duration
	// Syncing, if set, is called once if the edge is still syncing after a few seconds, so callers can tell the user why it's slow.
	Syncing func()
}

// WaitForResult follows the result stream of a session until enough fetches are done, the server closes the stream, or MaxWait runs out.
// It returns the latest snapshot, which is nil if nothing arrived.
// If it times out or the stream breaks, the error comes with any partial snapshot received.
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
			// A blank line ends the event.
			payload := data.String()
			data.Reset()
			switch event {
			case "waitingForSync":
				// The server sends one of these every second while syncing.
				// A couple can show up even when the fiddle is already live, so wait for three before telling the caller.
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
		// Don't let a partial snapshot pass for a complete answer.
		return latest, fmt.Errorf("timed out after %s waiting for the sandbox: %w", opts.MaxWait, ctx.Err())
	}
	if err := scanner.Err(); err != nil {
		return latest, err
	}
	// The server closed the stream, so what we have is the answer.
	// A nil result means this session produced nothing, and the caller can retry.
	return latest, nil
}

// Run executes the fiddle and waits for the result.
// Right after a publish, the server often closes the stream before anything is done, so Run retries with new sessions.
func (c *Client) Run(ctx context.Context, id string, cacheID int, opts StreamOptions) (*Result, error) {
	if opts.MaxWait <= 0 {
		opts.MaxWait = defaultMaxWait
	}
	deadline := time.Now().Add(opts.MaxWait)
	// MaxWait covers the whole run, including Execute and the pauses between retries.
	runCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()
	var lastErr error
	for attempt := 0; ; attempt++ {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			if lastErr != nil {
				return nil, lastErr
			}
			return nil, fmt.Errorf("timed out after %s waiting for the sandbox", opts.MaxWait)
		}
		sessionID, err := c.Execute(runCtx, id, cacheID)
		if err != nil {
			if ctx.Err() == nil && runCtx.Err() != nil {
				// Report running out of time as a regular timeout, not a raw context error.
				continue
			}
			return nil, err
		}
		attemptOpts := opts
		attemptOpts.MaxWait = remaining
		result, err := c.WaitForResult(runCtx, sessionID, attemptOpts)
		if err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return nil, err
			}
		} else if result != nil && result.CompleteFetches() >= opts.WantFetches {
			return result, nil
		}
		// Only the first attempt may report that the edge is syncing.
		if opts.Syncing != nil {
			opts.Syncing = nil
		}
		if err := sleep(runCtx, time.Duration(attempt+1)*time.Second); err != nil && ctx.Err() != nil {
			return nil, err
		}
	}
}
