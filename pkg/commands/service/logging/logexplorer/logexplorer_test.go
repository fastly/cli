package logexplorer_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/commands/service"
	"github.com/fastly/cli/pkg/commands/service/logging"
	"github.com/fastly/cli/pkg/commands/service/logging/logexplorer"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	testServiceID = "123"
	testStart     = "2026-08-12T15:00:00Z"
	testEnd       = "2026-08-13T15:00:00Z"
)

var errLogExplorerTest = errors.New("log explorer test error")

func TestLogExplorer(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --start flag",
			Args:      fmt.Sprintf("--service-id %s --end %s", testServiceID, testEnd),
			WantError: "required flag --start not provided",
		},
		{
			Name:      "validate missing --end flag",
			Args:      fmt.Sprintf("--service-id %s --start %s", testServiceID, testStart),
			WantError: "required flag --end not provided",
		},
		{
			Name: "validate missing service ID",
			Args: fmt.Sprintf("--start %s --end %s", testStart, testEnd),
			EnvVars: map[string]string{
				"FASTLY_SERVICE_ID": "",
			},
			WantError: "error reading service",
		},
		{
			Name: "validate API success absorbs all pages",
			Args: fmt.Sprintf("--service-id %s --start %s --end %s", testServiceID, testStart, testEnd),
			API: &mock.API{
				GetLogRecordsFn: getLogRecordsOK,
			},
			WantOutputs: []string{
				"TIMESTAMP",
				"GET",
				"example.com",
				"/health",
				"200",
				"IAD",
				"/status",
				"204",
				"LHR",
			},
			DontWantOutputs: []string{
				"cursor",
			},
		},
		{
			Name: "validate optional --json flag",
			Args: fmt.Sprintf("--service-id %s --start %s --end %s --json", testServiceID, testStart, testEnd),
			API: &mock.API{
				GetLogRecordsFn: getLogRecordsOK,
			},
			WantOutputs: []string{
				`"request_path": "/health"`,
				`"response_status": 200`,
				`"request_path": "/status"`,
				`"response_status": 204`,
			},
			DontWantOutputs: []string{
				"next_cursor",
			},
		},
		{
			Name: "validate optional request flags",
			Args: fmt.Sprintf(
				"--service-id %s --start %s --end %s --filter response_time,gte,0 --filter response_status,in,200,201",
				testServiceID,
				testStart,
				testEnd,
			),
			API: &mock.API{
				GetLogRecordsFn: getLogRecordsWithOptions,
			},
			WantOutput: "No log records found.",
		},
		{
			Name:      "validate malformed --filter",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --filter response_time,gte", testServiceID, testStart, testEnd),
			API:       &mock.API{},
			WantError: "invalid --filter value",
		},
		{
			Name:      "validate unsupported --filter field",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --filter invalid,gte,0", testServiceID, testStart, testEnd),
			API:       &mock.API{},
			WantError: "field must be one of",
		},
		{
			Name:      "validate unsupported --filter operator",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --filter response_time,invalid,0", testServiceID, testStart, testEnd),
			API:       &mock.API{},
			WantError: "operator must be one of",
		},
		{
			Name: "validate API error",
			Args: fmt.Sprintf("--service-id %s --start %s --end %s", testServiceID, testStart, testEnd),
			API: &mock.API{
				GetLogRecordsFn: getLogRecordsError,
			},
			WantError: errLogExplorerTest.Error(),
		},
		{
			Name:      "validate --verbose and --json are mutually exclusive",
			Args:      fmt.Sprintf("--service-id %s --start %s --end %s --verbose --json", testServiceID, testStart, testEnd),
			WantError: "invalid flag combination, --verbose and --json",
		},
	}

	testutil.RunCLIScenarios(
		t,
		[]string{service.CommandName, logging.CommandName, logexplorer.CommandName},
		scenarios,
	)
}

