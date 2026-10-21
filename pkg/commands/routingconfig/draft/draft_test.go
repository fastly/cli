package draft_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	sdkdraft "github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/draft"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"

	routingconfig "github.com/fastly/cli/pkg/commands/routingconfig"
	root "github.com/fastly/cli/pkg/commands/routingconfig/draft"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDraftDiff(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	p := "/fiesta"

	resp := testutil.GenJSON(sdkdraft.Diff{
		Added: []sdkdraft.PathChange{
			{
				Path: &paths.Data{
					Path:   p,
					PathID: pid,
				},
			},
		},
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --routing-config-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --json", rcid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: string(resp),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s", rcid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(resp)),
					},
				},
			},
			WantOutput: fmt.Sprintf("Added path '%s' (path-id: %s)", p, pid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s", rcid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "diff"}, scenarios)
}

func TestDraftUpdate(t *testing.T) {
	rcid := "routing-config-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --comment %s", rcid, "testing"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(sdkdraft.Data{
							Comment:   "testing",
							VersionID: "version-id",
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated draft comment for routing config (routing-config-id: %s)", rcid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --comment %s", rcid, "testing"),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "update"}, scenarios)
}

func TestDraftDelete(t *testing.T) {
	rcid := "routing-config-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --routing-config-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s", rcid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Discarded draft version for routing config (routing-config-id: %s)", rcid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s", rcid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "delete"}, scenarios)
}
