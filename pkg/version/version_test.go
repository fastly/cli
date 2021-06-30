package version_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersion(t *testing.T) {
	var buf bytes.Buffer
	args := testutil.Args("version")
	api := mock.API{}
	ara := testutil.NewAppRunArgs(args, api, &buf)
	err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
	testutil.AssertNoError(t, err)
	testutil.AssertString(t, strings.Join([]string{
		"Fastly CLI version v0.0.0-unknown (unknown)",
		"Built with go version unknown",
		"",
	}, "\n"), buf.String())
}
