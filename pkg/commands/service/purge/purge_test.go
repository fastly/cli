package purge_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	purge "github.com/fastly/cli/pkg/commands/service/purge"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPurgeAll(t *testing.T) {
	const (
		testServiceID = "123"
		testStatus    = "ok"
	)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--all",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate --soft flag isn't usable",
			Args:      "--all --service-id " + testServiceID + " --soft",
			WantError: "purge-all requests cannot be done in soft mode (--soft) and will always immediately invalidate all cached content associated with the service",
		},
		{
			Name: "validate PurgeAll API error",
			API: mock.API{
				PurgeAllFn: func(_ context.Context, _ *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--all --service-id " + testServiceID,
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeAll API success",
			API: mock.API{
				PurgeAllFn: func(_ context.Context, _ *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: fastly.ToPointer(testStatus),
					}, nil
				},
			},
			Args:       "--all --service-id " + testServiceID,
			WantOutput: "Purge all status: " + testStatus,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, purge.CommandName}, scenarios)
}

func TestPurgeKeys(t *testing.T) {
	const (
		testServiceID = "123"
		testKey1      = "foo"
		testKey2      = "bar"
		testKey3      = "baz"
		testPurgeID1  = "123"
		testPurgeID2  = "456"
		testPurgeID3  = "789"
	)

	var keys []string
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--file ./testdata/keys",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeKeys API error",
			API: mock.API{
				PurgeKeysFn: func(_ context.Context, _ *fastly.PurgeKeysInput) (map[string]string, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--file ./testdata/keys --service-id " + testServiceID,
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeKeys API success",
			API: mock.API{
				PurgeKeysFn: func(_ context.Context, i *fastly.PurgeKeysInput) (map[string]string, error) {
					// Track the keys parsed
					keys = i.Keys

					return map[string]string{
						testKey1: testPurgeID1,
						testKey2: testPurgeID2,
						testKey3: testPurgeID3,
					}, nil
				},
			},
			Args:       "--file ./testdata/keys --service-id " + testServiceID,
			WantOutput: "KEY  ID\nbar  456\nbaz  789\nfoo  123\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, purge.CommandName}, scenarios)
	assertKeys(keys, t)
}

// assertKeys validates that the --file flag is parsed correctly. It does this
// by ensuring the internal logic has parsed the given file and generated the
// correct []string type.
func assertKeys(keys []string, t *testing.T) {
	want := []string{"foo", "bar", "baz"}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("wanted %s, have %s", want, keys)
	}
}

func TestPurgeKey(t *testing.T) {
	const (
		testServiceID = "123"
		testKey       = "foobar"
		testPurgeID   = "123"
		testStatus    = "ok"
	)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--key " + testKey,
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeKeys API error",
			API: mock.API{
				PurgeKeysFn: func(_ context.Context, _ *fastly.PurgeKeysInput) (map[string]string, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--key " + testKey + " --service-id " + testServiceID,
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeKeys API success",
			API: mock.API{
				PurgeKeysFn: func(_ context.Context, _ *fastly.PurgeKeysInput) (map[string]string, error) {
					return map[string]string{
						testKey: testPurgeID,
					}, nil
				},
			},
			Args:       "--key " + testKey + " --service-id " + testServiceID,
			WantOutput: "Purged key: " + testKey + " (soft: false). Status: " + testStatus + ", ID: " + testPurgeID,
		},
		{
			Name: "validate PurgeKeys API success with soft purge",
			API: mock.API{
				PurgeKeysFn: func(_ context.Context, _ *fastly.PurgeKeysInput) (map[string]string, error) {
					return map[string]string{
						testKey: testPurgeID,
					}, nil
				},
			},
			Args:       "--key " + testKey + " --service-id " + testServiceID + " --soft",
			WantOutput: "Purged key: " + testKey + " (soft: true). Status: " + testStatus + ", ID: " + testPurgeID,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, purge.CommandName}, scenarios)
}

func TestPurgeURL(t *testing.T) {
	const (
		testServiceID = "123"
		testURL       = "https://example.com"
		testPurgeID   = "123"
		testStatus    = "ok"
	)

	scenarios := []testutil.CLIScenario{
		{
			Name: "validate Purge API error",
			API: mock.API{
				PurgeFn: func(_ context.Context, _ *fastly.PurgeInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id " + testServiceID + " --url " + testURL,
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Purge API success",
			API: mock.API{
				PurgeFn: func(_ context.Context, _ *fastly.PurgeInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status:  fastly.ToPointer(testStatus),
						PurgeID: fastly.ToPointer(testPurgeID),
					}, nil
				},
			},
			Args:       "--service-id " + testServiceID + " --url " + testURL,
			WantOutput: "Purged URL: " + testURL + " (soft: false). Status: " + testStatus + ", ID: " + testPurgeID,
		},
		{
			Name: "validate Purge API success with soft purge",
			API: mock.API{
				PurgeFn: func(_ context.Context, _ *fastly.PurgeInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status:  fastly.ToPointer(testStatus),
						PurgeID: fastly.ToPointer(testPurgeID),
					}, nil
				},
			},
			Args:       "--service-id " + testServiceID + " --soft --url " + testURL,
			WantOutput: "Purged URL: " + testURL + " (soft: true). Status: " + testStatus + ", ID: " + testPurgeID,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, purge.CommandName}, scenarios)
}
