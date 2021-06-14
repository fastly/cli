package pop_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestAllDatacenters(t *testing.T) {
	var (
		args           = []string{"pops", "--token", "123"}
		env            = config.Environment{}
		file           = config.File{}
		configFileName = "/dev/null"
		clientFactory  = mock.APIClient(mock.API{
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
		})
		httpClient   api.HTTPClient   = nil
		cliVersioner update.Versioner = mock.Versioner{Version: "v1.2.3"}
		in           io.Reader        = nil
		out          bytes.Buffer
	)
	err := app.Run(args, env, file, configFileName, clientFactory, httpClient, cliVersioner, in, &out)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, "\nNAME    CODE  GROUP  SHIELD  COORDINATES\nFoobar  FBR   Bar    Baz     {Latitude:1 Longtitude:2 X:3 Y:4}\n", out.String())
}
