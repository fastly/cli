package fiddle

import (
	"fmt"
	"strings"
)

// ExecHost is the hostname of a deployed fiddle on its edge execution pool.
// Requests sent directly to this host run the fiddle's VCL and return the
// complete response, exactly like requests declared inside the fiddle.
//
// The observed format is
//
//	<session>-<request>-<fiddleID>v<version>-<cacheID>-<n>.<pool>.fiddle.fastly.dev
//
// where the session, request, and cacheID segments are free-form (the edge
// derives the fiddle and cache bucket from the name itself), but the pool is
// fixed: a fiddle only syncs to the execution pool the service assigned it,
// which is only learnable from an execution result.
type ExecHost struct {
	FiddleID string
	Version  int
	// Pool is e.g. "exec64", including the numeric suffix.
	Pool string
}

// ParseExecHost extracts the stable parts from an execHost reported by a
// result stream.
func ParseExecHost(host string) (*ExecHost, error) {
	labels := strings.SplitN(host, ".", 2)
	if len(labels) != 2 || !strings.HasSuffix(labels[1], ".fiddle.fastly.dev") {
		return nil, fmt.Errorf("unrecognized execution host %q", host)
	}
	pool := strings.TrimSuffix(labels[1], ".fiddle.fastly.dev")
	// Anchor on the right: the session and request segments are free-form
	// and could themselves contain hyphens, but the label always ends with
	// <fiddleID>v<version>-<cacheID>-<n>.
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

// Hostname renders the host for a given cache bucket; requests sharing a
// bucket share cache state.
// The session and request segments are protocol constants of our choosing.
func (h *ExecHost) Hostname(cacheID int) string {
	return fmt.Sprintf("00fa571c00-00000000-%sv%d-%d-100.%s.fiddle.fastly.dev",
		h.FiddleID, h.Version, cacheID, h.Pool)
}
