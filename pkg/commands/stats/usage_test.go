package stats_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestUsage(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		name       string
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			name:       "success plain",
			args:       args("stats usage"),
			api:        mock.API{GetUsageFn: getUsageOK},
			wantOutput: "Region: usa",
		},
		{
			name:       "success json",
			args:       args("stats usage --format=json"),
			api:        mock.API{GetUsageFn: getUsageOK},
			wantOutput: "bandwidth",
		},
		{
			name:       "success by-service",
			args:       args("stats usage --by-service"),
			api:        mock.API{GetUsageByServiceFn: getUsageByServiceOK},
			wantOutput: "Service: svc123",
		},
		{
			name:       "success by-service json",
			args:       args("stats usage --by-service --format=json"),
			api:        mock.API{GetUsageByServiceFn: getUsageByServiceOK},
			wantOutput: "svc123",
		},
		{
			name:       "nil usage entry json",
			args:       args("stats usage --format=json"),
			api:        mock.API{GetUsageFn: getUsageNilEntry},
			wantOutput: `"bandwidth"`,
		},
		{
			name:       "nil usage entry table skipped",
			args:       args("stats usage"),
			api:        mock.API{GetUsageFn: getUsageWithNilEntry},
			wantOutput: "Region: europe",
		},
		{
			name:      "non-success status",
			args:      args("stats usage"),
			api:       mock.API{GetUsageFn: getUsageNonSuccess},
			wantError: "non-success response",
		},
		{
			name:      "api error",
			args:      args("stats usage"),
			api:       mock.API{GetUsageFn: getUsageError},
			wantError: errTest.Error(),
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

func getUsageOK(_ context.Context, _ *fastly.GetUsageInput) (*fastly.UsageResponse, error) {
	return &fastly.UsageResponse{
		Status: fastly.ToPointer("success"),
		Data: &fastly.RegionsUsage{
			"usa": &fastly.Usage{
				Bandwidth:       fastly.ToPointer(uint64(1000)),
				Requests:        fastly.ToPointer(uint64(500)),
				ComputeRequests: fastly.ToPointer(uint64(100)),
			},
		},
	}, nil
}

func getUsageByServiceOK(_ context.Context, _ *fastly.GetUsageInput) (*fastly.UsageByServiceResponse, error) {
	return &fastly.UsageByServiceResponse{
		Status: fastly.ToPointer("success"),
		Data: &fastly.ServicesByRegionsUsage{
			"usa": &fastly.ServicesUsage{
				"svc123": &fastly.Usage{
					Bandwidth:       fastly.ToPointer(uint64(1000)),
					Requests:        fastly.ToPointer(uint64(500)),
					ComputeRequests: fastly.ToPointer(uint64(100)),
				},
			},
		},
	}, nil
}

func getUsageNilEntry(_ context.Context, _ *fastly.GetUsageInput) (*fastly.UsageResponse, error) {
	return &fastly.UsageResponse{
		Status: fastly.ToPointer("success"),
		Data: &fastly.RegionsUsage{
			"empty_region": nil,
		},
	}, nil
}

func getUsageWithNilEntry(_ context.Context, _ *fastly.GetUsageInput) (*fastly.UsageResponse, error) {
	return &fastly.UsageResponse{
		Status: fastly.ToPointer("success"),
		Data: &fastly.RegionsUsage{
			"empty_region": nil,
			"europe": &fastly.Usage{
				Bandwidth:       fastly.ToPointer(uint64(2000)),
				Requests:        fastly.ToPointer(uint64(300)),
				ComputeRequests: fastly.ToPointer(uint64(50)),
			},
		},
	}, nil
}

func getUsageNonSuccess(_ context.Context, _ *fastly.GetUsageInput) (*fastly.UsageResponse, error) {
	return &fastly.UsageResponse{
		Status:  fastly.ToPointer("error"),
		Message: fastly.ToPointer("bad request"),
	}, nil
}

func getUsageError(_ context.Context, _ *fastly.GetUsageInput) (*fastly.UsageResponse, error) {
	return nil, errTest
}