// getLogRecordsOK simulates a two-page result: the first call (no cursor)
// returns a page with a next cursor, and the second call (with that cursor)
// returns the final page with no further cursor.
func getLogRecordsOK(_ context.Context, input *fastly.GetLogRecordsInput) (*fastly.LogRecordsResponse, error) {
	if input.ServiceID != testServiceID {
		return nil, fmt.Errorf("expected service ID %q, got %q", testServiceID, input.ServiceID)
	}
	if input.Start != testStart {
		return nil, fmt.Errorf("expected start %q, got %q", testStart, input.Start)
	}
	if input.End != testEnd {
		return nil, fmt.Errorf("expected end %q, got %q", testEnd, input.End)
	}
	if input.Limit == nil || *input.Limit != 100 {
		return nil, fmt.Errorf("expected limit 100, got %v", input.Limit)
	}

	if input.NextCursor == nil {
		return &fastly.LogRecordsResponse{
			Data: []*fastly.LogRecord{
				{
					Timestamp:      fastly.ToPointer("2026-08-13T14:30:23Z"),
					RequestMethod:  fastly.ToPointer("GET"),
					RequestHost:    fastly.ToPointer("example.com"),
					RequestPath:    fastly.ToPointer("/health"),
					ResponseStatus: fastly.ToPointer(200),
					FastlyPOP:      fastly.ToPointer("IAD"),
					IsCacheHit:     fastly.ToPointer(true),
					ResponseTime:   fastly.ToPointer(0.093),
				},
			},
			Meta: &fastly.LogExplorerMeta{
				NextCursor: fastly.ToPointer("cursor-2"),
			},
		}, nil
	}

	if *input.NextCursor != "cursor-2" {
		return nil, fmt.Errorf("expected next cursor cursor-2, got %q", *input.NextCursor)
	}

	return &fastly.LogRecordsResponse{
		Data: []*fastly.LogRecord{
			{
				Timestamp:      fastly.ToPointer("2026-08-13T14:35:00Z"),
				RequestMethod:  fastly.ToPointer("GET"),
				RequestHost:    fastly.ToPointer("example.com"),
				RequestPath:    fastly.ToPointer("/status"),
				ResponseStatus: fastly.ToPointer(204),
				FastlyPOP:      fastly.ToPointer("LHR"),
				IsCacheHit:     fastly.ToPointer(false),
				ResponseTime:   fastly.ToPointer(0.05),
			},
		},
	}, nil
}

func getLogRecordsWithOptions(_ context.Context, input *fastly.GetLogRecordsInput) (*fastly.LogRecordsResponse, error) {
	if input.Limit == nil || *input.Limit != 100 {
		return nil, fmt.Errorf("expected limit 100, got %v", input.Limit)
	}
	if input.NextCursor != nil {
		return nil, fmt.Errorf("expected no next cursor on first request, got %v", *input.NextCursor)
	}
	if len(input.Filters) != 2 {
		return nil, fmt.Errorf("expected 2 filters, got %d", len(input.Filters))
	}

	first := input.Filters[0]
	if first.Field != fastly.LogExplorerFilterFieldResponseTime ||
		first.Operator != fastly.LogExplorerFilterOperatorGTE ||
		first.Value != "0" {
		return nil, fmt.Errorf("unexpected first filter: %#v", first)
	}

	second := input.Filters[1]
	if second.Field != fastly.LogExplorerFilterFieldResponseStatus ||
		second.Operator != fastly.LogExplorerFilterOperatorIn ||
		second.Value != "200,201" {
		return nil, fmt.Errorf("unexpected second filter: %#v", second)
	}

	return &fastly.LogRecordsResponse{
		Data: nil,
		Meta: &fastly.LogExplorerMeta{},
	}, nil
}

func getLogRecordsError(_ context.Context, _ *fastly.GetLogRecordsInput) (*fastly.LogRecordsResponse, error) {
	return nil, errLogExplorerTest
}
