package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/testutil"
)

// TestBuildRust validates that the rust ecosystem is in place and accurate.
//
// NOTE:
// The defined tests rely on some key pieces of information:
//
// 1. The `fastly` crate internally consumes a `fastly-sys` crate.
// 2. The `fastly-sys` create didn't exist until `fastly` 0.4.0.
// 3. Users of `fastly` should always have the latest `fastly-sys`.
// 4. Users of `fastly` shouldn't need to know about `fastly-sys`.
// 5. Each test has a 'default' Cargo.toml created for it.
// 6. Each test can override the default Cargo.toml by defining a `cargoManifest`.
//
// You can locate the default Cargo.toml here:
// pkg/compute/testdata/build/Cargo.toml
func TestBuildRust(t *testing.T) {
	args := testutil.Args
	if os.Getenv("TEST_COMPUTE_BUILD_RUST") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	for _, testcase := range []struct {
		name                 string
		args                 []string
		applicationConfig    config.File
		fastlyManifest       string
		cargoManifest        string
		cargoLock            string
		client               api.HTTPClient
		wantError            string
		wantRemediationError string
		wantOutputContains   string
	}{
		{
			name:      "no fastly.toml manifest",
			args:      args("compute build"),
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "error reading package manifest: open fastly.toml:", // actual message differs on Windows
		},
		{
			name: "empty language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"`,
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			language = "rust"`,
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "foobar"`,
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "unsupported language foobar",
		},
		{
			name: "error reading cargo metadata",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"`,
			client:    versionClient{fastlyVersions: []string{"0.4.0"}},
			wantError: "reading cargo metadata",
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
						ToolchainConstraint: ">= 1.49.0 < 2.0.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: "0.0.0",
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
		},
		{
			name: "fastly-sys crate not found",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.3.2"`,
			cargoLock: `
			[[package]]
			name = "test"
			version = "0.1.0"

			[[package]]
			name = "fastly"
			version = "0.3.2"`,
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
						ToolchainConstraint: ">= 1.49.0 < 2.0.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: "0.0.0",
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
			client: versionClient{
				fastlyVersions: []string{"0.4.0"},
			},
			wantError:            "fastly-sys crate not found", // fastly 0.3.3 is where fastly-sys was introduced
			wantRemediationError: "fastly = \"^0.4.0\"",
		},
		{
			name: "fastly-sys crate out-of-date",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.4.0"`,
			cargoLock: `
			[[package]]
			name = "fastly-sys"
			version = "0.3.7"`,
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
						ToolchainConstraint: ">= 1.49.0 < 2.0.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: ">= 0.4.0 <= 0.9.0", // the fastly-sys version in 0.6.0 is actually ^0.3.6 so a minimum of 0.4.0 causes the constraint to fail
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
			client: versionClient{
				fastlyVersions: []string{"0.6.0"},
			},
			wantError:            "fastly crate not up-to-date",
			wantRemediationError: "fastly = \"^0.6.0\"",
		},
		{
			name: "rust toolchain does not match the constraint",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.4.0"`,
			cargoLock: `
			[[package]]
			name = "fastly-sys"
			version = "0.3.7"`,
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						// NOTE: my local rust environment has versions 1.[45|46|49].0
						// So I've set the constraint to ensure the test fails.
						// Example, the code logic will select 1.49.0, which is outside the constraint limit 1.40.0
						ToolchainVersion:    "1.0.0",
						ToolchainConstraint: ">= 1.0.0 < 1.40.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: ">= 0.4.0 <= 0.9.0", // the fastly-sys version in 0.6.0 is actually ^0.3.6 so a minimum of 0.4.0 causes the constraint to fail
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
			client: versionClient{
				fastlyVersions: []string{"0.6.0"},
			},
			wantError:            "rust toolchain 1.49.0 is incompatible with the constraint >= 1.0.0 < 1.40.0",
			wantRemediationError: "To fix this error, run the following command with a version within the given range >= 1.0.0 < 1.40.0:\n\n\t$ rustup toolchain install <version>\n",
		},
		{
			name: "fastly crate prerelease",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
						ToolchainConstraint: ">= 1.49.0 < 2.0.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: ">= 0.3.0 <= 0.6.0",
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "0.6.0"`,
			cargoLock: `
			[[package]]
			name = "fastly-sys"
			version = "0.3.7"

			[[package]]
			name = "fastly"
			version = "0.6.0"`,
			client: versionClient{
				fastlyVersions: []string{"0.6.0"},
			},
			wantOutputContains: "Built rust package test",
		},
		{
			name: "Rust success",
			args: args("compute build"),
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
						ToolchainConstraint: ">= 1.49.0 < 2.0.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: ">= 0.3.0 <= 0.6.0",
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			cargoManifest: `
			[package]
			name = "test"
			version = "0.1.0"

			[dependencies]
			fastly = "=0.6.0"`,
			cargoLock: `
			[[package]]
			name = "fastly"
			version = "0.6.0"

			[[package]]
			name = "fastly-sys"
			version = "0.3.7"`,
			client: versionClient{
				fastlyVersions: []string{"0.6.0"},
			},
			wantOutputContains: "Built rust package test",
		},
	} {
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
					{Src: filepath.Join("testdata", "build", "Cargo.lock"), Dst: "Cargo.lock"},
					{Src: filepath.Join("testdata", "build", "Cargo.toml"), Dst: "Cargo.toml"},
					{Src: filepath.Join("testdata", "build", "src", "main.rs"), Dst: filepath.Join("src", "main.rs")},
					{Src: filepath.Join("testdata", "deploy", "pkg", "package.tar.gz"), Dst: filepath.Join("pkg", "package.tar.gz")},
				},
				Write: []testutil.FileIO{
					{Src: testcase.fastlyManifest, Dst: manifest.Filename},
					{Src: testcase.cargoManifest, Dst: "Cargo.toml"},
					{Src: testcase.cargoLock, Dst: "Cargo.lock"},
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
			opts.HTTPClient = testcase.client
			err = app.Run(opts)
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
		wantError            string
		wantRemediationError string
		wantOutputContains   string
	}{
		{
			name:      "no fastly.toml manifest",
			args:      args("compute build"),
			wantError: "error reading package manifest: open fastly.toml:", // actual message differs on Windows
		},
		{
			name: "empty language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"`,
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			language = "assemblyscript"`,
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "unknown language",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "javascript"`,
			wantError: "unsupported language javascript",
		},
		{
			name: "AssemblyScript success",
			args: args("compute build"),
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "assemblyscript"`,
			wantOutputContains: "Built assemblyscript package test",
		},
	} {
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
					{Src: filepath.Join("testdata", "build", "package.json"), Dst: "package.json"},
					{Src: filepath.Join("testdata", "build", "assembly", "index.ts"), Dst: filepath.Join("assembly", "index.ts")},
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
