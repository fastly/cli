package insights_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/commands/service"
	"github.com/fastly/cli/pkg/commands/service/logging"
	"github.com/fastly/cli/pkg/commands/service/logging/insights"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	testServiceID = "123"
	testStart     = "2026-08-12T15:00:00Z"
	testEnd       = "2026-08-13T15:00:00Z"
)

var errLogInsightsTest = errors.New("log insights test error")

func TestLogInsights(t *testing.T) {
	const visualization = "top-url-by-requests"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --start flag",
			Args:      fmt.Sprintf("--service-id %s --end %s --visualization %s", testServiceID, testEnd, visualization),
			WantError: "required flag --start not provided",
		},
		{
			Name:      "validate missing --end flag",
			Args:      fmt.Sprintf("--service-id %s --start %s --visualization %s", testServiceID, testStart, visualization),
			WantError: "required flag --end not provided",
		},
		{
			Name:      "validate missing --visualization flag",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s", testServiceID, testStart, testEnd),
			WantError: "required flag --visualization not provided",
		},
		{
			Name:      "validate invalid --visualization value",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --visualization invalid", testServiceID, testStart, testEnd),
			WantError: "enum value must be one of",
		},
		{
			Name: "validate missing service ID",
			Args: fmt.Sprintf("--start %s --end %s --visualization %s", testStart, testEnd, visualization),
			EnvVars: map[string]string{
				"FASTLY_SERVICE_ID": "",
			},
			WantError: "error reading service",
		},
		{
			Name: "validate API success",
			Args: fmt.Sprintf("--service-id %s --start %s --end %s --visualization %s", testServiceID, testStart, testEnd, visualization),
			API: &mock.API{
				GetLogInsightsFn: getLogInsightsOK,
			},
			WantOutputs: []string{
				"DIMENSIONS",
				"VALUES",
				"url=GET /health",
				"request_percentage=0.5161290322580645",
			},
		},
		{
			Name: "validate status code dimension output",
			Args: fmt.Sprintf(
				"--service-id %s --start %s --end %s --visualization response-status-codes",
				testServiceID,
				testStart,
				testEnd,
			),
			API: &mock.API{
				GetLogInsightsFn: getLogInsightsStatusCode,
			},
			WantOutput: "status-code=200",
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --start %s --end %s --visualization %s --json", testServiceID, testStart, testEnd, visualization),
			API: &mock.API{
				GetLogInsightsFn: getLogInsightsOK,
			},
			WantOutputs: []string{
				`"url": "GET /health"`,
				`"request_percentage": 0.5161290322580645`,
				`"service_id": "123"`,
			},
		},
		{
			Name:      "validate invalid --domain-exact-match value",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --visualization %s --domain-exact-match invalid", testServiceID, testStart, testEnd, visualization),
			WantError: "'domain-exact-match' flag must be one of the following [true, false]",
		},
		{
			Name: "validate optional request flags",
			Args: fmt.Sprintf(
				"--service-id %s --start %s --end %s --visualization %s --domain example.com --domain-exact-match=false --limit 5 --pops IAD,DFW",
				testServiceID,
				testStart,
				testEnd,
				visualization,
			),
			API: &mock.API{
				GetLogInsightsFn: getLogInsightsWithOptions,
			},
			WantOutput: "No log insights found.",
		},
		{
			Name: "validate optional request flags with JSON output",
			Args: fmt.Sprintf(
				"--service-id %s --start %s --end %s --visualization %s --domain example.com --domain-exact-match=false --limit 5 --pops IAD,DFW --json",
				testServiceID,
				testStart,
				testEnd,
				visualization,
			),
			API: &mock.API{
				GetLogInsightsFn: getLogInsightsWithOptions,
			},
			WantOutputs: []string{
				`"domain": "example.com"`,
				`"domain_exact_match": false`,
				`"pops": [`,
				`"IAD"`,
				`"DFW"`,
			},
		},
		{
			Name: "validate API error",
			Args: fmt.Sprintf("--service-id %s --start %s --end %s --visualization %s", testServiceID, testStart, testEnd, visualization),
			API: &mock.API{
				GetLogInsightsFn: getLogInsightsError,
			},
			WantError: errLogInsightsTest.Error(),
		},
		{
			Name:      "validate --verbose and --json are mutually exclusive",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --visualization %s --verbose --json", testServiceID, testStart, testEnd, visualization),
			WantError: "invalid flag combination, --verbose and --json",
		},
	}

	testutil.RunCLIScenarios(
		t,
		[]string{service.CommandName, logging.CommandName, insights.CommandName},
		scenarios,
	)
}

