package compute_test

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/sync"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
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
			args:      []string{"compute", "build"},
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "error reading package manifest: open fastly.toml:", // actual message differs on Windows
		},
		{
			name: "empty language",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			name = "test"`,
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			language = "rust"`,
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "unknown language",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "foobar"`,
			client:    versionClient{fastlyVersions: []string{"0.0.0"}},
			wantError: "unsupported language foobar",
		},
		{
			name: "error reading cargo metadata",
			args: []string{"compute", "build"},
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
						// TODO: pull actual version from .github/workflows/pr_test.yml
						// when doing local run of integration tests.
						ToolchainVersion:    "1.49.0",
						WasmWasiTarget:      "wasm32-wasi",
						FastlySysConstraint: "0.0.0",
						RustupConstraint:    ">= 1.23.0",
					},
				},
			},
		},
		{
			name: "fastly-sys crate not found",
			args: []string{"compute", "build"},
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
			args: []string{"compute", "build"},
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
			name: "fastly crate prerelease",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "rust"`,
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
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
			args: []string{"compute", "build"},
			applicationConfig: config.File{
				Language: config.Language{
					Rust: config.Rust{
						ToolchainVersion:    "1.49.0",
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

			// Create our build environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeRustBuildEnvironment(t, testcase.fastlyManifest, testcase.cargoManifest, testcase.cargoLock)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = testcase.applicationConfig
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(mock.API{})
				httpClient                     = testcase.client
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = sync.NewWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			if testcase.wantOutputContains != "" {
				testutil.AssertStringContains(t, buf.String(), testcase.wantOutputContains)
			}
		})
	}
}

func makeRustBuildEnvironment(t *testing.T, fastlyManifestContent, cargoManifestContent, cargoLockContent string) (rootdir string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-build-*")
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range [][]string{
		{"Cargo.toml"},
		{"Cargo.lock"},
		{"src", "main.rs"},
	} {
		fromFilename := filepath.Join("testdata", "build", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		testutil.CopyFile(t, fromFilename, toFilename)
	}

	if fastlyManifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(fastlyManifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	if cargoManifestContent != "" {
		filename := filepath.Join(rootdir, "Cargo.toml")
		if err := os.WriteFile(filename, []byte(cargoManifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	if cargoLockContent != "" {
		filename := filepath.Join(rootdir, "Cargo.lock")
		if err := os.WriteFile(filename, []byte(cargoLockContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}

func TestBuildAssemblyScript(t *testing.T) {
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
			args:      []string{"compute", "build"},
			wantError: "error reading package manifest: open fastly.toml:", // actual message differs on Windows
		},
		{
			name: "empty language",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			name = "test"`,
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name: "empty name",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			language = "assemblyscript"`,
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name: "unknown language",
			args: []string{"compute", "build"},
			fastlyManifest: `
			manifest_version = 1
			name = "test"
			language = "javascript"`,
			wantError: "unsupported language javascript",
		},
		{
			name: "AssemblyScript success",
			args: []string{"compute", "build"},
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

			// Create our build environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeAssemblyScriptBuildEnvironment(t, testcase.fastlyManifest)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(mock.API{})
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = sync.NewWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertRemediationErrorContains(t, err, testcase.wantRemediationError)
			if testcase.wantOutputContains != "" {
				testutil.AssertStringContains(t, buf.String(), testcase.wantOutputContains)
			}
		})
	}
}

func makeAssemblyScriptBuildEnvironment(t *testing.T, fastlyManifestContent string) (rootdir string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-build-*")
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range [][]string{
		{"package.json"},
		{"assembly", "index.ts"},
	} {
		fromFilename := filepath.Join("testdata", "build", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		testutil.CopyFile(t, fromFilename, toFilename)
	}

	if fastlyManifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(fastlyManifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command("npm", "install")
	cmd.Dir = rootdir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	return rootdir
}
