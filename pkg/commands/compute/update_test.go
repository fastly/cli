package compute_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
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
	defer func() {
		_ = os.Chdir(pwd)
	}()

	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "package API error",
			Args: args("compute update -s 123 --version 1 --package pkg/package.tar.gz --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageError,
			},
			WantError: fmt.Sprintf("error uploading package: %s", testutil.Err.Error()),
			WantOutputs: []string{
				"Uploading package",
			},
		},
		{
			Name: "success",
			Args: args("compute update -s 123 --version 2 --package pkg/package.tar.gz --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdatePackageFn: updatePackageOk,
			},
			WantOutputs: []string{
				"Uploading package",
				"Updated package (service 123, version 4)",
			},
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err = app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			for _, s := range testcase.WantOutputs {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}
