package timeseries_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/ngwaf"
	sub "github.com/fastly/cli/pkg/commands/ngwaf/timeseries"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
	gots "github.com/fastly/go-fastly/v15/fastly/ngwaf/v1/timeseries"
)

var timeseriesResponse = gots.Timeseries{
	Data: []gots.DataPoint{
		{
			Dimensions: gots.Dimensions{Time: "2026-06-15T11:00:00Z"},
			Values:     []map[string]any{{"XSS": float64(0), "SQLI": float64(0)}},
		},
		{
			Dimensions: gots.Dimensions{Time: "2026-06-15T12:00:00Z"},
			Values:     []map[string]any{{"XSS": float64(1), "SQLI": float64(0)}},
		},
	},
	Meta: gots.MetaTimeseries{Total: 2},
}

func TestTimeseriesList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --from flag",
			Args:      "--metrics XSS",
			WantError: "error parsing arguments: required flag --from not provided",
		},
		{
			Name:      "validate missing --metrics flag",
			Args:      "--from 2026-06-15T11:00:00Z",
			WantError: "error parsing arguments: required flag --metrics not provided",
		},
		{
			Name: "validate bad --from value",
			Args: "--from not-a-valid-date --metrics XSS",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
						Status:     http.StatusText(http.StatusBadRequest),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(`
							{
								"title": "The request is not valid ('from' is required and must be a valid RFC3339 date).",
								"status": 400
							}
						`))),
					},
				},
			},
			WantError: "400 - Bad Request",
		},
		{
			Name: "validate internal server error",
			Args: "--from 2026-06-15T11:00:00Z --metrics XSS",
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
			Name: "validate API success",
			Args: "--from 2026-06-15T11:00:00Z --metrics XSS,SQLI",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(timeseriesResponse))),
					},
				},
			},
			WantOutput: listTimeseriesString,
		},
		{
			Name: "validate optional --json flag",
			Args: "--from 2026-06-15T11:00:00Z --metrics XSS,SQLI --json",
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
			Name:      "validate --verbose and --json are mutually exclusive",
			Args:      "--from 2026-06-15T11:00:00Z --metrics XSS --verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate API success with zero results",
			Args: "--from 2026-06-15T11:00:00Z --metrics XSS",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(gots.Timeseries{
							Data: []gots.DataPoint{},
							Meta: gots.MetaTimeseries{Total: 0},
						}))),
					},
				},
			},
			WantOutput: "Total: 0\n",
		},
		{
			Name: "validate optional flags --to --granularity --dimensions",
			Args: "--from 2026-06-15T11:00:00Z --metrics XSS,SQLI --to 2026-06-15T12:00:00Z --granularity 60 --dimensions workspaces,time",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(timeseriesResponse))),
					},
				},
			},
			WantOutput: listTimeseriesString,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

var listTimeseriesString = strings.TrimSpace(`
Time                  Workspace  SQLI  XSS
2026-06-15T11:00:00Z             0     0
2026-06-15T12:00:00Z             0     1

Total: 2
`) + "\n"
