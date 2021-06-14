package ip_test

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

func TestAllIPs(t *testing.T) {
	var (
		args           = []string{"ip-list", "--token", "123"}
		env            = config.Environment{}
		file           = config.File{}
		configFileName = "/dev/null"
		clientFactory  = mock.APIClient(mock.API{
			AllIPsFn: func() (v4, v6 fastly.IPAddrs, err error) {
				return []string{
						"00.123.45.6/78",
					}, []string{
						"0a12:3b45::/67",
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
	testutil.AssertString(t, "\nIPv4\n\t00.123.45.6/78\n\nIPv6\n\t0a12:3b45::/67\n", out.String())
}
