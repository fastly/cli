package compute_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
	"github.com/fastly/go-fastly/fastly"
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
			name:      "no token",
			args:      []string{"compute", "init"},
			wantError: "no token provided",
		},
		{
			name:       "unkown repository",
			args:       []string{"compute", "init", "--from", "https://example.com/template"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
				DeleteServiceFn: deleteServiceOK,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
			},
			wantError: "error fetching package template:",
		},
		{
			name:       "create service error",
			args:       []string{"compute", "init"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceError,
			},
			wantError: "error creating service: fixture error",
		},
		{
			name:       "create domain error",
			args:       []string{"compute", "init"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainError,
				DeleteServiceFn: deleteServiceOK,
			},
			wantError: "error creating domain: fixture error",
		},
		{
			name:       "create backend error",
			args:       []string{"compute", "init"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendError,
				DeleteServiceFn: deleteServiceOK,
				DeleteDomainFn:  deleteDomainOK,
			},
			wantError: "error creating backend: fixture error",
		},
		{
			name:       "with name",
			args:       []string{"compute", "init", "--name", "test"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `name = "test"`,
		},
		{
			name:       "with service",
			args:       []string{"compute", "init", "-s", "test"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				GetServiceFn:    getServiceOK,
				ListVersionsFn:  listVersionsActiveOk,
				CloneVersionFn:  cloneVersionOk,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `name = "test"`,
		},
		{
			name:       "with description",
			args:       []string{"compute", "init", "--description", "test"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `description = "test"`,
		},
		{
			name:       "with author",
			args:       []string{"compute", "init", "--author", "test@example.com"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `authors = ["test@example.com"]`,
		},
		{
			name:       "with multiple authors",
			args:       []string{"compute", "init", "--author", "test1@example.com", "--author", "test2@example.com"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
			manifestIncludes: `authors = ["test1@example.com", "test2@example.com"]`,
		},

		{
			name:       "with from repository and branch",
			args:       []string{"compute", "init", "--from", "https://github.com/fastly/fastly-template-rust-default.git", "--branch", "master"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Fetching package template...",
				"Updating package manifest...",
			},
		},
		{
			name:       "with existing package manifest",
			args:       []string{"compute", "init"},
			configFile: config.File{Token: "123"},
			manifest: strings.Join([]string{
				"service_id = \"1234\"",
				"name = \"test\"",
				"language = \"rust\"",
				"description = \"test\"",
				"authors = [\"test@fastly.com\"]",
			}, "\n"),
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				GetServiceFn:    getServiceOK,
				ListVersionsFn:  listVersionsActiveOk,
				CloneVersionFn:  cloneVersionOk,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: []string{
				"Initializing...",
				"Creating domain...",
				"Creating backend...",
				"Updating package manifest...",
			},
		},
		{
			name: "default",
			args: []string{"compute", "init"},
			configFile: config.File{
				Token: "123",
				Email: "test@example.com",
			},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
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
				"Creating service...",
				"Creating domain...",
				"Creating backend...",
				"Fetching package template...",
				"Updating package manifest...",
			},
		},
		{
			name:       "with default name inferred from directory",
			args:       []string{"compute", "init"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			manifestIncludes: `name = "fastly-init`,
		},
		{
			name:       "with AssemblyScript language",
			args:       []string{"compute", "init", "--language", "assemblyscript"},
			configFile: config.File{Token: "123"},
			api: mock.API{
				GetTokenSelfFn:  tokenOK,
				GetUserFn:       getUserOk,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
			},
			manifestIncludes: `name = "fastly-init`,
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
				versioner     update.Versioner = nil
				in            io.Reader        = bytes.NewBufferString(testcase.stdin)
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, out)
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
				content, err := ioutil.ReadFile(filepath.Join(rootdir, compute.ManifestFilename))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}

func TestBuildRust(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_BUILD_RUST") == "" && os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	for _, testcase := range []struct {
		name                 string
		args                 []string
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
			client:    versionClient{[]string{"0.0.0"}},
			wantError: "error reading package manifest: open fastly.toml:", // actual message differs on Windows
		},
		{
			name:           "empty language",
			args:           []string{"compute", "build"},
			fastlyManifest: "name = \"test\"\n",
			client:         versionClient{[]string{"0.0.0"}},
			wantError:      "language cannot be empty, please provide a language",
		},
		{
			name:           "empty name",
			args:           []string{"compute", "build"},
			fastlyManifest: "language = \"rust\"\n",
			client:         versionClient{[]string{"0.0.0"}},
			wantError:      "name cannot be empty, please provide a name",
		},
		{
			name:           "unknown language",
			args:           []string{"compute", "build"},
			fastlyManifest: "name = \"test\"\nlanguage = \"javascript\"\n",
			client:         versionClient{[]string{"0.0.0"}},
			wantError:      "unsupported language javascript",
		},
		{
			name:           "error reading cargo metadata",
			args:           []string{"compute", "build"},
			fastlyManifest: "name = \"test\"\nlanguage = \"rust\"\n",
			cargoManifest:  "[package]\nname = \"test\"",
			client:         versionClient{[]string{"0.4.0"}},
			wantError:      "reading cargo metadata",
		},
		{
			name:                 "fastly-sys crate not found",
			args:                 []string{"compute", "build"},
			fastlyManifest:       "name = \"test\"\nlanguage = \"rust\"\n",
			cargoManifest:        "[package]\nname = \"test\"\nversion = \"0.1.0\"\n\n[dependencies]\nfastly = \"=0.3.2\"",
			cargoLock:            "[[package]]\nname = \"test\"\nversion = \"0.1.0\"\n\n[[package]]\nname = \"fastly\"\nversion = \"0.3.2\"",
			client:               versionClient{[]string{"0.4.0"}},
			wantError:            "fastly-sys crate not found",
			wantRemediationError: "fastly = \"^0.4.0\"",
		},
		{
			name:                 "fastly-sys crate out-of-date",
			args:                 []string{"compute", "build"},
			fastlyManifest:       "name = \"test\"\nlanguage = \"rust\"\n",
			cargoLock:            "[[package]]\nname = \"fastly-sys\"\nversion = \"0.3.2\"",
			client:               versionClient{[]string{"0.4.0"}},
			wantError:            "fastly crate not up-to-date",
			wantRemediationError: "fastly = \"^0.4.0\"",
		},
		{
			name:               "Rust success",
			args:               []string{"compute", "build"},
			fastlyManifest:     "name = \"test\"\nlanguage = \"rust\"\n",
			cargoLock:          "[[package]]\nname = \"fastly\"\nversion = \"0.3.2\"\n\n[[package]]\nname = \"fastly-sys\"\nversion = \"0.3.2\"",
			client:             versionClient{[]string{"0.0.0"}},
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
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(mock.API{})
				httpClient                     = testcase.client
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, out)
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
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
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
			name:           "empty language",
			args:           []string{"compute", "build"},
			fastlyManifest: "name = \"test\"\n",
			wantError:      "language cannot be empty, please provide a language",
		},
		{
			name:           "empty name",
			args:           []string{"compute", "build"},
			fastlyManifest: "language = \"assemblyscript\"\n",
			wantError:      "name cannot be empty, please provide a name",
		},
		{
			name:           "unknown language",
			args:           []string{"compute", "build"},
			fastlyManifest: "name = \"test\"\nlanguage = \"javascript\"\n",
			wantError:      "unsupported language javascript",
		},
		{
			name:               "AssemblyScript success",
			args:               []string{"compute", "build"},
			fastlyManifest:     "name = \"test\"\nlanguage = \"assemblyscript\"\n",
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
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, out)
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
	}{
		{
			name:      "no fastly.toml manifest",
			args:      []string{"compute", "deploy"},
			wantError: "error reading package manifest",
			wantOutput: []string{
				"Reading package manifest...",
			},
		},
		{
			name:       "path with no service ID",
			args:       []string{"compute", "deploy", "-p", "pkg/package.tar.gz"},
			manifest:   "name = \"package\"\n",
			wantError:  "error reading service: no service ID found. Please provide one via the --service-id flag or within your package manifest",
			wantOutput: []string{
				// "Reading package manifest...",
			},
		},
		{
			name:      "empty service ID",
			args:      []string{"compute", "deploy"},
			manifest:  "name = \"package\"\n",
			wantError: "error reading service: no service ID found. Please provide one via the --service-id flag or within your package manifest",
			wantOutput: []string{
				"Reading package manifest...",
			},
		},
		{
			name:      "latest version error",
			args:      []string{"compute", "deploy"},
			api:       mock.API{ListVersionsFn: listVersionsError},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error listing service versions: fixture error",
			wantOutput: []string{
				"Reading package manifest...",
				"Fetching latest version...",
			},
		},
		{
			name: "clone version error",
			args: []string{"compute", "deploy"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetPackageFn:   getPackageOk,
				CloneVersionFn: cloneVersionError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error cloning latest service version: fixture error",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
			},
		},
		{
			name: "package API error",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn:  listVersionsActiveOk,
				GetPackageFn:    getPackageOk,
				CloneVersionFn:  cloneVersionOk,
				UpdatePackageFn: updatePackageError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error uploading package: fixture error",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
				"Uploading package...",
			},
		},
		{
			name: "activate error",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn:    listVersionsActiveOk,
				GetPackageFn:      getPackageOk,
				CloneVersionFn:    cloneVersionOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error activating version: fixture error",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
				"Uploading package...",
				"Activating version...",
			},
		},
		{
			name: "list domains error",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn:    listVersionsActiveOk,
				GetPackageFn:      getPackageOk,
				CloneVersionFn:    cloneVersionOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsError,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
				"Uploading package...",
				"Activating version...",
				"Manage this service at:",
				"https://manage.fastly.com/configure/services/123",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "indentical package",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetPackageFn:   getPackageIdentical,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Reading package manifest...",
				"Fetching latest version...",
				"Validating package...",
				"Skipping package deployment",
			},
		},
		{
			name: "success",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn:    listVersionsActiveOk,
				GetPackageFn:      getPackageOk,
				CloneVersionFn:    cloneVersionOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			manifest: "name = \"package\"\nservice_id = \"123\"\n",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
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
			args: []string{"compute", "deploy", "-t", "123", "-p", "pkg/package.tar.gz", "-s", "123"},
			api: mock.API{
				ListVersionsFn:    listVersionsActiveOk,
				GetPackageFn:      getPackageOk,
				CloneVersionFn:    cloneVersionOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			wantOutput: []string{
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "success with inactive version",
			args: []string{"compute", "deploy", "-t", "123", "-p", "pkg/package.tar.gz", "-s", "123"},
			api: mock.API{
				ListVersionsFn:    listVersionsInactiveOk,
				GetPackageFn:      getPackageOk,
				CloneVersionFn:    cloneVersionOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			wantOutput: []string{
				"Validating package...",
				"Fetching latest version...",
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 2)",
			},
		},
		{
			name: "success with version",
			args: []string{"compute", "deploy", "-t", "123", "-p", "pkg/package.tar.gz", "-s", "123", "--version", "2"},
			api: mock.API{
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			manifestIncludes: "version = 2",
			wantOutput: []string{
				"Validating package...",
				"Uploading package...",
				"Activating version...",
				"Updating package manifest...",
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
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
			}
			if testcase.manifestIncludes != "" {
				content, err := ioutil.ReadFile(filepath.Join(rootdir, compute.ManifestFilename))
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
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			for _, s := range testcase.wantOutput {
				testutil.AssertStringContains(t, buf.String(), s)
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
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				buf           bytes.Buffer
				out           io.Writer = common.NewSyncWriter(&buf)
			)
			err = app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, buf.String(), testcase.wantOutput)
		})
	}
}

