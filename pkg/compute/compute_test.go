package compute

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/fastly/kingpin"
	"github.com/mholt/archiver/v3"
)

// TestPublishFlagDivergence validates that the manually curated list of flags
// within the `compute publish` command doesn't fall out of sync with the
// `compute build` and `compute deploy` commands from which publish is composed.
func TestPublishFlagDivergence(t *testing.T) {
	var cfg config.Data
	acmd := kingpin.New("foo", "bar")

	rcmd := NewRootCommand(acmd, &cfg)
	bcmd := NewBuildCommand(rcmd.CmdClause, client{}, &cfg)
	dcmd := NewDeployCommand(rcmd.CmdClause, client{}, &cfg)
	pcmd := NewPublishCommand(rcmd.CmdClause, &cfg, bcmd, dcmd)

	buildFlags := getFlags(bcmd.CmdClause)
	deployFlags := getFlags(dcmd.CmdClause)
	publishFlags := getFlags(pcmd.CmdClause)

	var (
		expect []string
		have   []string
	)

	iter := buildFlags.MapRange()
	for iter.Next() {
		expect = append(expect, fmt.Sprintf("%s", iter.Key()))
	}
	iter = deployFlags.MapRange()
	for iter.Next() {
		expect = append(expect, fmt.Sprintf("%s", iter.Key()))
	}

	iter = publishFlags.MapRange()
	for iter.Next() {
		have = append(have, fmt.Sprintf("%s", iter.Key()))
	}

	sort.Strings(expect)
	sort.Strings(have)

	errMsg := "the flags between build/deploy and publish don't match"

	if len(expect) != len(have) {
		t.Fatal(errMsg)
	}

	for i, v := range expect {
		if have[i] != v {
			t.Fatalf("%s, expected: %s, got: %s", errMsg, v, have[i])
		}
	}
}

type client struct{}

func (c client) Do(*http.Request) (*http.Response, error) {
	var resp http.Response
	return &resp, nil
}

func getFlags(cmd *kingpin.CmdClause) reflect.Value {
	return reflect.ValueOf(cmd).Elem().FieldByName("cmdMixin").FieldByName("flagGroup").Elem().FieldByName("long")
}

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

func TestGetLatestCrateVersion(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		inputClient api.HTTPClient
		wantVersion *semver.Version
		wantError   string
	}{
		{
			name:        "http error",
			inputClient: &errorClient{errTest},
			wantError:   "fixture error",
		},
		{
			name:        "no valid versions",
			inputClient: &versionClient{[]string{}},
			wantError:   "no valid crate versions found",
		},
		{
			name:        "unsorted",
			inputClient: &versionClient{[]string{"0.5.23", "0.1.0", "1.2.3", "0.7.3"}},
			wantVersion: semver.MustParse("1.2.3"),
		},
		{
			name:        "reverse chronological",
			inputClient: &versionClient{[]string{"1.2.3", "0.8.3", "0.3.2"}},
			wantVersion: semver.MustParse("1.2.3"),
		},
		{
			name:        "contains pre-release",
			inputClient: &versionClient{[]string{"0.2.3", "0.8.3", "0.3.2", "0.9.0-beta.2"}},
			wantVersion: semver.MustParse("0.8.3"),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v, err := getLatestCrateVersion(testcase.inputClient, "fastly")
			testutil.AssertErrorContains(t, err, testcase.wantError)
			if err == nil && !v.Equal(testcase.wantVersion) {
				t.Errorf("wanted version %s, got %s", testcase.wantVersion, v)
			}
		})
	}
}

func TestGetCrateVersionFromMetadata(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		inputLock   CargoMetadata
		inputCrate  string
		wantVersion *semver.Version
		wantError   string
	}{
		{
			name:       "crate not found",
			inputLock:  CargoMetadata{},
			inputCrate: "fastly",
			wantError:  "fastly crate not found",
		},
		{
			name: "parsing error",
			inputLock: CargoMetadata{
				Package: []CargoPackage{
					{
						Name:    "foo",
						Version: "1.2.3",
					},
					{
						Name:    "fastly",
						Version: "dgsfdfg",
					},
				},
			},
			inputCrate: "fastly",
			wantError:  "error parsing cargo metadata",
		},
		{
			name: "success",
			inputLock: CargoMetadata{
				Package: []CargoPackage{
					{
						Name:    "foo",
						Version: "1.2.3",
					},
					{
						Name:    "fastly-sys",
						Version: "0.3.0",
					},
					{
						Name:    "fastly",
						Version: "3.0.0",
					},
				},
			},
			inputCrate:  "fastly",
			wantVersion: semver.MustParse("3.0.0"),
		},
		{
			name: "success nested",
			inputLock: CargoMetadata{
				Package: []CargoPackage{
					{
						Name:    "foo",
						Version: "1.2.3",
					},
					{
						Name:    "fastly",
						Version: "3.0.0",
						Dependencies: []CargoPackage{
							{
								Name:    "fastly-sys",
								Version: "0.3.0",
							},
						},
					},
				},
			},
			inputCrate:  "fastly-sys",
			wantVersion: semver.MustParse("0.3.0"),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v, err := getCrateVersionFromMetadata(testcase.inputLock, testcase.inputCrate)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			if err == nil && !v.Equal(testcase.wantVersion) {
				t.Errorf("wanted version %s, got %s", testcase.wantVersion, v)
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
		if err := filesystem.CopyFile(fromFilename, toFilename); err != nil {
			t.Fatal(err)
		}
	}

	if fastlyIgnoreContent != "" {
		filename := filepath.Join(rootdir, IgnoreFilePath)
		if err := os.WriteFile(filename, []byte(fastlyIgnoreContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}

var errTest = errors.New("fixture error")

type errorClient struct {
	err error
}

func (c errorClient) Do(*http.Request) (*http.Response, error) {
	return nil, c.err
}

type versionClient struct {
	versions []string
}

func (v versionClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()

	var versions []string
	for _, vv := range v.versions {
		versions = append(versions, fmt.Sprintf(`{"num":"%s"}`, vv))
	}

	_, err := rec.Write([]byte(fmt.Sprintf(`{"versions":[%s]}`, strings.Join(versions, ","))))
	if err != nil {
		return nil, err
	}
	return rec.Result(), nil
}
