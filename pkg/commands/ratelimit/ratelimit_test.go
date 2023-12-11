package ratelimit_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestRateLimitCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate CreateERL API error",
			API: mock.API{
				CreateERLFn: func(i *fastly.CreateERLInput) (*fastly.ERL, error) {
					return nil, testutil.Err
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("rate-limit create --name example --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateERL API success",
			API: mock.API{
				CreateERLFn: func(i *fastly.CreateERLInput) (*fastly.ERL, error) {
					return &fastly.ERL{
						Name:          i.Name,
						RateLimiterID: fastly.ToPointer("123"),
					}, nil
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:       args("rate-limit create --name example --service-id 123 --version 3"),
			WantOutput: "Created rate limiter 'example' (123)",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestRateLimitDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate DeleteERL API error",
			API: mock.API{
				DeleteERLFn: func(i *fastly.DeleteERLInput) error {
					return testutil.Err
				},
			},
			Args:      args("rate-limit delete --id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteERL API success",
			API: mock.API{
				DeleteERLFn: func(i *fastly.DeleteERLInput) error {
					return nil
				},
			},
			Args:       args("rate-limit delete --id 123"),
			WantOutput: "SUCCESS: Deleted rate limiter '123'\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestRateLimitDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate GetERL API error",
			API: mock.API{
				GetERLFn: func(i *fastly.GetERLInput) (*fastly.ERL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("rate-limit describe --id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListERL API success",
			API: mock.API{
				GetERLFn: func(i *fastly.GetERLInput) (*fastly.ERL, error) {
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
			Args:       args("rate-limit describe --id 123"),
			WantOutput: "\nAction: response\nClient Key: []\nFeature Revision: 0\nHTTP Methods: []\nID: 123\nLogger Type: \nName: example\nPenalty Box Duration: 20\nResponse: \nResponse Object Name: \nRPS Limit: 10\nService ID: \nURI Dictionary Name: \nVersion: 0\nWindowSize: 60\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestRateLimitList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate ListERL API error",
			API: mock.API{
				ListERLsFn: func(i *fastly.ListERLsInput) ([]*fastly.ERL, error) {
					return nil, testutil.Err
				},
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("rate-limit list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListERL API success",
			API: mock.API{
				ListERLsFn: func(i *fastly.ListERLsInput) ([]*fastly.ERL, error) {
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
			Args:       args("rate-limit list --service-id 123 --version 3"),
			WantOutput: "ID   NAME     ACTION    RPS LIMIT  WINDOW SIZE  PENALTY BOX DURATION\n123  example  response  10         60           20\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TesRateLimittUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate UpdateERL API error",
			API: mock.API{
				UpdateERLFn: func(i *fastly.UpdateERLInput) (*fastly.ERL, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("rate-limit update --id 123 --name example"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateERL API success",
			API: mock.API{
				UpdateERLFn: func(i *fastly.UpdateERLInput) (*fastly.ERL, error) {
					return &fastly.ERL{
						Name:          i.Name,
						RateLimiterID: fastly.ToPointer("123"),
					}, nil
				},
			},
			Args:       args("rate-limit update --id 123 --name example"),
			WantOutput: "Updated rate limiter 'example' (123)",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
