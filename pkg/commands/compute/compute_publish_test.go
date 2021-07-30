package compute_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPublish(t *testing.T) {
	// We're going to chdir to a temp environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	fastlyManifest := `
		manifest_version = 1
		service_id = "123"
		name = "test"
		language = "rust"`

	cargoManifest := `
		[package]
		name = "test"
		version = "0.1.0"

		[dependencies]
		fastly = "=0.6.0"`

	cargoLock := `
		[[package]]
		name = "fastly"
		version = "0.6.0"

		[[package]]
		name = "fastly-sys"
		version = "0.3.7"`

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
			{Src: fastlyManifest, Dst: manifest.Filename},
			{Src: cargoManifest, Dst: "Cargo.toml"},
			{Src: cargoLock, Dst: "Cargo.lock"},
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

	api := mock.API{
		ActivateVersionFn: activateVersionOk,
		CreateBackendFn:   createBackendOK,
		CreateDomainFn:    createDomainOK,
		GetPackageFn:      getPackageOk,
		GetServiceFn:      getServiceOK,
		ListBackendsFn:    listBackendsOk,
		ListDomainsFn:     listDomainsOk,
		ListVersionsFn:    testutil.ListVersions,
		UpdatePackageFn:   updatePackageOk,
	}

	configFile := config.File{
		Language: config.Language{
			Rust: config.Rust{
				ToolchainVersion:    "1.49.0",
				ToolchainConstraint: ">= 1.49.0 < 2.0.0",
				WasmWasiTarget:      "wasm32-wasi",
				FastlySysConstraint: ">= 0.3.0 <= 0.6.0",
				RustupConstraint:    ">= 1.23.0",
			},
		},
	}

	wantOutput := []string{
		"Built rust package test",
		"Uploading package...",
		"Activating version...",
		"Manage this service at:",
		"https://manage.fastly.com/configure/services/123",
		"View this service at:",
		"https://directly-careful-coyote.edgecompute.app",
	}

	args := testutil.Args
	for _, testcase := range []struct {
		api              mock.API
		args             []string
		manifestIncludes string
		name             string
		wantError        string
		wantOutput       []string
	}{
		{
			name: "success no command flags",
			args: args("compute publish -t 123"),
		},
		{
			name: "success with build command flags",
			args: args("compute publish -t 123 --name test --language rust --include-source --force"),
		},
		{
			name: "success with deploy command flags",
			args: args("compute publish -t 123 --version 2 --path pkg/test.tar.gz"),
			api: mock.API{
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantOutput: []string{
				"Deployed package (service 123, version 4)",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			testcase.args = append(testcase.args, "--verbose") // verbose has a side effect of avoiding spinners when the test fails in CI
			testcase.args = append(testcase.args, "--timeout", "120")
			api.CloneVersionFn = testcase.api.CloneVersionFn
			opts := testutil.NewRunOpts(testcase.args, io.MultiWriter(&stdout, testutil.LogWriter{T: t}))
			opts.APIClient = mock.APIClient(api)
			opts.ConfigFile = configFile
			err = app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)

			wantOutput = append(wantOutput, testcase.wantOutput...)
			for _, s := range wantOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}
			if testcase.manifestIncludes != "" {
				content, err := os.ReadFile(filepath.Join(rootdir, manifest.Filename))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}
