package ip_test

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

func TestAllIPs(t *testing.T) {
	var stdout bytes.Buffer
	args := testutil.Args("ip-list")
	api := mock.API{
		AllIPsFn: func() (v4, v6 fastly.IPAddrs, err error) {
			return []string{
					"00.123.45.6/78",
				}, []string{
					"0a12:3b45::/67",
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
	testutil.AssertString(t, "\nIPv4\n\t00.123.45.6/78\n\nIPv6\n\t0a12:3b45::/67\n", stdout.String())
}
