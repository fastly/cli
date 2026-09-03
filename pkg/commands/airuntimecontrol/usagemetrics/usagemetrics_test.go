package usagemetrics_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/airuntimecontrol"
	sub "github.com/fastly/cli/pkg/commands/airuntimecontrol/usagemetrics"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/usagemetrics"
)

func TestUsageMetricsList(t *testing.T) {
	metrics := usagemetrics.UsageMetrics{
		Data: []usagemetrics.UsageMetric{
			{
				Date:           "2026-07-29",
				UsageType:      "requests",
				Quantity:       42,
				VirtualKeyID:   "75352aad10d9828b8de7",
				VirtualKeyName: "go-fastly-test-key",
				Provider:       "Anthropic",
				Model:          "claude-sonnet-4-20250514",
			},
		},
		Meta: usagemetrics.Meta{Total: 1},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(metrics))),
					},
				},
			},
			WantOutputs: []string{metrics.Data[0].VirtualKeyName, "requests"},
		},
		{
			Name: "validate optional --json flag",
			Args: "--json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(metrics))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(metrics),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestUsageMetricsExport(t *testing.T) {
	csv := "date,usage_type,quantity\n2026-07-29,requests,42\n"

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate API success",
			Args: "",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader([]byte(csv))),
					},
				},
			},
			WantOutput: csv,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "export"}, scenarios)
}
