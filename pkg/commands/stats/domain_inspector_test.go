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

func TestDomainInspector(t *testing.T) {
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
			args: args("stats domain-inspector --service-id 123"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsOK,
			},
			wantOutput: "REQUESTS",
		},
		{
			name: "success json",
			args: args("stats domain-inspector --service-id 123 --format=json"),
			api: mock.API{
				GetDomainMetricsForServiceJSONFn: getDomainMetricsJSONOK,
			},
			wantOutput: "status",
		},
		{
			name: "success json alias",
			args: args("stats domain-inspector --service-id 123 --json"),
			api: mock.API{
				GetDomainMetricsForServiceJSONFn: getDomainMetricsJSONOK,
			},
			wantOutput: "status",
		},
		{
			name:      "verbose json combo",
			args:      args("stats domain-inspector --service-id 123 --json --verbose"),
			api:       mock.API{},
			wantError: "invalid flag combination",
		},
		{
			name: "non-success status",
			args: args("stats domain-inspector --service-id 123"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsNonSuccess,
			},
			wantError: "non-success response",
		},
		{
			name:      "missing service ID",
			args:      args("stats domain-inspector"),
			api:       mock.API{},
			wantError: "error reading service",
		},
		{
			name: "non-success status json",
			args: args("stats domain-inspector --service-id 123 --format=json"),
			api: mock.API{
				GetDomainMetricsForServiceJSONFn: getDomainMetricsJSONNonSuccess,
			},
			wantError: "non-success response",
		},
		{
			name: "api error",
			args: args("stats domain-inspector --service-id 123"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsError,
			},
			wantError: errTest.Error(),
		},
		{
			name: "from RFC3339 maps to Start",
			args: args("stats domain-inspector --service-id 123 --from 2024-01-15T10:00:00Z"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsAssertStart(time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)),
			},
			wantOutput: "REQUESTS",
		},
		{
			name: "from Unix epoch maps to Start",
			args: args("stats domain-inspector --service-id 123 --from 1705312800"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsAssertStart(time.Unix(1705312800, 0)),
			},
			wantOutput: "REQUESTS",
		},
		{
			name: "to RFC3339 maps to End",
			args: args("stats domain-inspector --service-id 123 --to 2024-01-15T11:00:00Z"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsAssertEnd(time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)),
			},
			wantOutput: "REQUESTS",
		},
		{
			name: "from invalid format error",
			args: args("stats domain-inspector --service-id 123 --from not-a-time"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsOK,
			},
			wantError: "invalid --from value",
		},
		{
			name: "to invalid format error",
			args: args("stats domain-inspector --service-id 123 --to not-a-time"),
			api: mock.API{
				GetDomainMetricsForServiceFn: getDomainMetricsOK,
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

func domainMetricsOKResult() (*fastly.DomainInspector, error) {
	return &fastly.DomainInspector{
		Status: fastly.ToPointer("success"),
		Meta: &fastly.DomainMeta{
			Start: fastly.ToPointer("2024-01-15T10:00:00Z"),
			End:   fastly.ToPointer("2024-01-15T11:00:00Z"),
		},
		Data: []*fastly.DomainData{
			{
				Values: []*fastly.DomainMetrics{
					{
						Requests:  fastly.ToPointer(uint64(100)),
						Bandwidth: fastly.ToPointer(uint64(5000)),
					},
				},
			},
		},
	}, nil
}

func getDomainMetricsOK(_ context.Context, _ *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
	return domainMetricsOKResult()
}

func getDomainMetricsJSONOK(_ context.Context, _ *fastly.GetDomainMetricsInput, dst any) error {
	msg := []byte(`{"status":"success","data":[]}`)
	return json.Unmarshal(msg, dst)
}

func getDomainMetricsJSONNonSuccess(_ context.Context, _ *fastly.GetDomainMetricsInput, dst any) error {
	msg := []byte(`{"status":"error","data":[]}`)
	return json.Unmarshal(msg, dst)
}

func getDomainMetricsNonSuccess(_ context.Context, _ *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
	return &fastly.DomainInspector{
		Status: fastly.ToPointer("error"),
	}, nil
}

func getDomainMetricsAssertStart(want time.Time) func(context.Context, *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
	return func(_ context.Context, i *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
		if i.Start == nil {
			return nil, fmt.Errorf("expected Start to be set, got nil")
		}
		if !i.Start.Equal(want) {
			return nil, fmt.Errorf("expected Start %v, got %v", want, *i.Start)
		}
		return domainMetricsOKResult()
	}
}

func getDomainMetricsAssertEnd(want time.Time) func(context.Context, *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
	return func(_ context.Context, i *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
		if i.End == nil {
			return nil, fmt.Errorf("expected End to be set, got nil")
		}
		if !i.End.Equal(want) {
			return nil, fmt.Errorf("expected End %v, got %v", want, *i.End)
		}
		return domainMetricsOKResult()
	}
}

func getDomainMetricsError(_ context.Context, _ *fastly.GetDomainMetricsInput) (*fastly.DomainInspector, error) {
	return nil, errTest
}
