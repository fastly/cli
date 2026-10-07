package version_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/versions"

	routingconfig "github.com/fastly/cli/pkg/commands/routingconfig"
	root "github.com/fastly/cli/pkg/commands/routingconfig/version"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersionList(t *testing.T) {
	rcid := "routing-config-id"
	vid := "version-id"

	data := []versions.Data{
		{
			VersionID: vid,
			Comment:   "a version",
		},
	}
	apiResp := testutil.GenJSON(versions.Collection{
		Data: data,
		Meta: versions.Meta{Limit: 20, Total: 1},
	})
	wantOutput := testutil.GenJSON(data)

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
						Body:       io.NopCloser(bytes.NewReader(apiResp)),
					},
				},
			},
			WantOutput: string(wantOutput),
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
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "list"}, scenarios)
}

func TestVersionActivate(t *testing.T) {
	rcid := "routing-config-id"
	vid := "version-id"
	name := "my-routing-config"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --version-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --version-id %s", rcid, vid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(routingconfigs.Data{
							Name:            name,
							RoutingConfigID: rcid,
							State:           "active",
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Activated version '%s' for routing config '%s' (routing-config-id: %s)", vid, name, rcid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --version-id %s", rcid, vid),
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
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "activate"}, scenarios)
}

func TestVersionDeleteInactive(t *testing.T) {
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
			WantOutput: fmt.Sprintf("SUCCESS: Deleted inactive versions for routing config (routing-config-id: %s)", rcid),
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
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "delete-inactive"}, scenarios)
}
