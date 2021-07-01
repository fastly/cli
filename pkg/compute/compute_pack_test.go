package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPack(t *testing.T) {
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
			args:     []string{"compute", "pack", "--path", "./main.wasm"},
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
			args:      []string{"compute", "pack"},
			manifest:  `name = "precompiled"`,
			wantError: "error parsing arguments: required flag --path not provided",
		},
		{
			name:      "error no path flag value provided",
			args:      []string{"compute", "pack", "--path", ""},
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

			// Create our test environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makePackEnvironment(t, testcase.manifest)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			err = app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
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

func makePackEnvironment(t *testing.T, manifestContent string) (rootdir string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-pack-*")
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range []string{"main.wasm"} {
		fromFilename := filepath.Join("testdata", "pack", filepath.Join(filename))
		toFilename := filepath.Join(rootdir, filepath.Join(filename))
		testutil.CopyFile(t, fromFilename, toFilename)
	}

	if manifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}
