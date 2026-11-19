package fiddle

import (
	"fmt"
	"strings"
)

// ExecHost is where a deployed fiddle runs on the edge.
// Requests sent straight to it go through the fiddle's VCL and get the full response back.
//
// Hostnames look like this:
//
//	<session>-<request>-<fiddleID>v<version>-<cacheID>-<n>.<pool>.fiddle.fastly.dev
//
// The session, request, and cache ID parts can be anything.
// The pool can't be guessed: Fiddle picks it, and it only shows up in an execution result.
type ExecHost struct {
	FiddleID string
	Version  int
	// Pool is something like "exec64".
	Pool string
}

// ParseExecHost gets the fiddle ID, version, and pool out of Result.ExecHost.
func ParseExecHost(host string) (*ExecHost, error) {
	labels := strings.SplitN(host, ".", 2)
	if len(labels) != 2 || !strings.HasSuffix(labels[1], ".fiddle.fastly.dev") {
		return nil, fmt.Errorf("unrecognized execution host %q", host)
	}
	pool := strings.TrimSuffix(labels[1], ".fiddle.fastly.dev")
	// Count from the right, since the session and request parts may contain hyphens.
	parts := strings.Split(labels[0], "-")
	if len(parts) < 5 {
		return nil, fmt.Errorf("unrecognized execution host %q", host)
	}
	id, ver, ok := strings.Cut(parts[len(parts)-3], "v")
	if !ok || id == "" {
		return nil, fmt.Errorf("unrecognized execution host %q", host)
	}
	var version int
	if _, err := fmt.Sscanf(ver, "%d", &version); err != nil {
		return nil, fmt.Errorf("unrecognized execution host %q", host)
	}
	return &ExecHost{FiddleID: id, Version: version, Pool: pool}, nil
}

// Hostname returns the host to use for a cache bucket.
// Requests to the same bucket share a cache.
// The session and request parts are arbitrary constants.
func (h *ExecHost) Hostname(cacheID int) string {
	return fmt.Sprintf("00fa571c00-00000000-%sv%d-%d-100.%s.fiddle.fastly.dev",
		h.FiddleID, h.Version, cacheID, h.Pool)
}
