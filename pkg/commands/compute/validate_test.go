package compute_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestValidate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:       "success",
			Args:       args("compute validate --package pkg/package.tar.gz"),
			WantError:  "",
			WantOutput: "Validated package",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
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

			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return testutil.MockGlobalData(testcase.Args, &stdout), nil
			}
			err = app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
