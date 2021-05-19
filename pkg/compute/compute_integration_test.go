package compute_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestInit(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_INIT") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_INIT to run this test")
	}

	for _, testcase := range []struct {
		name             string
		args             []string
		configFile       config.File
		api              mock.API
		manifest         string
		wantFiles        []string
		unwantedFiles    []string
		stdin            string
		wantError        string
		wantOutput       []string
		manifestIncludes string
	}{
		{
			name:      "unknown repository",
			args:      []string{"compute", "init", "--from", "https://example.com/template"},
			wantError: "error fetching package template:",
		},
		{
			name: "with name",
			args: []string{"compute", "init", "--name", "test"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `name = "test"`,
		},
		{
			name: "with description",
			args: []string{"compute", "init", "--description", "test"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `description = "test"`,
		},
		{
			name: "with author",
			args: []string{"compute", "init", "--author", "test@example.com"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `authors = ["test@example.com"]`,
		},
		{
			name: "with multiple authors",
			args: []string{"compute", "init", "--author", "test1@example.com", "--author", "test2@example.com"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `authors = ["test1@example.com", "test2@example.com"]`,
		},
		{
			name: "with from repository and branch",
			args: []string{"compute", "init", "--from", "https://github.com/fastly/compute-starter-kit-rust-default.git", "--branch", "main"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
		},
		{
			name: "with existing package manifest",
			args: []string{"compute", "init", "--force"}, // --force will ignore that the directory isn't empty
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			manifest: strings.Join([]string{
				"manifest_version = \"1\"",
				"service_id = \"1234\"",
				"name = \"test\"",
				"language = \"rust\"",
				"description = \"test\"",
				"authors = [\"test@fastly.com\"]",
			}, "\n"),
			wantOutput: []string{
				"Updating package manifest...",
				"Initializing package...",
			},
		},
		{
			name: "default",
			args: []string{"compute", "init"},
			configFile: config.File{
				User: config.User{
					Email: "test@example.com",
				},
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			manifestIncludes: `authors = ["test@example.com"]`,
			wantFiles: []string{
				"Cargo.toml",
				"fastly.toml",
				"src/main.rs",
			},
			unwantedFiles: []string{
				"SECURITY.md",
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
		},
		{
			name: "non empty directory",
			args: []string{"compute", "init"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			wantError: "project directory not empty",
			manifest:  `name = "test"`, // causes a file to be created as part of test setup
		},
		{
			name: "with default name inferred from directory",
			args: []string{"compute", "init"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name:   "Default",
							Path:   "https://github.com/fastly/compute-starter-kit-rust-default.git",
							Branch: "0.6.0",
						},
					},
				},
			},
			manifestIncludes: `name = "fastly-init`,
		},
		{
			name: "with AssemblyScript language",
			args: []string{"compute", "init", "--language", "assemblyscript"},
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					AssemblyScript: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-assemblyscript-default",
							Tag:  "v0.2.0",
						},
					},
				},
			},
			manifestIncludes: `name = "fastly-init`,
		},
		{
			name:             "with pre-compiled Wasm binary",
			args:             []string{"compute", "init", "--language", "other"},
			manifestIncludes: `language = "other"`,
			wantOutput: []string{
				"Initialized package",
				"To package a pre-compiled Wasm binary for deployment",
				"SUCCESS: Initialized package",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to an init environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our init environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeInitEnvironment(t, testcase.manifest)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the init environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(pwd)

			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = testcase.configFile
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = bytes.NewBufferString(testcase.stdin)
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, file := range testcase.wantFiles {
				if _, err := os.Stat(filepath.Join(rootdir, file)); err != nil {
					t.Errorf("wanted file %s not found", file)
				}
			}
			for _, file := range testcase.unwantedFiles {
				if _, err := os.Stat(filepath.Join(rootdir, file)); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("unwanted file %s found", file)
				}
			}
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
			}
			if testcase.manifestIncludes != "" {
				content, err := os.ReadFile(filepath.Join(rootdir, compute.ManifestFilename))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}

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
				out           io.Writer = common.NewSyncWriter(&buf)
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
				out           io.Writer = common.NewSyncWriter(&buf)
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

func TestDeploy(t *testing.T) {
	for _, testcase := range []struct {
		name             string
		args             []string
		manifest         string
		api              mock.API
		wantError        string
		wantOutput       []string
		manifestIncludes string
		in               *strings.Reader // to handle text.Input prompts
	}{
		{
			name:      "no token",
			args:      []string{"compute", "deploy"},
			wantError: "no token provided",
		},
		{
			name:      "no fastly.toml manifest",
			args:      []string{"compute", "deploy", "--token", "123"},
			in:        strings.NewReader(""),
			wantError: "error reading package manifest",
		},
		{
			// If no Service ID defined via flag or manifest, then the expectation is
			// for the service to be created via the API and for the returned ID to
			// be stored into the manifest.
			//
			// Additionally it validates that the specified path (files generated by
			// the test suite `makeDeployEnvironment` function) cause no issues.
			name: "path with no service ID",
			args: []string{"compute", "deploy", "--token", "123", "-v", "-p", "pkg/package.tar.gz"},
			in:   strings.NewReader(""),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			manifest: "name = \"package\"\n",
			wantOutput: []string{
				"Setting service ID in manifest to \"12345\"...",
				"Deployed package (service 12345, version 1)",
			},
		},
		// Same validation as above with the exception that we use the default path
		// parsing logic (i.e. we don't explicitly pass a path via `-p` flag).
		{
			name:     "empty service ID",
			args:     []string{"compute", "deploy", "--token", "123", "-v"},
			manifest: "name = \"package\"\n",
			in:       strings.NewReader(""),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			wantOutput: []string{
				"Setting service ID in manifest to \"12345\"...",
				"Deployed package (service 12345, version 1)",
			},
		},
		{
			name: "list versions error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: listVersionsError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error listing service versions: fixture error",
		},
		{
			name: "clone version error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error cloning latest service version: fixture error",
		},
		{
			name: "list domains error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
				ListDomainsFn:  listDomainsError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error fetching service domains: fixture error",
		},
		{
			name: "list backends error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
				ListDomainsFn:  listDomainsOk,
				ListBackendsFn: listBackendsError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error fetching service backends: fixture error",
		},
		// The following test doesn't just validate the package API error behaviour
		// but as a side effect it validates that when deleting the created
		// service, the Service ID is also cleared out from the manifest.
		{
			name: "package API error",
			args: []string{"compute", "deploy", "--token", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				ListVersionsFn:  listVersionsActiveOk,
				CloneVersionFn:  cloneVersionOk,
				ListDomainsFn:   listDomainsOk,
				ListBackendsFn:  listBackendsOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
				GetPackageFn:    getPackageOk,
				UpdatePackageFn: updatePackageError,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			manifest:  `name = "package"`,
			wantError: "error uploading package: fixture error",
			wantOutput: []string{
				"Uploading package...",
			},
			manifestIncludes: `service_id = ""`,
		},
		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. We mock the API call to fail, and we expect to see a
		// relevant error message related to that error.
		{
			name: "service create error",
			args: []string{"compute", "deploy", "--token", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				CreateServiceFn: createServiceError,
			},
			manifest:  "name = \"package\"\n",
			wantError: "error creating service: fixture error",
			wantOutput: []string{
				"Creating service...",
			},
		},
		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. We mock the service creation to be successful while we
		// mock the domain API call to fail, and we expect to see a relevant error
		// message related to that error.
		{
			name: "service domain error",
			args: []string{"compute", "deploy", "--token", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainError,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			manifest:  "name = \"package\"\n",
			wantError: "error creating domain: fixture error",
			wantOutput: []string{
				"Creating service...",
				"Creating domain...",
			},
		},
		// The following test mocks the backend API call to fail, and we expect to
		// see a relevant error message related to that error.

		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. We mock the service creation to be successful while we
		// mock the backend API call to fail, and we expect to see a relevant error
		// message related to that error.
		{
			name: "service backend error",
			args: []string{"compute", "deploy", "--token", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				CreateServiceFn: createServiceOK,
				CloneVersionFn:  cloneVersionOk,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendError,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			manifest:  "name = \"package\"\n",
			wantError: "error creating backend: fixture error",
			wantOutput: []string{
				"Creating service...",
				"Creating domain...",
				"Creating backend...",
			},
		},
		// The following test additionally validates that the undoStack is executed
		// as expected (e.g. the backend and domain resources are deleted).
		{
			name: "activate error",
			args: []string{"compute", "deploy", "--token", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error activating version: fixture error",
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
			},
		},
		{
			name: "indentical package",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
				ListDomainsFn:  listDomainsOk,
				ListBackendsFn: listBackendsOk,
				GetPackageFn:   getPackageIdentical,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Skipping package deployment",
			},
		},
		{
			name: "success",
			args: []string{"compute", "deploy", "--token", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Manage this service at:",
				"https://manage.fastly.com/configure/services/123",
				"View this service at:",
				"https://directly-careful-coyote.edgecompute.app",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "success with path",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Manage this service at:",
				"https://manage.fastly.com/configure/services/123",
				"View this service at:",
				"https://directly-careful-coyote.edgecompute.app",
				"Deployed package (service 123, version 2)",
			},
		},
		// The following test validates when the ideal latest version is 'inactive',
		// then we don't clone the version as we can just go ahead and activate it.
		{
			name: "success with inactive version",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsInactiveOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "success with version",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123", "--version", "2"},
			in:   strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 2)",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our deploy environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeDeployEnvironment(t, testcase.manifest)
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
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = testcase.in
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)

			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)

			testutil.AssertErrorContains(t, err, testcase.wantError)

			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
			}

			if testcase.manifestIncludes != "" {
				content, err := os.ReadFile(filepath.Join(rootdir, compute.ManifestFilename))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}

func TestPublish(t *testing.T) {
	for _, testcase := range []struct {
		name              string
		args              []string
		applicationConfig config.File
		fastlyManifest    string
		cargoManifest     string
		cargoLock         string
		client            api.HTTPClient
		in                io.Reader
		api               mock.API
		wantError         string
		wantOutput        []string
		manifestIncludes  string
	}{
		{
			name: "success no command flags",
			args: []string{"compute", "publish", "-t", "123"},
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
			service_id = "123"
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
			in: strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ListBackendsFn:    listBackendsOk,
				ListDomainsFn:     listDomainsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				ActivateVersionFn: activateVersionOk,
			},
			wantOutput: []string{
				"Built rust package test",
				"Uploading package...",
				"Activating version...",
				"Manage this service at:",
				"https://manage.fastly.com/configure/services/123",
				"View this service at:",
				"https://directly-careful-coyote.edgecompute.app",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "success with build command flags",
			args: []string{"compute", "publish", "-t", "123", "--name", "test", "--language", "rust", "--include-source", "--force"},
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
			service_id = "123"
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
			in: strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ListBackendsFn:    listBackendsOk,
				ListDomainsFn:     listDomainsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				ActivateVersionFn: activateVersionOk,
			},
			wantOutput: []string{
				"Built rust package test",
				"Uploading package...",
				"Activating version...",
				"Manage this service at:",
				"https://manage.fastly.com/configure/services/123",
				"View this service at:",
				"https://directly-careful-coyote.edgecompute.app",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "success with deploy command flags",
			args: []string{"compute", "publish", "-t", "123", "--version", "2", "--path", "pkg/test.tar.gz"},
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
			service_id = "123"
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
			in: strings.NewReader(""),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				ActivateVersionFn: activateVersionOk,
			},
			wantOutput: []string{
				"Built rust package test",
				"Uploading package...",
				"Activating version...",
				"Manage this service at:",
				"https://manage.fastly.com/configure/services/123",
				"View this service at:",
				"https://directly-careful-coyote.edgecompute.app",
				"Deployed package (service 123, version 2)",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our publish environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makePublishEnvironment(t, testcase.fastlyManifest, testcase.cargoManifest, testcase.cargoLock)
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
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = testcase.client
				cliVersioner  update.Versioner = nil
				in            io.Reader        = testcase.in
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
			}
			if testcase.manifestIncludes != "" {
				content, err := os.ReadFile(filepath.Join(rootdir, compute.ManifestFilename))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		args       []string
		api        mock.API
		wantError  string
		wantOutput []string
	}{
		{
			name: "package API error",
			args: []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz", "-t", "123"},
			api: mock.API{
				UpdatePackageFn: updatePackageError,
			},
			wantError: "error uploading package: fixture error",
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
			},
		},
		{
			name: "success",
			args: []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz", "-t", "123"},
			api: mock.API{
				UpdatePackageFn: updatePackageOk,
			},
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
				"Updated package (service 123, version 1)",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our deploy environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeDeployEnvironment(t, "")
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
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
			}
		})
	}
}

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
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
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

func TestValidate(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		args       []string
		wantError  string
		wantOutput string
	}{
		{
			name:       "success",
			args:       []string{"compute", "validate", "-p", "pkg/package.tar.gz"},
			wantError:  "",
			wantOutput: "Validated package",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create our deploy environment in a temp dir.
			// Defer a call to clean it up.
			rootdir := makeDeployEnvironment(t, "")
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
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, buf.String(), testcase.wantOutput)
		})
	}
}

