package compute_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPack(t *testing.T) {
	args := testutil.SplitArgs
	for _, testcase := range []struct {
		name          string
		args          []string
		manifest      string
		wantError     string
		wantOutput    []string
		expectedFiles [][]string
	}{
		{
			name: "success",
			args: args("compute pack --wasm-binary ./main.wasm"),
			manifest: `
			manifest_version = 2
			name = "mypackagename"`,
			wantOutput: []string{
				"Copying wasm binary",
				"Copying manifest",
				"Creating mypackagename.tar.gz file",
			},
			expectedFiles: [][]string{
				{"pkg", "mypackagename.tar.gz"},
			},
		},
		{
			name:      "no wasm binary path flag",
			args:      args("compute pack"),
			manifest:  `name = "precompiled"`,
			wantError: "error parsing arguments: required flag --wasm-binary not provided",
		},
		{
			name: "no wasm binary path flag value",
			args: args("compute pack --wasm-binary "),
			manifest: `
			manifest_version = 2
			name = "precompiled"`,
			wantError: "error copying wasm binary",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a test environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create test environment
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Copy: []testutil.FileIO{
					{Src: filepath.Join("testdata", "pack", "main.wasm"), Dst: "main.wasm"},
				},
				Write: []testutil.FileIO{
					{Src: testcase.manifest, Dst: manifest.Filename},
				},
			})
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(pwd)
			}()

			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return testutil.MockGlobalData(testcase.args, &stdout), nil
			}
			err = app.Run(testcase.args, nil)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}

			for _, files := range testcase.expectedFiles {
				fpath := filepath.Join(rootdir, filepath.Join(files...))
				_, err = os.Stat(fpath)
				if err != nil {
					t.Fatalf("the specified file is not in the expected location: %v", err)
				}
			}
		})
	}
}
