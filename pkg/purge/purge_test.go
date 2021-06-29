package purge_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestPurgeAll(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing token",
			Args:      args("purge --all"),
			WantError: "no token provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("purge --all --token 123"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate PurgeAll API error",
			API: mock.API{
				PurgeAllFn: func(i *fastly.PurgeAllInput) (*fastly.Purge, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("purge --all --service-id 123 --token 456"),
			WantError: testutil.ErrAPI.Error(),
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
			Args:       args("purge --all --service-id 123 --token 456"),
			WantOutput: "Purge all status: ok",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}
