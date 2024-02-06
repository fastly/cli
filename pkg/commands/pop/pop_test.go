package pop_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
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
					Name:   fastly.ToPointer("Foobar"),
					Code:   fastly.ToPointer("FBR"),
					Group:  fastly.ToPointer("Bar"),
					Shield: fastly.ToPointer("Baz"),
					Coordinates: &fastly.Coordinates{
						Latitude:   fastly.ToPointer(float64(1)),
						Longtitude: fastly.ToPointer(float64(2)),
						X:          fastly.ToPointer(float64(3)),
						Y:          fastly.ToPointer(float64(4)),
					},
				},
			}, nil
		},
	}
	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
		opts := testutil.MockGlobalData(args, &stdout)
		opts.APIClientFactory = mock.APIClient(api)
		return opts, nil
	}
	err := app.Run(args, nil)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, "\nNAME    CODE  GROUP  SHIELD  COORDINATES\nFoobar  FBR   Bar    Baz     {Latitude:1 Longtitude:2 X:3 Y:4}\n", stdout.String())
}
