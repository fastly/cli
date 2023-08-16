package compute_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

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
		wantOutput           []string
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading fastly.toml",
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
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "foobar"`,
			wantError: "unsupported language foobar",
		},
		// The following test validates that the project compiles successfully even
		// though the fastly.toml manifest has no build script. There should be a
		// default build script inserted and it should use the same name as the
		// project/package name in the Cargo.toml.
		//
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "build script inserted dynamically when missing",
			args: args("compute build --verbose"),
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
			name = "my-project"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			fastlyManifest: `
			manifest_version = 2
			name = "test"
      language = "rust"`,
			wantOutput: []string{
				"No [scripts.build] found in fastly.toml.", // requires --verbose
				"The following default build command for",
				"cargo build --bin my-project",
			},
		},
		{
			name: "build error",
			args: args("compute build"),
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
			name = "fastly-compute-project"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "rust"

      [scripts]
      build = "echo no compilation happening"`,
			wantRemediationError: compute.DefaultBuildErrorRemediation,
		},
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "successful build",
			args: args("compute build --verbose"),
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
			name = "fastly-compute-project"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "rust"

      [scripts]
      build = "%s"`, fmt.Sprintf(compute.RustDefaultBuildCommand, compute.RustDefaultPackageName)),
			wantOutput: []string{
				"Creating ./bin directory (for Wasm binary)",
				"Built package",
			},
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

			var stdout threadsafe.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ConfigFile = testcase.applicationConfig
			err = app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			// NOTE: Some errors we want to assert only the remediation.
			// e.g. a 'stat' error isn't the same across operating systems/platforms.
			if testcase.wantError != "" {
				testutil.AssertErrorContains(t, err, testcase.wantError)
			}
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}

func TestBuildGo(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_BUILD_GO") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	args := testutil.Args

	scenarios := []struct {
		name                 string
		args                 []string
		applicationConfig    config.File
		fastlyManifest       string
		wantError            string
		wantRemediationError string
		wantOutput           []string
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading fastly.toml",
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
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "foobar"`,
			wantError: "unsupported language foobar",
		},
		// The following test validates that the project compiles successfully even
		// though the fastly.toml manifest has no build script. There should be a
		// default build script inserted.
		//
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "build success",
			args: args("compute build --verbose"),
			applicationConfig: config.File{
				Language: config.Language{
					Go: config.Go{
						TinyGoConstraint:          ">= 0.26.0-0",
						ToolchainConstraintTinyGo: ">= 1.18",
						ToolchainConstraint:       ">= 1.21",
					},
				},
			},
			fastlyManifest: `
			manifest_version = 2
			name = "test"
      language = "go"
      [scripts]
      build = "go build -o bin/main.wasm ./"
      `,
			wantOutput: []string{
				"The Fastly CLI build step requires a go version '>= 1.21'",
				"Build script to execute",
				"go build -o bin/main.wasm ./",
				"Creating ./bin directory (for Wasm binary)",
				"Built package",
			},
		},
		// The following test case is expected to fail because we specify a custom
		// build script that doesn't actually produce a ./bin/main.wasm
		{
			name: "build error",
			args: args("compute build"),
			applicationConfig: config.File{
				Language: config.Language{
					Go: config.Go{
						TinyGoConstraint:          ">= 0.26.0-0",
						ToolchainConstraintTinyGo: ">= 1.18",
						ToolchainConstraint:       ">= 1.21",
					},
				},
			},
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "go"

      [scripts]
      build = "echo no compilation happening"`,
			wantRemediationError: compute.DefaultBuildErrorRemediation,
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
					{Src: filepath.Join("testdata", "build", "go", "go.mod"), Dst: "go.mod"},
					{Src: filepath.Join("testdata", "build", "go", "main.go"), Dst: "main.go"},
				},
				Write: []testutil.FileIO{
					{Src: testcase.fastlyManifest, Dst: manifest.Filename},
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

			var stdout threadsafe.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ConfigFile = testcase.applicationConfig
			err = app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			// NOTE: Some errors we want to assert only the remediation.
			// e.g. a 'stat' error isn't the same across operating systems/platforms.
			if testcase.wantError != "" {
				testutil.AssertErrorContains(t, err, testcase.wantError)
			}
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}

func TestBuildJavaScript(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_BUILD_JAVASCRIPT") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	args := testutil.Args

	scenarios := []struct {
		name                 string
		args                 []string
		fastlyManifest       string
		wantError            string
		wantRemediationError string
		wantOutput           []string
		npmInstall           bool
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading fastly.toml",
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
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "foobar"`,
			wantError: "unsupported language foobar",
		},
		// The following test validates that the project compiles successfully even
		// though the fastly.toml manifest has no build script. There should be a
		// default build script inserted.
		//
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "build script inserted dynamically when missing",
			args: args("compute build --verbose"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
      language = "javascript"`,
			wantOutput: []string{
				"No [scripts.build] found in fastly.toml.", // requires --verbose
				"The following default build command for",
				"npm exec webpack", // our testdata package.json references webpack
			},
		},
		{
			name: "build error",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "javascript"

      [scripts]
      build = "echo no compilation happening"`,
			wantRemediationError: compute.DefaultBuildErrorRemediation,
		},
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "successful build",
			args: args("compute build --verbose"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "javascript"

      [scripts]
      build = "%s"`, compute.JsDefaultBuildCommandForWebpack),
			wantOutput: []string{
				"Creating ./bin directory (for Wasm binary)",
				"Built package",
			},
			npmInstall: true,
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
					{Src: filepath.Join("testdata", "build", "javascript", "package.json"), Dst: "package.json"},
					{Src: filepath.Join("testdata", "build", "javascript", "webpack.config.js"), Dst: "webpack.config.js"},
					{Src: filepath.Join("testdata", "build", "javascript", "src", "index.js"), Dst: filepath.Join("src", "index.js")},
				},
				Write: []testutil.FileIO{
					{Src: testcase.fastlyManifest, Dst: manifest.Filename},
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

			// NOTE: We only want to run `npm install` for the success case.
			if testcase.npmInstall {
				// gosec flagged this:
				// G204 (CWE-78): Subprocess launched with variable
				// Disabling as we control this command.
				// #nosec
				// nosemgrep
				cmd := exec.Command("npm", "install")

				err = cmd.Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			var stdout threadsafe.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			err = app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			// NOTE: Some errors we want to assert only the remediation.
			// e.g. a 'stat' error isn't the same across operating systems/platforms.
			if testcase.wantError != "" {
				testutil.AssertErrorContains(t, err, testcase.wantError)
			}
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}

func TestBuildAssemblyScript(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_BUILD_ASSEMBLYSCRIPT") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	args := testutil.Args

	scenarios := []struct {
		name                 string
		args                 []string
		fastlyManifest       string
		wantError            string
		wantRemediationError string
		wantOutput           []string
		npmInstall           bool
	}{
		{
			name:                 "no fastly.toml manifest",
			args:                 args("compute build"),
			wantError:            "error reading fastly.toml",
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
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "foobar"`,
			wantError: "unsupported language foobar",
		},
		// The following test validates that the project compiles successfully even
		// though the fastly.toml manifest has no build script. There should be a
		// default build script inserted.
		//
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "build script inserted dynamically when missing",
			args: args("compute build --verbose"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
      language = "assemblyscript"`,
			wantOutput: []string{
				"No [scripts.build] found in fastly.toml.", // requires --verbose
				"The following default build command for",
				"npm exec -- asc",
			},
		},
		{
			name: "build error",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "assemblyscript"

      [scripts]
      build = "echo no compilation happening"`,
			wantRemediationError: compute.DefaultBuildErrorRemediation,
		},
		// NOTE: This test passes --verbose so we can validate specific outputs.
		{
			name: "successful build",
			args: args("compute build --verbose"),
			fastlyManifest: fmt.Sprintf(`
			manifest_version = 2
			name = "test"
			language = "assemblyscript"

      [scripts]
      build = "%s"`, compute.AsDefaultBuildCommand),
			wantOutput: []string{
				"Creating ./bin directory (for Wasm binary)",
				"Built package",
			},
			npmInstall: true,
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
					{Src: filepath.Join("testdata", "build", "assemblyscript", "package.json"), Dst: "package.json"},
					{Src: filepath.Join("testdata", "build", "assemblyscript", "assembly", "index.ts"), Dst: filepath.Join("assembly", "index.ts")},
				},
				Write: []testutil.FileIO{
					{Src: testcase.fastlyManifest, Dst: manifest.Filename},
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

			// NOTE: We only want to run `npm install` for the success case.
			if testcase.npmInstall {
				// gosec flagged this:
				// G204 (CWE-78): Subprocess launched with variable
				// Disabling as we control this command.
				// #nosec
				// nosemgrep
				cmd := exec.Command("npm", "install")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err = cmd.Run()
				if err != nil {
					t.Fatal(err)
				}
			}

			var stdout threadsafe.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			err = app.Run(opts)
			t.Log(stdout.String())
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			// NOTE: Some errors we want to assert only the remediation.
			// e.g. a 'stat' error isn't the same across operating systems/platforms.
			if testcase.wantError != "" {
				testutil.AssertErrorContains(t, err, testcase.wantError)
			}
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
		})
	}
}

