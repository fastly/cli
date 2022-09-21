package compute_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

// TestBuildRust validates that the rust ecosystem is in place and accurate.
//
// NOTE:
// The defined tests rely on some key pieces of information:
//
// 1. Each test has a 'default' Cargo.toml created for it.
// 2. Each test can override the default Cargo.toml by defining a `cargoManifest`.
//
// You can locate the default Cargo.toml here:
// ./testdata/build/rust/Cargo.toml
func TestBuildRust(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_BUILD_RUST") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	args := testutil.Args

	scenarios := []struct {
		name                 string
		args                 []string
		applicationConfig    config.File
		fastlyManifest       string
		cargoManifest        string
		wantError            string
		wantRemediationError string
		wantOutputContains   string
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading package manifest",
			wantRemediationError: "Run `fastly compute init` to ensure a correctly configured manifest.",
		},
		{
			name: "empty language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"`,
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			language = "rust"`,
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "foobar"`,
			wantError: "unsupported language foobar",
		},
		{
			name: "missing fastly dependency",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"`,
			wantError: "failed to find SDK 'fastly' in the 'Cargo.toml' manifest",
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainConstraint: ">= 1.54.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
		},
		{
			name: "rust toolchain does not match the constraint",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.4.0"`,
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						// NOTE: This test is purposely setting the constraint lower to what
						// is reasonably going to be the current stable release installed to
						// force the test failure scenario to be hit.
						ToolchainConstraint: ">= 1.0.0 < 1.40.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
			wantError:            "toolchain version 1.63.0 didn't meet the constraint >= 1.0.0 < 1.40.0",
			wantRemediationError: "Run `rustup update stable`, or ensure your `rust-toolchain` file specifies a version matching the constraint (e.g. `channel = \"stable\"`).",
		},
		{
			name: "fastly crate prerelease",
			args: args("compute build"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "rust"

      [scripts]
      build = "%s"`, compute.RustDefaultBuildCommand),
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainConstraint: ">= 1.54.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "0.6.0"`,
			wantOutputContains: "Built package 'test'",
		},
		{
			name: "successful build",
			args: args("compute build"),
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainConstraint: ">= 1.54.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "rust"

      [scripts]
      build = "%s"`, compute.RustDefaultBuildCommand),
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			wantOutputContains: "Built package 'test'",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.name, func(t *testing.T) {
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
					{Src: filepath.Join("testdata", "deploy", "pkg", "package.tar.gz"), Dst: filepath.Join("pkg", "package.tar.gz")},
				},
				Write: []testutil.FileIO{
					{Src: testcase.fastlyManifest, Dst: manifest.Filename},
					{Src: testcase.cargoManifest, Dst: "Cargo.toml"},
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
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ConfigFile = testcase.applicationConfig
			err = app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			if testcase.wantOutputContains != "" {
				testutil.AssertStringContains(t, stdout.String(), testcase.wantOutputContains)
			}
		})
	}
}

func TestBuildAssemblyScript(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_BUILD_ASSEMBLYSCRIPT") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD_ASSEMBLYSCRIPT or TEST_COMPUTE_BUILD to run this test")
	}

	for _, testcase := range []struct {
		name                 string
		args                 []string
		fastlyManifest       string
		skipWindows          bool
		wantError            string
		wantRemediationError string
		wantOutputContains   string
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading package manifest",
			wantRemediationError: "Run `fastly compute init` to ensure a correctly configured manifest.",
		},
		{
			name: "empty language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"`,
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			language = "assemblyscript"`,
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "foobar"`,
			wantError: "unsupported language foobar",
		},
		{
			name: "successful build",
			args: args("compute build"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "assemblyscript"

      [scripts]
      build = "%s"`, compute.AsDefaultBuildCommand),
			wantOutputContains: "Built package 'test'",
			skipWindows:        true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if fstruntime.Windows && testcase.skipWindows {
				t.Skip()
			}

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
					{Src: filepath.Join("testdata", "build", "assemblyscript", "package.json"), Dst: "package.json"},
					{Src: filepath.Join("testdata", "build", "assemblyscript", "assembly", "index.ts"), Dst: filepath.Join("assembly", "index.ts")},
				},
				Write: []testutil.FileIO{
					{Src: testcase.fastlyManifest, Dst: manifest.Filename},
				},
				Exec: []string{"npm", "install"},
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
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			err = app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			if testcase.wantOutputContains != "" {
				testutil.AssertStringContains(t, stdout.String(), testcase.wantOutputContains)
			}
		})
	}
}

func TestBuildJavaScript(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_BUILD_JAVASCRIPT") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD_JAVASCRIPT or TEST_COMPUTE_BUILD to run this test")
	}

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
			{Src: filepath.Join("testdata", "build", "javascript", "package.json"), Dst: "package.json"},
			{Src: filepath.Join("testdata", "build", "javascript", "webpack.config.js"), Dst: "webpack.config.js"},
			{Src: filepath.Join("testdata", "build", "javascript", "src", "index.js"), Dst: filepath.Join("src", "index.js")},
		},
		Exec: []string{"npm", "install"},
	})
	defer os.RemoveAll(rootdir)

	// Before running the test, chdir into the build environment.
	// When we're done, chdir back to our original location.
	// This is so we can reliably copy the testdata/ fixtures.
	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(pwd)

	for _, testcase := range []struct {
		name                 string
		args                 []string
		fastlyManifest       string
		skipWindows          bool
		sourceOverride       string
		wantError            string
		wantRemediationError string
		wantOutputContains   string
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading package manifest",
			wantRemediationError: "Run `fastly compute init` to ensure a correctly configured manifest.",
		},
		{
			name: "empty language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"`,
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			language = "javascript"`,
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "compilation error",
			args: args("compute build --verbose"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "javascript"

      [scripts]
      build = "%s"`, compute.JsDefaultBuildCommand),
			sourceOverride: `D"F;
			'GREGERgregeg '
			ERG`,
			wantError: "error during execution process",
		},
		{
			name: "successful build",
			args: args("compute build"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "javascript"

      [scripts]
      build = "%s"`, compute.JsDefaultBuildCommand),
			wantOutputContains: "Built package 'test'",
			skipWindows:        true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if fstruntime.Windows && testcase.skipWindows {
				t.Skip()
			}

			if testcase.fastlyManifest != "" {
				if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(testcase.fastlyManifest), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			// We want to ensure the original `index.js` is put back in case of a test
			// case overriding its content using `sourceOverride`.
			src := filepath.Join(rootdir, "src", "index.js")
			b, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			defer func(src string, b []byte) {
				err := os.WriteFile(src, b, 0o644)
				if err != nil {
					t.Fatal(err)
				}
			}(src, b)

			if testcase.sourceOverride != "" {
				if err := os.WriteFile(src, []byte(testcase.sourceOverride), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			var stdout threadsafe.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			if testcase.wantOutputContains != "" {
				testutil.AssertStringContains(t, stdout.String(), testcase.wantOutputContains)
			}
		})
	}
}

func TestBuildGo(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_BUILD_GO") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD_GO or TEST_COMPUTE_BUILD to run this test")
	}

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
			{Src: filepath.Join("testdata", "build", "go", "go.mod"), Dst: "go.mod"},
			{Src: filepath.Join("testdata", "build", "go", "main.go"), Dst: "main.go"},
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

	for _, testcase := range []struct {
		name                 string
		args                 []string
		fastlyManifest       string
		sourceOverride       string
		wantError            string
		wantRemediationError string
		wantOutputContains   string
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading package manifest",
			wantRemediationError: "Run `fastly compute init` to ensure a correctly configured manifest.",
		},
		{
			name: "empty language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"`,
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			language = "go"`,
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "syntax error",
			args: args("compute build --verbose"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "go"

      [scripts]
      build = "%s"`, compute.GoDefaultBuildCommand),
			sourceOverride: `D"F;
			'GREGERgregeg '
			ERG`,
			wantError: "error during execution process",
		},
		{
			name: "successful build",
			args: args("compute build"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "go"

      [scripts]
      build = "%s"`, compute.GoDefaultBuildCommand),
			wantOutputContains: "Built package 'test'",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.fastlyManifest != "" {
				if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(testcase.fastlyManifest), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			// We want to ensure the original `main.go` is put back in case of a test
			// case overriding its content using `sourceOverride`.
			src := filepath.Join(rootdir, "main.go")
			b, err := os.ReadFile(src)
			if err != nil {
				t.Fatal(err)
			}
			defer func(src string, b []byte) {
				err := os.WriteFile(src, b, 0o644)
				if err != nil {
					t.Fatal(err)
				}
			}(src, b)

			if testcase.sourceOverride != "" {
				if err := os.WriteFile(src, []byte(testcase.sourceOverride), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			var stdout threadsafe.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)

			// NOTE: The following constraints should be kept in-sync with
			// ./pkg/config/config.toml
			opts.ConfigFile.Language.Go.TinyGoConstraint = ">= 0.24.0-0" // NOTE: -0 is to allow prereleases.
			opts.ConfigFile.Language.Go.ToolchainConstraint = ">= 1.17 < 1.19"

			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			if testcase.wantOutputContains != "" {
				testutil.AssertStringContains(t, stdout.String(), testcase.wantOutputContains)
			}
		})
	}
}

func TestOtherBuild(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	// We're going to chdir to a build environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	//
	// NOTE: Our only requirement is that there be a bin directory. The custom
	// build script we're using in the test is not going to use any files in the
	// directory (the script will just `echo` a message).
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Write: []testutil.FileIO{
			{Src: "mock content", Dst: "bin/testfile"},
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

	for _, testcase := range []struct {
		args                 []string
		dontWantOutput       []string
		fastlyManifest       string
		name                 string
		stdin                string
		wantError            string
		wantOutput           []string
		wantRemediationError string
	}{
		{
			name: "stop build process",
			args: args("compute build --language other"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "other"
			[scripts]
			build = "echo custom build"
      post_build = "echo doing a post build"`,
			stdin: "N",
			wantOutput: []string{
				"echo doing a post build",
				"Are you sure you want to continue with the post build step?",
				"Stopping the post build process",
			},
		},
		{
			name: "allow build process",
			args: args("compute build --language other"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "other"
			[scripts]
			build = "echo custom build"
      post_build = "echo doing a post build"`,
			stdin: "Y",
			wantOutput: []string{
				"echo doing a post build",
				"Are you sure you want to continue with the post build step?",
				"Running [scripts.build]",
				"Built package 'test'",
			},
		},
		{
			name: "language pulled from manifest",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "other"
			[scripts]
			build = "echo custom build"
      post_build = "echo doing a post build"`,
			stdin: "Y",
			wantOutput: []string{
				"echo doing a post build",
				"Are you sure you want to continue with the post build step?",
				"Running [scripts.build]",
				"Built package 'test'",
			},
		},
		{
			name: "avoid prompt confirmation",
			args: args("compute build --auto-yes --language other"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "other"
			[scripts]
			build = "echo custom build"`,
			wantOutput: []string{
				"Running [scripts.build]",
				"Built package 'test'",
			},
			dontWantOutput: []string{
				"Are you sure you want to continue with the build step?",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.fastlyManifest != "" {
				if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(testcase.fastlyManifest), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.Stdin = strings.NewReader(testcase.stdin) // NOTE: build only has one prompt when dealing with a custom build
			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
			for _, s := range testcase.dontWantOutput {
				testutil.AssertStringDoesntContain(t, stdout.String(), s)
			}
		})
	}
}

func TestCustomPostBuild(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

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
		// NOTE: Our only requirement is that there be a bin directory. The custom
		// build script we're using in the test is not going to use any files in the
		// directory (the script will just `echo` a message).
		Write: []testutil.FileIO{
			{Src: "mock content", Dst: "bin/testfile"},
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

	scenarios := []struct {
		applicationConfig    config.File
		args                 []string
		cargoManifest        string
		dontWantOutput       []string
		fastlyManifest       string
		name                 string
		stdin                string
		wantError            string
		wantOutput           []string
		wantRemediationError string
	}{
		// NOTE: We need a fully functioning environment for the following tests,
		// otherwise the call to language.Verify() would fail before reaching
		// language.Build() and we need the build to complete because the
		// post_build isn't executed until AFTER a build is successful.
		{
			name: "stop post_build process",
			args: args("compute build"),
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainConstraint: ">= 1.54.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "rust"
			[scripts]
      build = "%s"
			post_build = "echo custom post_build"`, compute.RustDefaultBuildCommand),
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			stdin: "N",
			wantOutput: []string{
				compute.CustomPostBuildScriptMessage,
				"echo custom post_build",
				"Are you sure you want to continue with the post build step?",
				"Stopping the post build process",
			},
		},
		{
			name: "allow post_build process",
			args: args("compute build"),
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainConstraint: ">= 1.54.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "rust"
			[scripts]
      build = "%s"
			post_build = "echo custom post_build"`, compute.RustDefaultBuildCommand),
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			stdin: "Y",
			wantOutput: []string{
				compute.CustomPostBuildScriptMessage,
				"echo custom post_build",
				"Are you sure you want to continue with the post build step?",
				"Running [scripts.build]",
				"Built package 'test'",
			},
		},
		{
			name: "avoid prompt confirmation",
			args: args("compute build --auto-yes"),
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainConstraint: ">= 1.54.0",
						WasmWasiTarget:      "wasm32-wasi",
					},
				},
			},
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "rust"
			[scripts]
      build = "%s"
			post_build = "echo custom post_build"`, compute.RustDefaultBuildCommand),
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			wantOutput: []string{
				"Running [scripts.build]",
				"Built package 'test'",
			},
			dontWantOutput: []string{
				compute.CustomPostBuildScriptMessage,
				"Are you sure you want to continue with the post build step?",
			},
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.fastlyManifest != "" {
				if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(testcase.fastlyManifest), 0o777); err != nil {
					t.Fatal(err)
				}
			}
			if testcase.cargoManifest != "" {
				if err := os.WriteFile(filepath.Join(rootdir, "Cargo.toml"), []byte(testcase.cargoManifest), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ConfigFile = testcase.applicationConfig
			opts.Stdin = strings.NewReader(testcase.stdin) // NOTE: build only has one prompt when dealing with a custom build
			err = app.Run(opts)

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
			for _, s := range testcase.dontWantOutput {
				testutil.AssertStringDoesntContain(t, stdout.String(), s)
			}
		})
	}
}
