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

func TestValidate(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		args       []string
		wantError  string
		wantOutput string
	}{
		{
			name:       "success",
			args:       []string{"compute", "validate", "-p", "pkg/package.tar.gz"},
			wantError:  "",
			wantOutput: "Validated package",
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
				clientFactory                  = mock.APIClient(mock.API{})
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = sync.NewWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, buf.String(), testcase.wantOutput)
		})
	}
}
