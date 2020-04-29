package compute_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
			rootdir := makeInitEnvironment(t)
			defer os.RemoveAll(rootdir)

			t.Logf("Temporary init environment: %s", rootdir)

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

func TestBuild(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_BUILD") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_BUILD to run this test")
	}

	for _, testcase := range []struct {
		name               string
		args               []string
		manifest           string
		wantError          string
		wantOutputContains string
	}{
		{
			name:      "no fastly.toml manifest",
			args:      []string{"compute", "build"},
			wantError: "error reading package manifest: open fastly.toml:", // actual message differs on Windows
		},
		{
			name:      "empty language",
			args:      []string{"compute", "build"},
			manifest:  "name = \"test\"\n",
			wantError: "language cannot be empty, please provide a language",
		},
		{
			name:      "empty name",
			args:      []string{"compute", "build"},
			manifest:  "language = \"rust\"\n",
			wantError: "name cannot be empty, please provide a name",
		},
		{
			name:      "unknown language",
			args:      []string{"compute", "build"},
			manifest:  "name = \"test\"\nlanguage = \"javascript\"\n",
			wantError: "unsupported language javascript",
		},
		{
			name:               "success",
			args:               []string{"compute", "build"},
			manifest:           "name = \"test\"\nlanguage = \"rust\"\n",
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
			rootdir := makeBuildEnvironment(t, testcase.manifest)
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
		client           api.HTTPClient
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
			name:      "path with no service ID",
			args:      []string{"compute", "deploy", "-p", "pkg/package.tar.gz"},
			manifest:  "name = \"package\"\n",
			wantError: "error reading service: no service ID found. Please provide one via the --service-id flag or within your package manifest",
			wantOutput: []string{
				"Validating package...",
			},
		},
		{
			name:      "empty service ID",
			args:      []string{"compute", "deploy"},
			manifest:  "name = \"package\"\n",
			wantError: "error reading service: no service ID found. Please provide one via the --service-id flag or within your package manifest",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
			},
		},
		{
			name:      "latest version error",
			args:      []string{"compute", "deploy"},
			api:       mock.API{ListVersionsFn: listVersionsError},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error getting listing service versions: fixture error",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
			},
		},
		{
			name: "clone version error",
			args: []string{"compute", "deploy"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
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
			name: "no token",
			args: []string{"compute", "deploy"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "no token provided",
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
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
			},
			client:    errorClient{err: errors.New("some network failure")},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error executing API request: some network failure",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
				"Uploading package...",
			},
		},
		{
			name: "package API error",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
			},
			client:    errorClient{err: errors.New("some network failure")},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error executing API request: some network failure",
			wantOutput: []string{
				"Reading package manifest...",
				"Validating package...",
				"Fetching latest version...",
				"Cloning latest version...",
				"Uploading package...",
			},
		},
		{
			name: "package API server error",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				CloneVersionFn: cloneVersionOk,
			},
			client:    codeClient{http.StatusInternalServerError},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: "error from API: 500 Internal Server Error",
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
				CloneVersionFn:    cloneVersionOk,
				ActivateVersionFn: activateVersionError,
			},
			client:    codeClient{http.StatusOK},
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
				CloneVersionFn:    cloneVersionOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsError,
			},
			client:   codeClient{http.StatusOK},
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
			name: "success",
			args: []string{"compute", "deploy", "-t", "123"},
			api: mock.API{
				ListVersionsFn:    listVersionsActiveOk,
				CloneVersionFn:    cloneVersionOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			client:   codeClient{http.StatusOK},
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
				CloneVersionFn:    cloneVersionOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			client: codeClient{http.StatusOK},
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
				CloneVersionFn:    cloneVersionOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			client: codeClient{http.StatusOK},
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
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			client:           codeClient{http.StatusOK},
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
				httpClient                     = testcase.client
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
		client     api.HTTPClient
		wantError  string
		wantOutput []string
	}{
		{
			name:      "no token",
			args:      []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz"},
			wantError: "no token provided",
			wantOutput: []string{
				"Initializing...",
			},
		},
		{
			name:       "invalid package path",
			args:       []string{"compute", "update", "-s", "123", "--version", "1", "-p", "unkown.tar.gz", "-t", "123"},
			wantError:  "error reading package: ",
			wantOutput: []string{},
		},
		{
			name:      "package API error",
			args:      []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz", "-t", "123"},
			client:    errorClient{err: errors.New("some network failure")},
			wantError: "error executing API request: some network failure",
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
			},
		},
		{
			name:      "package API server error",
			args:      []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz", "-t", "123"},
			client:    codeClient{http.StatusInternalServerError},
			wantError: "error from API: 500 Internal Server Error",
			wantOutput: []string{
				"Initializing...",
				"Uploading package...",
			},
		},
		{
			name:   "success",
			args:   []string{"compute", "update", "-s", "123", "--version", "1", "-p", "pkg/package.tar.gz", "-t", "123"},
			client: codeClient{http.StatusOK},
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
				clientFactory                  = mock.APIClient(mock.API{})
				httpClient                     = testcase.client
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

func TestUploadPackage(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		client    *compute.Client
		serviceID string
		version   int
		path      string
		wantError string
	}{
		{
			name:      "no package",
			client:    compute.NewClient(codeClient{http.StatusOK}, "", ""),
			serviceID: "123",
			version:   1,
			path:      "pkg/unkown.pkg.tar.gz",
			wantError: "error reading package",
		},
		{
			name:      "package API error",
			client:    compute.NewClient(errorClient{err: errors.New("some network failure")}, "", ""),
			serviceID: "123",
			version:   1,
			path:      "pkg/package.tar.gz",
			wantError: "error executing API request: some network failure",
		},
		{
			name:      "package API not found",
			client:    compute.NewClient(codeClient{http.StatusNotFound}, "", ""),
			serviceID: "123",
			version:   1,
			path:      "pkg/package.tar.gz",
			wantError: "error from API: 404 Not Found",
		},
		{
			name:      "package API server error",
			client:    compute.NewClient(codeClient{http.StatusInternalServerError}, "", ""),
			serviceID: "123",
			version:   1,
			path:      "pkg/package.tar.gz",
			wantError: "error from API: 500 Internal Server Error",
		},
		{
			name:      "success",
			client:    compute.NewClient(codeClient{http.StatusOK}, "", ""),
			serviceID: "123",
			version:   1,
			path:      "pkg/package.tar.gz",
			wantError: "",
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

			err = testcase.client.UpdatePackage(testcase.serviceID, testcase.version, testcase.path)
			testutil.AssertErrorContains(t, err, testcase.wantError)
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
				{Number: 1, Active: false},
				{Number: 2, Active: true},
				{Number: 3, Active: false},
			},
			wantVersion: 2,
		},
		{
			name: "active not latest",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false},
				{Number: 2, Active: true},
				{Number: 3, Active: false},
			},
			wantVersion: 2,
		},
		{
			name: "active and locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false},
				{Number: 2, Active: true},
				{Number: 3, Active: false, Locked: true},
			},
			wantVersion: 2,
		},
		{
			name: "locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false},
				{Number: 2, Active: false, Locked: true},
			},
			wantVersion: 2,
		},
		{
			name: "locked not latest",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false},
				{Number: 2, Active: false, Locked: true},
				{Number: 3, Active: false},
			},
			wantVersion: 2,
		},
		{
			name: "no active or locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false},
				{Number: 2, Active: false},
				{Number: 3, Active: false},
			},
			wantVersion: 3,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v, err := compute.GetLatestIdealVersion(testcase.inputVersions)
			testutil.AssertNoError(t, err)
			if v.Number != testcase.wantVersion {
				t.Errorf("wanted version %d, got %d", testcase.wantVersion, v.Number)
			}
		})
	}
}

