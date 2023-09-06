package pop_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAllDatacenters(t *testing.T) {
	var stdout bytes.Buffer
	args := testutil.Args("pops")
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
	opts := testutil.NewRunOpts(args, &stdout)
	opts.APIClient = mock.APIClient(api)
	err := app.Run(opts)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, "\nNAME    CODE  GROUP  SHIELD  COORDINATES\nFoobar  FBR   Bar    Baz     {Latitude:1 Longtitude:2 X:3 Y:4}\n", stdout.String())
}
