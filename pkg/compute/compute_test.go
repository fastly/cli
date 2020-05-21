package compute

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/mholt/archiver/v3"
)

func TestCreatePackageArchive(t *testing.T) {
	for _, testcase := range []struct {
		name            string
		destination     string
		inputFiles      []string
		wantDirectories []string
		wantFiles       []string
	}{
		{
			name:        "success",
			destination: "cli.tar.gz",
			inputFiles: []string{
				"Cargo.toml",
				"Cargo.lock",
				"src/main.rs",
			},
			wantDirectories: []string{
				"cli",
				"src",
			},
			wantFiles: []string{
				"Cargo.lock",
				"Cargo.toml",
				"main.rs",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// we're going to chdir to a build environment,
			// so save the pwd to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// create our build environment in a temp dir.
			// defer a call to clean it up.
			rootdir := makeBuildEnvironment(t, "")
			defer os.RemoveAll(rootdir)

			// before running the test, chdir into the build environment.
			// when we're done, chdir back to our original location.
			// this is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			err = createPackageArchive(testcase.inputFiles, testcase.destination)
			testutil.AssertNoError(t, err)

			var files, directories []string
			if err := archiver.Walk(testcase.destination, func(f archiver.File) error {
				if f.IsDir() {
					directories = append(directories, f.Name())
				} else {
					files = append(files, f.Name())
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}

			testutil.AssertEqual(t, testcase.wantDirectories, directories)
			testutil.AssertEqual(t, testcase.wantFiles, files)
		})
	}
}

func TestFileNameWithoutExtension(t *testing.T) {
	for _, testcase := range []struct {
		input      string
		wantOutput string
	}{
		{
			input:      "foo/bar/baz.tar.gz",
			wantOutput: "baz",
		},
		{
			input:      "foo/bar/baz.wasm",
			wantOutput: "baz",
		},
		{
			input:      "foo.tar",
			wantOutput: "foo",
		},
	} {
		t.Run(testcase.input, func(t *testing.T) {
			output := fileNameWithoutExtension(testcase.input)
			testutil.AssertString(t, testcase.wantOutput, output)
		})
	}
}

func TestGetIgnoredFiles(t *testing.T) {
	for _, testcase := range []struct {
		name         string
		fastlyignore string
		wantfiles    []string
	}{
		{
			name:         "ignore src",
			fastlyignore: "src/*",
			wantfiles: []string{
				filepath.Join("src/main.rs"),
			},
		},
		{
			name:         "ignore cargo files",
			fastlyignore: "Cargo.*",
			wantfiles: []string{
				"Cargo.lock",
				"Cargo.toml",
			},
		},
		{
			name:         "ignore all",
			fastlyignore: "*",
			wantfiles: []string{
				".fastlyignore",
				"Cargo.lock",
				"Cargo.toml",
				"src",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// we're going to chdir to a build environment,
			// so save the pwd to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// create our build environment in a temp dir.
			// defer a call to clean it up.
			rootdir := makeBuildEnvironment(t, testcase.fastlyignore)
			defer os.RemoveAll(rootdir)

			// before running the test, chdir into the build environment.
			// when we're done, chdir back to our original location.
			// this is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			output, err := getIgnoredFiles(IgnoreFilePath)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, testcase.wantfiles, output)
		})
	}
}

func TestGetNonIgnoredFiles(t *testing.T) {
	for _, testcase := range []struct {
		name         string
		path         string
		ignoredFiles []string
		wantFiles    []string
	}{
		{
			name:         "no ignored files",
			path:         ".",
			ignoredFiles: []string{},
			wantFiles: []string{
				"Cargo.lock",
				"Cargo.toml",
				filepath.Join("src/main.rs"),
			},
		},
		{
			name: "one ignored file",
			path: ".",
			ignoredFiles: []string{
				filepath.Join("src/main.rs"),
			},
			wantFiles: []string{
				"Cargo.lock",
				"Cargo.toml",
			},
		},
		{
			name: "multiple ignored files",
			path: ".",
			ignoredFiles: []string{
				"Cargo.toml",
				"Cargo.lock",
			},
			wantFiles: []string{
				filepath.Join("src/main.rs"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a build environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our build environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeBuildEnvironment(t, "")
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			output, err := getNonIgnoredFiles(testcase.path, testcase.ignoredFiles)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, testcase.wantFiles, output)
		})
	}
}

func makeBuildEnvironment(t *testing.T, fastlyIgnoreContent string) (rootdir string) {
	t.Helper()

	p := make([]byte, 8)
	n, err := rand.Read(p)
	if err != nil {
		t.Fatal(err)
	}

	rootdir = filepath.Join(
		os.TempDir(),
		fmt.Sprintf("fastly-build-%x", p[:n]),
	)

	if err := os.MkdirAll(rootdir, 0700); err != nil {
		t.Fatal(err)
	}

	for _, filename := range [][]string{
		{"Cargo.toml"},
		{"Cargo.lock"},
		{"src", "main.rs"},
	} {
		fromFilename := filepath.Join("testdata", "build", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		if err := common.CopyFile(fromFilename, toFilename); err != nil {
			t.Fatal(err)
		}
	}

	if fastlyIgnoreContent != "" {
		filename := filepath.Join(rootdir, IgnoreFilePath)
		if err := ioutil.WriteFile(filename, []byte(fastlyIgnoreContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}
