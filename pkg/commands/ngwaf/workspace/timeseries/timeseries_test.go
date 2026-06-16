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
	gots "github.com/fastly/go-fastly/v15/fastly/ngwaf/v1/workspaces/timeseries"
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
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, workspace.CommandName, sub.CommandName, "get"}, scenarios)
}

var zeroTimeseriesString = strings.TrimSpace(`
Total: 0
`) + "\n"
