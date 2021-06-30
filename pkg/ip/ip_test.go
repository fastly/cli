package ip_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestAllIPs(t *testing.T) {
	var buf bytes.Buffer
	args := []string{"ip-list", "--token", "123"}
	api := mock.API{
		AllIPsFn: func() (v4, v6 fastly.IPAddrs, err error) {
			return []string{
					"00.123.45.6/78",
				}, []string{
					"0a12:3b45::/67",
				}, nil
		},
	}
	ara := testutil.NewAppRunArgs(args, api, &buf)
	err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, "\nIPv4\n\t00.123.45.6/78\n\nIPv6\n\t0a12:3b45::/67\n", buf.String())
}
