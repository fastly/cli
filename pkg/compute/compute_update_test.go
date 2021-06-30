package compute_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/sync"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
)

func TestUpdate(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		args       []string
		api        mock.API
		wantError  string
		wantOutput []string
	}{
		{
			name: "package API error",
			args: []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz", "-t", "123", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageError,
			},
			wantError: "error uploading package: fixture error",
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
			},
		},
		{
			name: "success",
			args: []string{"compute", "update", "-s", "123", "--version", "2", "-p", "pkg/package.tar.gz", "-t", "123", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageOk,
			},
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
				"Updated package (service 123, version 4)",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our deploy environment in a temp dir.
			// Defer a call to clean it up.
			// TODO: abstraction needed for creating environments.
			rootdir := makeDeployEnvironment(t, "")
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = sync.NewWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
			}
		})
	}
}
