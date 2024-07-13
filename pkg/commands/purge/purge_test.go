package purge_test

import (
	"reflect"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	baseCommand = "purge"
)

func TestPurgeAll(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--all",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate --soft flag isn't usable",
			Arg:       "--all --service-id 123 --soft",
			WantError: "purge-all requests cannot be done in soft mode (--soft) and will always immediately invalidate all cached content associated with the service",
		},
		{
			Name: "validate PurgeAll API error",
			API: mock.API{
				PurgeAllFn: func(i *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--all --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeAll API success",
			API: mock.API{
				PurgeAllFn: func(i *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status: fastly.ToPointer("ok"),
					}, nil
				},
			},
			Arg:        "--all --service-id 123",
			WantOutput: "Purge all status: ok",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand}, scenarios)
}

func TestPurgeKeys(t *testing.T) {
	var keys []string
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--file ./testdata/keys",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeKeys API error",
			API: mock.API{
				PurgeKeysFn: func(i *fastly.PurgeKeysInput) (map[string]string, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--file ./testdata/keys --service-id 123",
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
			Arg:        "--file ./testdata/keys --service-id 123",
			WantOutput: "KEY  ID\nbar  456\nbaz  789\nfoo  123\n",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand}, scenarios)
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
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--key foobar",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeKey API error",
			API: mock.API{
				PurgeKeyFn: func(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--key foobar --service-id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate PurgeKey API success",
			API: mock.API{
				PurgeKeyFn: func(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status:  fastly.ToPointer("ok"),
						PurgeID: fastly.ToPointer("123"),
					}, nil
				},
			},
			Arg:        "--key foobar --service-id 123",
			WantOutput: "Purged key: foobar (soft: false). Status: ok, ID: 123",
		},
		{
			Name: "validate PurgeKey API success with soft purge",
			API: mock.API{
				PurgeKeyFn: func(i *fastly.PurgeKeyInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status:  fastly.ToPointer("ok"),
						PurgeID: fastly.ToPointer("123"),
					}, nil
				},
			},
			Arg:        "--key foobar --service-id 123 --soft",
			WantOutput: "Purged key: foobar (soft: true). Status: ok, ID: 123",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand}, scenarios)
}

func TestPurgeURL(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name: "validate Purge API error",
			API: mock.API{
				PurgeFn: func(i *fastly.PurgeInput) (*fastly.Purge, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--service-id 123 --url https://example.com",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Purge API success",
			API: mock.API{
				PurgeFn: func(i *fastly.PurgeInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status:  fastly.ToPointer("ok"),
						PurgeID: fastly.ToPointer("123"),
					}, nil
				},
			},
			Arg:        "--service-id 123 --url https://example.com",
			WantOutput: "Purged URL: https://example.com (soft: false). Status: ok, ID: 123",
		},
		{
			Name: "validate Purge API success with soft purge",
			API: mock.API{
				PurgeFn: func(i *fastly.PurgeInput) (*fastly.Purge, error) {
					return &fastly.Purge{
						Status:  fastly.ToPointer("ok"),
						PurgeID: fastly.ToPointer("123"),
					}, nil
				},
			},
			Arg:        "--service-id 123 --soft --url https://example.com",
			WantOutput: "Purged URL: https://example.com (soft: true). Status: ok, ID: 123",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand}, scenarios)
}
