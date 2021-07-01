package compute_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestUpdate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		name       string
		args       []string
		api        mock.API
		wantError  string
		wantOutput []string
	}{
		{
			name: "package API error",
			args: args("compute update -s 123 --version 1 -p pkg/package.tar.gz -t 123 --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageError,
			},
			wantError: fmt.Sprintf("error uploading package: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
			},
		},
		{
			name: "success",
			args: args("compute update -s 123 --version 2 -p pkg/package.tar.gz -t 123 --autoclone"),
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

			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)
			err = app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}
