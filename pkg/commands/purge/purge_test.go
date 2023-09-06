package purge_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPurgeAll(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      args("purge --all"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate --soft flag isn't usable",
			Args:      args("purge --all --service-id 123 --soft"),
			WantError: "purge-all requests cannot be done in soft mode (--soft) and will always immediately invalidate all cached content associated with the service",
		},
		{
			Name: "validate PurgeAll API error",
			API: mock.API{
				PurgeAllFn: func(i *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("purge --all --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeAll API success",
			API: mock.API{
				PurgeAllFn: func(i *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: "ok",
					}, nil
				},
			},
			Args:       args("purge --all --service-id 123"),
			WantOutput: "Purge all status: ok",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestPurgeKeys(t *testing.T) {
	var keys []string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      args("purge --file ./testdata/keys"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeKeys API error",
			API: mock.API{
				PurgeKeysFn: func(i *fastly.PurgeKeysInput) (map[string]string, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("purge --file ./testdata/keys --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeKeys API success",
			API: mock.API{
				PurgeKeysFn: func(i *fastly.PurgeKeysInput) (map[string]string, error) {
					// Track the keys parsed
					keys = i.Keys

					return map[string]string{
						"foo": "123",
						"bar": "456",
						"baz": "789",
					}, nil
				},
			},
			Args:       args("purge --file ./testdata/keys --service-id 123"),
			WantOutput: "KEY  ID\nbar  456\nbaz  789\nfoo  123\n",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
			assertKeys(testcase.WantError, testcase.Args, keys, t)
		})
	}
}

// assertKeys validates that the --file flag is parsed correctly. It does this
// by ensuring the internal logic has parsed the given file and generated the
// correct []string type.
func assertKeys(wantError string, args []string, keys []string, t *testing.T) {
	if wantError == "" {
		for _, a := range args {
			if a == "--file" {
				want := []string{"foo", "bar", "baz"}
				if !reflect.DeepEqual(keys, want) {
					t.Errorf("wanted %s, have %s", want, keys)
				}
				break
			}
		}
	}
}

func TestPurgeKey(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      args("purge --key foobar"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeKey API error",
			API: mock.API{
				PurgeKeyFn: func(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("purge --key foobar --service-id 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeKey API success",
			API: mock.API{
				PurgeKeyFn: func(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: "ok",
						ID:     "123",
					}, nil
				},
			},
			Args:       args("purge --key foobar --service-id 123"),
			WantOutput: "Purged key: foobar (soft: false). Status: ok, ID: 123",
		},
		{
			Name: "validate PurgeKey API success with soft purge",
			API: mock.API{
				PurgeKeyFn: func(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: "ok",
						ID:     "123",
					}, nil
				},
			},
			Args:       args("purge --key foobar --service-id 123 --soft"),
			WantOutput: "Purged key: foobar (soft: true). Status: ok, ID: 123",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestPurgeURL(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "validate Purge API error",
			API: mock.API{
				PurgeFn: func(i *fastly.PurgeInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("purge --service-id 123 --url https://example.com"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Purge API success",
			API: mock.API{
				PurgeFn: func(i *fastly.PurgeInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: "ok",
						ID:     "123",
					}, nil
				},
			},
			Args:       args("purge --service-id 123 --url https://example.com"),
			WantOutput: "Purged URL: https://example.com (soft: false). Status: ok, ID: 123",
		},
		{
			Name: "validate Purge API success with soft purge",
			API: mock.API{
				PurgeFn: func(i *fastly.PurgeInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: "ok",
						ID:     "123",
					}, nil
				},
			},
			Args:       args("purge --service-id 123 --soft --url https://example.com"),
			WantOutput: "Purged URL: https://example.com (soft: true). Status: ok, ID: 123",
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