func getLogInsightsOK(_ context.Context, input *fastly.GetLogInsightsInput) (*fastly.LogInsightsResponse, error) {
	if input.ServiceID != testServiceID {
		return nil, fmt.Errorf("expected service ID %q, got %q", testServiceID, input.ServiceID)
	}
	if input.Start != testStart {
		return nil, fmt.Errorf("expected start %q, got %q", testStart, input.Start)
	}
	if input.End != testEnd {
		return nil, fmt.Errorf("expected end %q, got %q", testEnd, input.End)
	}
	if input.Visualization != fastly.LogInsightsVisualizationTopURLByRequests {
		return nil, fmt.Errorf("unexpected visualization %q", input.Visualization)
	}

	return &fastly.LogInsightsResponse{
		Data: []*fastly.LogInsightsData{
			{
				Dimensions: &fastly.LogInsightsDimensions{
					URL: fastly.ToPointer("GET /health"),
				},
				Values: []*fastly.LogInsightsValue{
					{
						RequestPercentage: fastly.ToPointer(0.5161290322580645),
					},
				},
			},
		},
		Meta: &fastly.LogInsightsMeta{
			Filters: &fastly.LogInsightsFilters{
				ServiceID:        fastly.ToPointer(testServiceID),
				Start:            fastly.ToPointer(testStart),
				End:              fastly.ToPointer(testEnd),
				DomainExactMatch: fastly.ToPointer(true),
				Limit:            fastly.ToPointer(10),
			},
		},
	}, nil
}

func getLogInsightsStatusCode(_ context.Context, input *fastly.GetLogInsightsInput) (*fastly.LogInsightsResponse, error) {
	if input.Visualization != fastly.LogInsightsVisualizationResponseStatusCodes {
		return nil, fmt.Errorf("unexpected visualization %q", input.Visualization)
	}

	return &fastly.LogInsightsResponse{
		Data: []*fastly.LogInsightsData{
			{
				Dimensions: &fastly.LogInsightsDimensions{
					StatusCode: fastly.ToPointer("200"),
				},
				Values: []*fastly.LogInsightsValue{
					{
						Rate: fastly.ToPointer(1.0),
					},
				},
			},
		},
	}, nil
}

func getLogInsightsWithOptions(_ context.Context, input *fastly.GetLogInsightsInput) (*fastly.LogInsightsResponse, error) {
	if input.Domain == nil || *input.Domain != "example.com" {
		return nil, fmt.Errorf("expected domain example.com, got %v", input.Domain)
	}
	if input.DomainExactMatch == nil || *input.DomainExactMatch {
		return nil, fmt.Errorf("expected domain exact match false, got %v", input.DomainExactMatch)
	}
	if input.Limit == nil || *input.Limit != 5 {
		return nil, fmt.Errorf("expected limit 5, got %v", input.Limit)
	}
	if len(input.POPs) != 2 || input.POPs[0] != "IAD" || input.POPs[1] != "DFW" {
		return nil, fmt.Errorf("expected POPs [IAD DFW], got %v", input.POPs)
	}

	return &fastly.LogInsightsResponse{
		Data: []*fastly.LogInsightsData{},
		Meta: &fastly.LogInsightsMeta{
			Filters: &fastly.LogInsightsFilters{
				Domain:           fastly.ToPointer("example.com"),
				DomainExactMatch: fastly.ToPointer(false),
				End:              fastly.ToPointer(testEnd),
				Limit:            fastly.ToPointer(5),
				POPs:             []string{"IAD", "DFW"},
				ServiceID:        fastly.ToPointer(testServiceID),
				Start:            fastly.ToPointer(testStart),
			},
		},
	}, nil
}

func getLogInsightsError(_ context.Context, _ *fastly.GetLogInsightsInput) (*fastly.LogInsightsResponse, error) {
	return nil, errLogInsightsTest
}
