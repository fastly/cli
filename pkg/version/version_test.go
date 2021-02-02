package version_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
)

func TestVersion(t *testing.T) {
	var (
		args                            = []string{"version"}
		env                             = config.Environment{}
		file                            = config.File{}
		configFileName                  = "/dev/null"
		clientFactory                   = mock.APIClient(mock.API{})
		httpClient     api.HTTPClient   = nil
		versioner      update.Versioner = mock.Versioner{Version: "v1.2.3"}
		in             io.Reader        = nil
		out            bytes.Buffer
	)
	err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, strings.Join([]string{
		"Fastly CLI version v0.0.0-unknown (unknown)",
		"Built with go version unknown",
		"",
	}, "\n"), out.String())
}