func makeInitEnvironment(t *testing.T, manifestContent string) (rootdir string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-init-*")
	if err != nil {
		t.Fatal(err)
	}

	if manifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
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
		copyFile(t, fromFilename, toFilename)
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
		copyFile(t, fromFilename, toFilename)
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

func makeDeployEnvironment(t *testing.T, manifestContent string) (rootdir string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-deploy-*")
	if err != nil {
		t.Fatal(err)
	}

	for _, filename := range [][]string{
		{"pkg", "package.tar.gz"},
	} {
		fromFilename := filepath.Join("testdata", "deploy", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		copyFile(t, fromFilename, toFilename)
	}

	if manifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
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
		copyFile(t, fromFilename, toFilename)
	}

	if manifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}

func makePublishEnvironment(t *testing.T, fastlyManifestContent, cargoManifestContent, cargoLockContent string) (rootdir string) {
	t.Helper()

	rootdir, err := os.MkdirTemp("", "fastly-publish-*")
	if err != nil {
		t.Fatal(err)
	}

	// BUILD REQUIREMENTS

	for _, filename := range [][]string{
		{"Cargo.toml"},
		{"Cargo.lock"},
		{"src", "main.rs"},
	} {
		fromFilename := filepath.Join("testdata", "build", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		copyFile(t, fromFilename, toFilename)
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

	// DEPLOY REQUIREMENTS

	for _, filename := range [][]string{
		{"pkg", "package.tar.gz"},
	} {
		fromFilename := filepath.Join("testdata", "deploy", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		copyFile(t, fromFilename, toFilename)
	}

	return rootdir
}

func copyFile(t *testing.T, fromFilename, toFilename string) {
	t.Helper()

	src, err := os.Open(fromFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()

	toDir := filepath.Dir(toFilename)
	if err := os.MkdirAll(toDir, 0777); err != nil {
		t.Fatal(err)
	}

	dst, err := os.Create(toFilename)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}

	if err := dst.Sync(); err != nil {
		t.Fatal(err)
	}

	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
}

var errTest = errors.New("fixture error")

func createServiceOK(i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: i.Name,
		Type: i.Type,
	}, nil
}

func getServiceOK(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: "test",
	}, nil
}

func createServiceError(*fastly.CreateServiceInput) (*fastly.Service, error) {
	return nil, errTest
}

func deleteServiceOK(i *fastly.DeleteServiceInput) error {
	return nil
}

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createDomainError(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func deleteDomainOK(i *fastly.DeleteDomainInput) error {
	return nil
}

func createBackendOK(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createBackendError(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return nil, errTest
}

func deleteBackendOK(i *fastly.DeleteBackendInput) error {
	return nil
}

func listVersionsInactiveOk(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    false,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func listVersionsActiveOk(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func listVersionsError(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return nil, errTest
}

func getPackageOk(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func getPackageIdentical(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Metadata: fastly.PackageMetadata{
			HashSum: "2b742f99854df7e024c287e36fb0fdfc5414942e012be717e52148ea0d6800d66fc659563f6f11105815051e82b14b61edc84b33b49789b790db1ed3446fb483",
		},
	}, nil
}

func cloneVersionOk(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion + 1}, nil
}

func cloneVersionError(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return nil, errTest
}

func updatePackageOk(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.ServiceID, ServiceVersion: i.ServiceVersion}, nil
}

func updatePackageError(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return nil, errTest
}

func activateVersionOk(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion}, nil
}

func activateVersionError(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, errTest
}

func listDomainsOk(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{Name: "https://directly-careful-coyote.edgecompute.app"},
	}, nil
}

func listDomainsError(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, errTest
}

func listBackendsOk(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{
		{Name: "foobar"},
	}, nil
}

func listBackendsError(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return nil, errTest
}

type versionClient struct {
	fastlyVersions    []string
	fastlySysVersions []string
}

func (v versionClient) Do(req *http.Request) (*http.Response, error) {
	var vs []string

	if strings.Contains(req.URL.String(), "crates/fastly-sys/") {
		vs = v.fastlySysVersions
	}
	if strings.Contains(req.URL.String(), "crates/fastly/") {
		vs = v.fastlyVersions
	}

	rec := httptest.NewRecorder()

	var versions []string
	for _, vv := range vs {
		versions = append(versions, fmt.Sprintf(`{"num":"%s"}`, vv))
	}

	_, err := rec.Write([]byte(fmt.Sprintf(`{"versions":[%s]}`, strings.Join(versions, ","))))
	if err != nil {
		return nil, err
	}
	return rec.Result(), nil
}