// NOTE: TestBuildOther also validates the post_build settings.
func TestBuildOther(t *testing.T) {
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
			[scripts]
			build = "touch ./bin/main.wasm"
      post_build = "echo doing a post build"`,
			stdin: "N",
			wantOutput: []string{
				"echo doing a post build",
				"Do you want to run this now?",
			},
			wantError: "build process stopped by user",
		},
		// NOTE: All following tests pass --verbose so we can see post_build output.
		{
			name: "allow build process",
			args: args("compute build --language other --verbose"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			[scripts]
			build = "touch ./bin/main.wasm"
      post_build = "echo doing a post build"`,
			stdin: "Y",
			wantOutput: []string{
				"echo doing a post build",
				"Do you want to run this now?",
				"Built package",
			},
		},
		{
			name: "language pulled from manifest",
			args: args("compute build --verbose"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			language = "other"
			[scripts]
			build = "touch ./bin/main.wasm"
      post_build = "echo doing a post build"`,
			stdin: "Y",
			wantOutput: []string{
				"echo doing a post build",
				"Do you want to run this now?",
				"Built package",
			},
		},
		{
			name: "avoid prompt confirmation",
			args: args("compute build --auto-yes --language other --verbose"),
			fastlyManifest: `
			manifest_version = 2
			name = "test"
			[scripts]
			build = "touch ./bin/main.wasm"
      post_build = "echo doing a post build with no confirmation prompt && exit 1"`, // force an error so post_build is displayed to validate it was run.
			wantOutput: []string{
				"doing a post build with no confirmation prompt",
			},
			dontWantOutput: []string{
				"Do you want to run this now?",
			},
			wantError: "exit status 1", // because we have to trigger an error to see the post_build output
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.fastlyManifest != "" {
				if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(testcase.fastlyManifest), 0o777); err != nil {
					t.Fatal(err)
				}
			}

			var stdout threadsafe.Buffer
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