func makeInitEnvironment(t *testing.T, manifestContent string) (rootdir string) {
	t.Helper()

	rootdir, err := ioutil.TempDir("", "fastly-init-*")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(rootdir, 0700); err != nil {
		t.Fatal(err)
	}

	if manifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := ioutil.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}

func makeRustBuildEnvironment(t *testing.T, fastlyManifestContent, cargoManifestContent, cargoLockContent string) (rootdir string) {
	t.Helper()

	rootdir, err := ioutil.TempDir("", "fastly-build-*")
	if err != nil {
		t.Fatal(err)
	}

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
		copyFile(t, fromFilename, toFilename)
	}

	if fastlyManifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := ioutil.WriteFile(filename, []byte(fastlyManifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	if cargoManifestContent != "" {
		filename := filepath.Join(rootdir, "Cargo.toml")
		if err := ioutil.WriteFile(filename, []byte(cargoManifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	if cargoLockContent != "" {
		filename := filepath.Join(rootdir, "Cargo.lock")
		if err := ioutil.WriteFile(filename, []byte(cargoLockContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}

func makeAssemblyScriptBuildEnvironment(t *testing.T, fastlyManifestContent string) (rootdir string) {
	t.Helper()

	rootdir, err := ioutil.TempDir("", "fastly-build-*")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(rootdir, 0700); err != nil {
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
		if err := ioutil.WriteFile(filename, []byte(fastlyManifestContent), 0777); err != nil {
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

	rootdir, err := ioutil.TempDir("", "fastly-deploy-*")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(rootdir, 0700); err != nil {
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
		if err := ioutil.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
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

func tokenOK() (*fastly.Token, error) { return &fastly.Token{}, nil }

func getUserOk(i *fastly.GetUserInput) (*fastly.User, error) {
	return &fastly.User{Login: "test@example.com"}, nil
}

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
		ServiceID: i.Service,
		Version:   i.Version,
		Name:      i.Name,
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
		ServiceID: i.Service,
		Version:   i.Version,
		Name:      i.Name,
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
			ServiceID: i.Service,
			Number:    1,
			Active:    false,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.Service,
			Number:    2,
			Active:    false,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func listVersionsActiveOk(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.Service,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.Service,
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
	return &fastly.Package{ServiceID: i.Service, Version: i.Version}, nil
}

func getPackageIdentical(i *fastly.GetPackageInput) (*fastly.Package, error) {
	return &fastly.Package{
		ServiceID: i.Service,
		Version:   i.Version,
		Metadata: fastly.PackageMetadata{
			HashSum: "2b742f99854df7e024c287e36fb0fdfc5414942e012be717e52148ea0d6800d66fc659563f6f11105815051e82b14b61edc84b33b49789b790db1ed3446fb483",
		},
	}, nil
}

func cloneVersionOk(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.Service, Number: i.Version + 1}, nil
}

func cloneVersionError(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return nil, errTest
}

func updatePackageOk(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return &fastly.Package{ServiceID: i.Service, Version: i.Version}, nil
}

func updatePackageError(i *fastly.UpdatePackageInput) (*fastly.Package, error) {
	return nil, errTest
}

func activateVersionOk(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.Service, Number: i.Version}, nil
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
