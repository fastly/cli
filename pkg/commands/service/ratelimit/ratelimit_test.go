package ratelimit_test

import (
	"context"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/ratelimit"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestRateLimitCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate CreateERL API error",
			API: mock.API{
				CreateERLFn: func(_ context.Context, _ *fastly.CreateERLInput) (*fastly.ERL, error) {
					return nil, testutil.Err
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--name example --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateERL API success",
			API: mock.API{
				CreateERLFn: func(_ context.Context, i *fastly.CreateERLInput) (*fastly.ERL, error) {
					return &fastly.ERL{
						Name:          i.Name,
						RateLimiterID: fastly.ToPointer("123"),
					}, nil
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:       "--name example --service-id 123 --version 3",
			WantOutput: "Created rate limiter 'example' (123)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestRateLimitDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate DeleteERL API error",
			API: mock.API{
				DeleteERLFn: func(_ context.Context, _ *fastly.DeleteERLInput) error {
					return testutil.Err
				},
			},
			Args:      "--id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteERL API success",
			API: mock.API{
				DeleteERLFn: func(_ context.Context, _ *fastly.DeleteERLInput) error {
					return nil
				},
			},
			Args:       "--id 123",
			WantOutput: "SUCCESS: Deleted rate limiter '123'\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestRateLimitDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate GetERL API error",
			API: mock.API{
				GetERLFn: func(_ context.Context, _ *fastly.GetERLInput) (*fastly.ERL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListERL API success",
			API: mock.API{
				GetERLFn: func(_ context.Context, _ *fastly.GetERLInput) (*fastly.ERL, error) {
					return &fastly.ERL{
						RateLimiterID:      fastly.ToPointer("123"),
						Name:               fastly.ToPointer("example"),
						Action:             fastly.ToPointer(fastly.ERLActionResponse),
						RpsLimit:           fastly.ToPointer(10),
						WindowSize:         fastly.ToPointer(fastly.ERLSize60),
						PenaltyBoxDuration: fastly.ToPointer(20),
					}, nil
				},
			},
			Args:       "--id 123",
			WantOutput: "\nAction: response\nClient Key: []\nFeature Revision: 0\nHTTP Methods: []\nID: 123\nLogger Type: \nName: example\nPenalty Box Duration: 20\nResponse: \nResponse Object Name: \nRPS Limit: 10\nService ID: \nURI Dictionary Name: \nVersion: 0\nWindowSize: 60\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestRateLimitList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate ListERL API error",
			API: mock.API{
				ListERLsFn: func(_ context.Context, _ *fastly.ListERLsInput) ([]*fastly.ERL, error) {
					return nil, testutil.Err
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListERL API success",
			API: mock.API{
				ListERLsFn: func(_ context.Context, _ *fastly.ListERLsInput) ([]*fastly.ERL, error) {
					return []*fastly.ERL{
						{
							RateLimiterID:      fastly.ToPointer("123"),
							Name:               fastly.ToPointer("example"),
							Action:             fastly.ToPointer(fastly.ERLActionResponse),
							RpsLimit:           fastly.ToPointer(10),
							WindowSize:         fastly.ToPointer(fastly.ERLSize60),
							PenaltyBoxDuration: fastly.ToPointer(20),
						},
					}, nil
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "ID   NAME     ACTION    RPS LIMIT  WINDOW SIZE  PENALTY BOX DURATION\n123  example  response  10         60           20\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TesRateLimittUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate UpdateERL API error",
			API: mock.API{
				UpdateERLFn: func(_ context.Context, _ *fastly.UpdateERLInput) (*fastly.ERL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id 123 --name example",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateERL API success",
			API: mock.API{
				UpdateERLFn: func(_ context.Context, i *fastly.UpdateERLInput) (*fastly.ERL, error) {
					return &fastly.ERL{
						Name:          i.Name,
						RateLimiterID: fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       "--id 123 --name example",
			WantOutput: "Updated rate limiter 'example' (123)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
