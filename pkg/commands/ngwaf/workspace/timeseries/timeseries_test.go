package timeseries_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	workspace "github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/workspace/timeseries"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	gots "github.com/fastly/go-fastly/v16/fastly/ngwaf/v1/workspaces/timeseries"
)

const (
	workspaceID = "workspaceID"
	from        = "2024-01-01T00:00:00Z"
	metrics     = "requests_total"
)

var timeseriesResponse = gots.TimeSeries{
	Data: []map[string]any{},
	Meta: gots.MetaTimeSeries{
		Total: 0,
	},
}

var timeseriesResponseSample = gots.TimeSeries{
	Data: []map[string]any{
		{"timestamp": "2026-06-22T18:34:00Z", "HTTP404": float64(1), "requests_total": float64(1)},
		{"timestamp": "2026-06-22T18:35:00Z", "HTTP404": float64(0), "requests_total": float64(3)},
	},
	Meta: gots.MetaTimeSeries{
		Total: 2,
	},
}

func TestWorkspaceTimeseriesGet(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --from flag",
			Args:      fmt.Sprintf("--metrics %s --workspace-id %s", metrics, workspaceID),
			WantError: "error parsing arguments: required flag --from not provided",
		},
		{
			Name:      "validate missing --metrics flag",
			Args:      fmt.Sprintf("--from %s --workspace-id %s", from, workspaceID),
			WantError: "error parsing arguments: required flag --metrics not provided",
		},
		{
			Name:      "validate missing --workspace-id flag",
			Args:      fmt.Sprintf("--from %s --metrics %s", from, metrics),
			WantError: "error reading workspace ID: no workspace ID found",
		},
		{
			Name: "validate internal server error",
			Args: fmt.Sprintf("--from %s --metrics %s --workspace-id %s", from, metrics, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
						Status:     http.StatusText(http.StatusInternalServerError),
					},
				},
			},
			WantError: "500 - Internal Server Error",
		},
		{
			Name: "validate API success (zero results)",
			Args: fmt.Sprintf("--from %s --metrics %s --workspace-id %s", from, metrics, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(timeseriesResponse))),
					},
				},
			},
			WantOutput: zeroTimeseriesString,
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--from %s --metrics %s --workspace-id %s --json", from, metrics, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(timeseriesResponse))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(timeseriesResponse),
		},
		{
			Name: "validate API success (sample results)",
			Args: fmt.Sprintf("--from %s --metrics %s --workspace-id %s", from, metrics, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(timeseriesResponseSample))),
					},
				},
			},
			WantOutput: sampleTimeseriesString,
		},
		{
			Name: "validate optional --json flag (sample results)",
			Args: fmt.Sprintf("--from %s --metrics %s --workspace-id %s --json", from, metrics, workspaceID),
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(timeseriesResponseSample))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(timeseriesResponseSample),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "get"}, scenarios)
}

var zeroTimeseriesString = strings.TrimSpace(`
Total: 0
`) + "\n"

var sampleTimeseriesString = strings.TrimSpace(`
Timestamp             HTTP404  requests_total
2026-06-22T18:34:00Z  1        1
2026-06-22T18:35:00Z  0        3

Total: 2
`) + "\n"
