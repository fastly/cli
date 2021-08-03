package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPack(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		name          string
		args          []string
		manifest      string
		wantError     string
		wantOutput    []string
		expectedFiles [][]string
	}{
		// The following test validates that the expected directory struture was
		// created successfully.
		{
			name:     "success",
			args:     args("compute pack --path ./main.wasm"),
			manifest: `name = "precompiled"`,
			wantOutput: []string{
				"Initializing...",
				"Copying wasm binary...",
				"Copying manifest...",
				"Creating .tar.gz file...",
			},
			expectedFiles: [][]string{
				{"pkg", "precompiled", "bin", "main.wasm"},
				{"pkg", "precompiled", "fastly.toml"},
				{"pkg", "precompiled.tar.gz"},
			},
		},
		// The following tests validate that a valid path flag value should be
		// provided.
		{
			name:      "error no path flag",
			args:      args("compute pack"),
			manifest:  `name = "precompiled"`,
			wantError: "error parsing arguments: required flag --path not provided",
		},
		{
			name:      "error no path flag value provided",
			args:      args("compute pack --path "),
			manifest:  `name = "precompiled"`,
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
			defer os.Chdir(pwd)

			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			err = app.Run(opts)
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