func makeInitEnvironment(t *testing.T) (rootdir string) {
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

	return rootdir
}

func makeBuildEnvironment(t *testing.T, manifestContent string) (rootdir string) {
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
		[]string{"Cargo.toml"},
		[]string{"Cargo.lock"},
		[]string{"src", "main.rs"},
	} {
		fromFilename := filepath.Join("testdata", "build", filepath.Join(filename...))
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

func makeDeployEnvironment(t *testing.T, manifestContent string) (rootdir string) {
	t.Helper()

	p := make([]byte, 8)
	n, err := rand.Read(p)
	if err != nil {
		t.Fatal(err)
	}

	rootdir = filepath.Join(
		os.TempDir(),
		fmt.Sprintf("fastly-deploy-%x", p[:n]),
	)

	if err := os.MkdirAll(rootdir, 0700); err != nil {
		t.Fatal(err)
	}

	for _, filename := range [][]string{
		[]string{"pkg", "package.tar.gz"},
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
		{ServiceID: i.Service, Number: 1, Active: false},
		{ServiceID: i.Service, Number: 2, Active: false},
	}, nil
}

func listVersionsActiveOk(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{ServiceID: i.Service, Number: 1, Active: true},
		{ServiceID: i.Service, Number: 2, Active: false, Locked: true},
	}, nil
}

func listVersionsError(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return nil, errTest
}

func cloneVersionOk(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.Service, Number: i.Version + 1}, nil
}

func cloneVersionError(i *fastly.CloneVersionInput) (*fastly.Version, error) {
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
		&fastly.Domain{Name: "https://directly-careful-coyote.edgecompute.app"},
	}, nil
}

func listDomainsError(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, errTest
}

type errorClient struct {
	err error
}

func (c errorClient) Do(*http.Request) (*http.Response, error) {
	return nil, c.err
}

type codeClient struct {
	code int
}

func (c codeClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(c.code)
	return rec.Result(), nil
}
