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
	"github.com/fastly/go-fastly/fastly"
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
		wantfiles    map[string]bool
	}{
		{
			name:         "ignore src",
			fastlyignore: "src/*",
			wantfiles: map[string]bool{
				filepath.Join("src/main.rs"): true,
			},
		},
		{
			name:         "ignore cargo files",
			fastlyignore: "Cargo.*",
			wantfiles: map[string]bool{
				"Cargo.lock": true,
				"Cargo.toml": true,
			},
		},
		{
			name:         "ignore all",
			fastlyignore: "*",
			wantfiles: map[string]bool{
				".fastlyignore": true,
				"Cargo.lock":    true,
				"Cargo.toml":    true,
				"src":           true,
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
		ignoredFiles map[string]bool
		wantFiles    []string
	}{
		{
			name:         "no ignored files",
			path:         ".",
			ignoredFiles: map[string]bool{},
			wantFiles: []string{
				"Cargo.lock",
				"Cargo.toml",
				filepath.Join("src/main.rs"),
			},
		},
		{
			name: "one ignored file",
			path: ".",
			ignoredFiles: map[string]bool{
				filepath.Join("src/main.rs"): true,
			},
			wantFiles: []string{
				"Cargo.lock",
				"Cargo.toml",
			},
		},
		{
			name: "multiple ignored files",
			path: ".",
			ignoredFiles: map[string]bool{
				"Cargo.toml": true,
				"Cargo.lock": true,
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

func TestGetIdealPackage(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		inputVersions []*fastly.Version
		wantVersion   int
	}{
		{
			name: "active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "active not latest",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "active and locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")}},
			wantVersion: 2,
		},
		{
			name: "locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "locked not latest",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "no active or locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantVersion: 3,
		},
		{
			name: "not sorted",
			inputVersions: []*fastly.Version{
				{Number: 3, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 4, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-04T01:00:00Z")},
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
			},
			wantVersion: 4,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v, err := getLatestIdealVersion(testcase.inputVersions)
			testutil.AssertNoError(t, err)
			if v.Number != testcase.wantVersion {
				t.Errorf("wanted version %d, got %d", testcase.wantVersion, v.Number)
			}
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
