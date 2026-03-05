package stats_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestOriginInspector(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		name       string
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			name: "success table",
			args: args("stats origin-inspector --service-id 123"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsOK,
			},
			wantOutput: "Responses:",
		},
		{
			name: "success json",
			args: args("stats origin-inspector --service-id 123 --format=json"),
			api: mock.API{
				GetOriginMetricsForServiceJSONFn: getOriginMetricsJSONOK,
			},
			wantOutput: "status",
		},
		{
			name: "non-success status",
			args: args("stats origin-inspector --service-id 123"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsNonSuccess,
			},
			wantError: "non-success response",
		},
		{
			name:      "missing service ID",
			args:      args("stats origin-inspector"),
			api:       mock.API{},
			wantError: "error reading service",
		},
		{
			name: "non-success status json",
			args: args("stats origin-inspector --service-id 123 --format=json"),
			api: mock.API{
				GetOriginMetricsForServiceJSONFn: getOriginMetricsJSONNonSuccess,
			},
			wantError: "non-success response",
		},
		{
			name: "api error",
			args: args("stats origin-inspector --service-id 123"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsError,
			},
			wantError: errTest.Error(),
		},
		{
			name: "from RFC3339 maps to Start",
			args: args("stats origin-inspector --service-id 123 --from 2024-01-15T10:00:00Z"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsAssertStart(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)),
			},
			wantOutput: "Responses:",
		},
		{
			name: "from Unix epoch maps to Start",
			args: args("stats origin-inspector --service-id 123 --from 1705312800"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsAssertStart(time.Unix(1705312800, 0)),
			},
			wantOutput: "Responses:",
		},
		{
			name: "to RFC3339 maps to End",
			args: args("stats origin-inspector --service-id 123 --to 2024-01-15T11:00:00Z"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsAssertEnd(time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)),
			},
			wantOutput: "Responses:",
		},
		{
			name: "from invalid format error",
			args: args("stats origin-inspector --service-id 123 --from not-a-time"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsOK,
			},
			wantError: "invalid --from value",
		},
		{
			name: "to invalid format error",
			args: args("stats origin-inspector --service-id 123 --to not-a-time"),
			api: mock.API{
				GetOriginMetricsForServiceFn: getOriginMetricsOK,
			},
			wantError: "invalid --to value",
		},
	}
	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(tc.args, &stdout)
				opts.APIClientFactory = mock.APIClient(tc.api)
				return opts, nil
			}
			err := app.Run(tc.args, nil)
			testutil.AssertErrorContains(t, err, tc.wantError)
			testutil.AssertStringContains(t, stdout.String(), tc.wantOutput)
		})
	}
}

func originMetricsOKResult() (*fastly.OriginInspector, error) {
	return &fastly.OriginInspector{
		Status: fastly.ToPointer("success"),
		Meta: &fastly.OriginMeta{
			Start: fastly.ToPointer("2024-01-15T10:00:00Z"),
			End:   fastly.ToPointer("2024-01-15T11:00:00Z"),
		},
		Data: []*fastly.OriginData{
			{
				Values: []*fastly.OriginMetrics{
					{
						Responses: fastly.ToPointer(uint64(200)),
						Status2xx: fastly.ToPointer(uint64(180)),
						Status4xx: fastly.ToPointer(uint64(15)),
						Status5xx: fastly.ToPointer(uint64(5)),
					},
				},
			},
		},
	}, nil
}

func getOriginMetricsOK(_ context.Context, _ *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
	return originMetricsOKResult()
}

func getOriginMetricsJSONOK(_ context.Context, _ *fastly.GetOriginMetricsInput, dst any) error {
	msg := []byte(`{"status":"success","data":[]}`)
	return json.Unmarshal(msg, dst)
}

func getOriginMetricsJSONNonSuccess(_ context.Context, _ *fastly.GetOriginMetricsInput, dst any) error {
	msg := []byte(`{"status":"error","data":[]}`)
	return json.Unmarshal(msg, dst)
}

func getOriginMetricsNonSuccess(_ context.Context, _ *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
	return &fastly.OriginInspector{
		Status: fastly.ToPointer("error"),
	}, nil
}

func getOriginMetricsAssertStart(want time.Time) func(context.Context, *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
	return func(_ context.Context, i *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
		if i.Start == nil {
			return nil, fmt.Errorf("expected Start to be set, got nil")
		}
		if !i.Start.Equal(want) {
			return nil, fmt.Errorf("expected Start %v, got %v", want, *i.Start)
		}
		return originMetricsOKResult()
	}
}

func getOriginMetricsAssertEnd(want time.Time) func(context.Context, *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
	return func(_ context.Context, i *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
		if i.End == nil {
			return nil, fmt.Errorf("expected End to be set, got nil")
		}
		if !i.End.Equal(want) {
			return nil, fmt.Errorf("expected End %v, got %v", want, *i.End)
		}
		return originMetricsOKResult()
	}
}

func getOriginMetricsError(_ context.Context, _ *fastly.GetOriginMetricsInput) (*fastly.OriginInspector, error) {
	return nil, errTest
}
