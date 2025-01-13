package compute_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/fastly/kingpin"
	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
)

// TestFlagDivergencePublish validates that the manually curated list of flags
// within the `compute publish` command doesn't fall out of sync with the
// `compute build` and `compute deploy` commands from which publish is composed.
func TestFlagDivergencePublish(t *testing.T) {
	var g global.Data
	g.Manifest = &manifest.Data{}
	acmd := kingpin.New("foo", "bar")

	rcmd := compute.NewRootCommand(acmd, &g)
	bcmd := compute.NewBuildCommand(rcmd.CmdClause, &g)
	dcmd := compute.NewDeployCommand(rcmd.CmdClause, &g)
	pcmd := compute.NewPublishCommand(rcmd.CmdClause, &g, bcmd, dcmd)

	buildFlags := getFlags(bcmd.CmdClause)
	deployFlags := getFlags(dcmd.CmdClause)
	publishFlags := getFlags(pcmd.CmdClause)

	var (
		expect = make(map[string]int)
		have   = make(map[string]int)
	)

	// Some flags on `compute build` are unique to it.
	// NOTE: There are no flags to ignore but I'm keeping the logic for future.
	ignoreBuildFlags := []string{}

	iter := buildFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreBuildFlags, flag) {
			expect[flag] = 1
		}
	}

	iter = deployFlags.MapRange()
	for iter.Next() {
		expect[iter.Key().String()] = 1
	}

	iter = publishFlags.MapRange()
	for iter.Next() {
		have[iter.Key().String()] = 1
	}

	if !reflect.DeepEqual(expect, have) {
		t.Fatalf("the flags between build/deploy and publish don't match\n\nexpect: %+v\nhave:   %+v\n\n", expect, have)
	}
}

// TestFlagDivergenceServe validates that the manually curated list of flags
// within the `compute serve` command doesn't fall out of sync with the
// `compute build` command as `compute serve` delegates to build.
func TestFlagDivergenceServe(t *testing.T) {
	var cfg global.Data
	acmd := kingpin.New("foo", "bar")

	rcmd := compute.NewRootCommand(acmd, &cfg)
	bcmd := compute.NewBuildCommand(rcmd.CmdClause, &cfg)
	scmd := compute.NewServeCommand(rcmd.CmdClause, &cfg, bcmd)

	buildFlags := getFlags(bcmd.CmdClause)
	serveFlags := getFlags(scmd.CmdClause)

	var (
		expect = make(map[string]int)
		have   = make(map[string]int)
	)

	// Some flags on `compute build` are unique to it.
	// NOTE: There are no flags to ignore but I'm keeping the logic for future.
	ignoreBuildFlags := []string{}

	iter := buildFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreBuildFlags, flag) {
			expect[flag] = 1
		}
	}

	// Some flags on `compute serve` are unique to it.
	// We only want to be sure serve contains all build flags.
	ignoreServeFlags := []string{
		"addr",
		"debug",
		"file",
		"profile-guest",
		"profile-guest-dir",
		"skip-build",
		"viceroy-args",
		"viceroy-check",
		"viceroy-path",
		"watch",
		"watch-dir",
	}

	iter = serveFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreServeFlags, flag) {
			have[flag] = 1
		}
	}

	if !reflect.DeepEqual(expect, have) {
		t.Fatalf("the flags between build and serve don't match\n\nexpect: %+v\nhave:   %+v\n\n", expect, have)
	}
}

// TestFlagDivergenceHashSum validates that the manually curated list of flags
// within the `compute hashsum` command doesn't fall out of sync with the
// `compute build` command as `compute hashsum` delegates to build.
func TestFlagDivergenceHashSum(t *testing.T) {
	var cfg global.Data
	acmd := kingpin.New("foo", "bar")

	rcmd := compute.NewRootCommand(acmd, &cfg)
	bcmd := compute.NewBuildCommand(rcmd.CmdClause, &cfg)
	hcmd := compute.NewHashsumCommand(rcmd.CmdClause, &cfg, bcmd)

	buildFlags := getFlags(bcmd.CmdClause)
	hashsumFlags := getFlags(hcmd.CmdClause)

	var (
		expect = make(map[string]int)
		have   = make(map[string]int)
	)

	// Some flags on `compute build` are unique to it.
	// NOTE: There are no flags to ignore but I'm keeping the logic for future.
	ignoreBuildFlags := []string{}

	iter := buildFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreBuildFlags, flag) {
			expect[flag] = 1
		}
	}

	// Some flags on `compute hashsum` are unique to it.
	// We only want to be sure hashsum contains all build flags.
	ignoreHashsumFlags := []string{
		"package",
		"skip-build",
	}

	iter = hashsumFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreHashsumFlags, flag) {
			have[flag] = 1
		}
	}

	if !reflect.DeepEqual(expect, have) {
		t.Fatalf("the flags between build and hashsum don't match\n\nexpect: %+v\nhave:   %+v\n\n", expect, have)
	}
}

