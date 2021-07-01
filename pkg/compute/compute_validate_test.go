package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/testutil"
)

func TestValidate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		name       string
		args       []string
		wantError  string
		wantOutput string
	}{
		{
			name:       "success",
			args:       args("compute validate -p pkg/package.tar.gz"),
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

			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			err = app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}
