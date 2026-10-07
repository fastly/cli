package routingconfig_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs"

	root "github.com/fastly/cli/pkg/commands/routingconfig"
	"github.com/fastly/cli/pkg/testutil"
)

func TestRoutingConfigCreate(t *testing.T) {
	name := "my-routing-config"
	rcid := "routing-config-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: fmt.Sprintf("--name %s", name),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(routingconfigs.Data{
							Name:            name,
							RoutingConfigID: rcid,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created routing config '%s' (routing-config-id: %s)", name, rcid),
		},
		{
			Args: fmt.Sprintf("--name %s", name),
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestRoutingConfigDescribe(t *testing.T) {
	name := "my-routing-config"
	rcid := "routing-config-id"

	resp := testutil.GenJSON(routingconfigs.Data{
		Name:            name,
		RoutingConfigID: rcid,
		State:           "active",
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
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body:       io.NopCloser(strings.NewReader(`{"error": "whoops"}`)),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestRoutingConfigList(t *testing.T) {
	name := "my-routing-config"
	rcid := "routing-config-id"

	data := []routingconfigs.Data{
		{
			Name:            name,
			RoutingConfigID: rcid,
			State:           "active",
		},
	}
	// The HTTP response is a paginated Collection envelope, but the SDK's
	// List() auto-paginates and the CLI's --json output is the flattened
	// []Data slice, not the raw envelope.
	apiResp := testutil.GenJSON(routingconfigs.Collection{
		Data: data,
		Meta: routingconfigs.Meta{Limit: 20, Total: 1},
	})
	wantOutput := testutil.GenJSON(data)

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args: "--json",
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
			Args: "",
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestRoutingConfigDelete(t *testing.T) {
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
			WantOutput: fmt.Sprintf("SUCCESS: Deleted routing config (routing-config-id: %s)", rcid),
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestRoutingConfigActivate(t *testing.T) {
	name := "my-routing-config"
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
			WantOutput: fmt.Sprintf("SUCCESS: Activated routing config '%s' (routing-config-id: %s)", name, rcid),
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, "activate"}, scenarios)
}

func TestRoutingConfigDeactivate(t *testing.T) {
	name := "my-routing-config"
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
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(routingconfigs.Data{
							Name:            name,
							RoutingConfigID: rcid,
							State:           "draft-only",
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Deactivated routing config '%s' (routing-config-id: %s)", name, rcid),
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, "deactivate"}, scenarios)
}