// TestFlagDivergenceHashFiles validates that the manually curated list of flags
// within the `compute hashsum` command doesn't fall out of sync with the
// `compute build` command as `compute hashsum` delegates to build.
func TestFlagDivergenceHashFiles(t *testing.T) {
	var cfg global.Data
	acmd := kingpin.New("foo", "bar")

	rcmd := compute.NewRootCommand(acmd, &cfg)
	bcmd := compute.NewBuildCommand(rcmd.CmdClause, &cfg)
	hcmd := compute.NewHashFilesCommand(rcmd.CmdClause, &cfg, bcmd)

	buildFlags := getFlags(bcmd.CmdClause)
	hashfilesFlags := getFlags(hcmd.CmdClause)

	var (
		expect = make(map[string]int)
		have   = make(map[string]int)
	)

	// Some flags on `compute build` are unique to it.
	// NOTE: There are no flags to ignore but I'm keeping the logic for future.
	ignoreBuildFlags := []string{}

	iter := buildFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreBuildFlags, flag) {
			expect[flag] = 1
		}
	}

	// Some flags on `compute hashsum` are unique to it.
	// We only want to be sure hashsum contains all build flags.
	ignoreHashfilesFlags := []string{
		"package",
		"skip-build",
	}

	iter = hashfilesFlags.MapRange()
	for iter.Next() {
		flag := iter.Key().String()
		if !ignoreFlag(ignoreHashfilesFlags, flag) {
			have[flag] = 1
		}
	}

	if !reflect.DeepEqual(expect, have) {
		t.Fatalf("the flags between build and hash-files don't match\n\nexpect: %+v\nhave:   %+v\n\n", expect, have)
	}
}

// ignoreFlag indicates if needle should be omitted from comparison.
func ignoreFlag(ignore []string, flag string) bool {
	for _, i := range ignore {
		if i == flag {
			return true
		}
	}
	return false
}

func getFlags(cmd *kingpin.CmdClause) reflect.Value {
	return reflect.ValueOf(cmd).Elem().FieldByName("cmdMixin").FieldByName("flagGroup").Elem().FieldByName("long")
}

func TestCreatePackageArchive(t *testing.T) {
	// we're going to chdir to a build environment,
	// so save the pwd to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Copy: []testutil.FileIO{
			{Src: filepath.Join("testdata", "build", "rust", "Cargo.lock"), Dst: "Cargo.lock"},
			{Src: filepath.Join("testdata", "build", "rust", "Cargo.toml"), Dst: "Cargo.toml"},
			{Src: filepath.Join("testdata", "build", "rust", "src", "main.rs"), Dst: filepath.Join("src", "main.rs")},
		},
	})
	defer os.RemoveAll(rootdir)

	// before running the test, chdir into the build environment.
	// when we're done, chdir back to our original location.
	// this is so we can reliably copy the testdata/ fixtures.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(pwd)
	}()

	destination := "cli.tar.gz"

	err = compute.CreatePackageArchive([]string{"Cargo.toml", "Cargo.lock", "src/main.rs"}, destination)
	testutil.AssertNoError(t, err)

	var files, directories []string
	if err := archiver.Walk(destination, func(f archiver.File) error {
		if f.IsDir() {
			directories = append(directories, f.Name())
		} else {
			files = append(files, f.Name())
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	wantDirectories := []string{"cli", "src"}
	testutil.AssertEqual(t, wantDirectories, directories)

	wantFiles := []string{"Cargo.lock", "Cargo.toml", "main.rs"}
	testutil.AssertEqual(t, wantFiles, files)
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
			output := compute.FileNameWithoutExtension(testcase.input)
			testutil.AssertString(t, testcase.wantOutput, output)
		})
	}
}

func TestGetIgnoredFiles(t *testing.T) {
	// we're going to chdir to a build environment,
	// so save the pwd to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Copy: []testutil.FileIO{
			{Src: filepath.Join("testdata", "build", "rust", "Cargo.lock"), Dst: "Cargo.lock"},
			{Src: filepath.Join("testdata", "build", "rust", "Cargo.toml"), Dst: "Cargo.toml"},
			{Src: filepath.Join("testdata", "build", "rust", "src", "main.rs"), Dst: filepath.Join("src", "main.rs")},
		},
	})
	defer os.RemoveAll(rootdir)

	// before running the test, chdir into the build environment.
	// when we're done, chdir back to our original location.
	// this is so we can reliably copy the testdata/ fixtures.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(pwd)
	}()

	for _, testcase := range []struct {
		name         string
		fastlyignore string
		wantfiles    map[string]bool
	}{
		{
			name:         "ignore src",
			fastlyignore: "src/*",
			wantfiles: map[string]bool{
				filepath.Join("src", "main.rs"): true,
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
			if err := os.WriteFile(filepath.Join(rootdir, compute.IgnoreFilePath), []byte(testcase.fastlyignore), 0o600); err != nil {
				t.Fatal(err)
			}
			output, err := compute.GetIgnoredFiles(compute.IgnoreFilePath)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, testcase.wantfiles, output)
		})
	}
}

func TestGetNonIgnoredFiles(t *testing.T) {
	// We're going to chdir to a build environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Copy: []testutil.FileIO{
			{Src: filepath.Join("testdata", "build", "rust", "Cargo.lock"), Dst: "Cargo.lock"},
			{Src: filepath.Join("testdata", "build", "rust", "Cargo.toml"), Dst: "Cargo.toml"},
			{Src: filepath.Join("testdata", "build", "rust", "src", "main.rs"), Dst: filepath.Join("src", "main.rs")},
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
				filepath.Join("src", "main.rs"),
			},
		},
		{
			name: "one ignored file",
			path: ".",
			ignoredFiles: map[string]bool{
				filepath.Join("src", "main.rs"): true,
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
				filepath.Join("src", "main.rs"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			output, err := compute.GetNonIgnoredFiles(testcase.path, testcase.ignoredFiles)
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, testcase.wantFiles, output)
		})
	}
}
