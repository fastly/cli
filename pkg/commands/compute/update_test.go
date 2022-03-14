// NOTE: We always pass the --token flag as this allows us to side-step the
// browser based authentication flow. This is because if a token is explicitly
// provided, then we respect the user knows what they're doing.
package compute_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestUpdate(t *testing.T) {
	// We're going to chdir to a deploy environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Copy: []testutil.FileIO{
			{
				Src: filepath.Join("testdata", "deploy", "pkg", "package.tar.gz"),
				Dst: filepath.Join("pkg", "package.tar.gz"),
			},
		},
	})
	defer os.RemoveAll(rootdir)

	// Before running the test, chdir into the build environment.
	// When we're done, chdir back to our original location.
	// This is so we can reliably copy the testdata/ fixtures.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)

	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "package API error",
			Args: args("compute update -s 123 --version 1 --package pkg/package.tar.gz --autoclone --token 123"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageError,
			},
			WantError: fmt.Sprintf("error uploading package: %s", testutil.Err.Error()),
			WantOutputs: []string{
				"Initializing...",
				"Uploading package...",
			},
		},
		{
			Name: "success",
			Args: args("compute update -s 123 --version 2 --package pkg/package.tar.gz --autoclone --token 123"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageOk,
			},
			WantOutputs: []string{
				"Initializing...",
				"Uploading package...",
				"Updated package (service 123, version 4)",
			},
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.API)
			err = app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			for _, s := range testcase.WantOutputs {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}
