package pop_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestAllDatacenters(t *testing.T) {
	var buf bytes.Buffer
	args := testutil.Args("pops --token 123")
	api := mock.API{
		AllDatacentersFn: func() ([]fastly.Datacenter, error) {
			return []fastly.Datacenter{
				{
					Name:   "Foobar",
					Code:   "FBR",
					Group:  "Bar",
					Shield: "Baz",
					Coordinates: fastly.Coordinates{
						Latitude:   1,
						Longtitude: 2,
						X:          3,
						Y:          4,
					},
				},
			}, nil
		},
	}
	ara := testutil.NewAppRunArgs(args, api, &buf)
	err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, "\nNAME    CODE  GROUP  SHIELD  COORDINATES\nFoobar  FBR   Bar    Baz     {Latitude:1 Longtitude:2 X:3 Y:4}\n", buf.String())
}
