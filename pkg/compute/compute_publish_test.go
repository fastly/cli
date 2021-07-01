package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPublish(t *testing.T) {
	for _, testcase := range []struct {
		name              string
		args              []string
		applicationConfig config.File
		fastlyManifest    string
		cargoManifest     string
		cargoLock         string
		client            api.HTTPClient
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
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
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
				"Deployed package (service 123, version 3)",
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
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
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
				"Deployed package (service 123, version 3)",
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
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
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
				"Deployed package (service 123, version 4)",
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

			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, testcase.api, &stdout)
			ara.SetFile(testcase.applicationConfig)
			err = app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
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

	// DEPLOY REQUIREMENTS

	for _, filename := range [][]string{
		{"pkg", "package.tar.gz"},
	} {
		fromFilename := filepath.Join("testdata", "deploy", filepath.Join(filename...))
		toFilename := filepath.Join(rootdir, filepath.Join(filename...))
		testutil.CopyFile(t, fromFilename, toFilename)
	}

	return rootdir
}
