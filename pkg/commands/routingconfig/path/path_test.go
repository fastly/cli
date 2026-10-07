package path_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths"

	routingconfig "github.com/fastly/cli/pkg/commands/routingconfig"
	root "github.com/fastly/cli/pkg/commands/routingconfig/path"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPathCreate(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	p := "/fiesta"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --path not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path %s", rcid, p),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(paths.Data{
							Path:   p,
							PathID: pid,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Created path '%s' (path-id: %s) in routing config (routing-config-id: %s)", p, pid, rcid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path %s", rcid, p),
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
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "create"}, scenarios)
}

func TestPathDescribe(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	p := "/fiesta"

	resp := testutil.GenJSON(paths.Data{
		Path:   p,
		PathID: pid,
	})

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --path-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --json", rcid, pid),
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
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
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
	testutil.RunCLIScenarios(t, []string{routingconfig.CommandName, root.CommandName, "describe"}, scenarios)
}

func TestPathList(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	p := "/fiesta"

	data := []paths.Data{
		{
			Path:   p,
			PathID: pid,
		},
	}
	apiResp := testutil.GenJSON(paths.Collection{
		Data: data,
		Meta: paths.Meta{Limit: 20, Total: 1},
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

func TestPathUpdate(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"
	p := "/fiesta2"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --path-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --path %s", rcid, pid, p),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(paths.Data{
							Path:   p,
							PathID: pid,
						}))),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Updated path '%s' (path-id: %s)", p, pid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s --path %s", rcid, pid, p),
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

func TestPathDelete(t *testing.T) {
	rcid := "routing-config-id"
	pid := "path-id"

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--routing-config-id %s", rcid),
			WantError: "error parsing arguments: required flag --path-id not provided",
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
					},
				},
			},
			WantOutput: fmt.Sprintf("SUCCESS: Deleted path (path-id: %s)", pid),
		},
		{
			Args: fmt.Sprintf("--routing-config-id %s --path-id %s", rcid, pid),
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
