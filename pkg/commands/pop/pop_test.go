package pop_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAllDatacenters(t *testing.T) {
	var stdout bytes.Buffer
	args := testutil.SplitArgs("pops")
	api := mock.API{
		AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
			return []fastly.Datacenter{
				{
					Name:   new("Foobar"),
					Code:   new("FBR"),
					Group:  new("Bar"),
					Shield: new("Baz"),
					Coordinates: &fastly.Coordinates{
						Latitude:  new(float64(1)),
						Longitude: new(float64(2)),
						X:         new(float64(3)),
						Y:         new(float64(4)),
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
	testutil.AssertString(t, "\nNAME    CODE  GROUP  SHIELD  COORDINATES\nFoobar  FBR   Bar    Baz     {Latitude:1 Longitude:2 X:3 Y:4}\n", stdout.String())
}
