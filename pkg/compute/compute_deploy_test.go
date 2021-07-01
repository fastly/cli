package compute_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

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
			name:      "no token",
			args:      []string{"compute", "deploy"},
			wantError: "no token provided",
		},
		{
			name:      "no fastly.toml manifest",
			args:      []string{"compute", "deploy", "--token", "123"},
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
				ListVersionsFn: testutil.ListVersionsError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: fmt.Sprintf("error listing service versions: %s", testutil.Err.Error()),
		},
		{
			name: "service version is active, clone version error",
			args: []string{"compute", "deploy", "--token", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: fmt.Sprintf("error cloning service version: %s", testutil.Err.Error()),
		},
		{
			name: "list domains error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: fmt.Sprintf("error fetching service domains: %s", testutil.Err.Error()),
		},
		{
			name: "list backends error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOk,
				ListBackendsFn: listBackendsError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: fmt.Sprintf("error fetching service backends: %s", testutil.Err.Error()),
		},
		// The following test doesn't just validate the package API error behaviour
		// but as a side effect it validates that when deleting the created
		// service, the Service ID is also cleared out from the manifest.
		{
			name: "package API error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				ListVersionsFn:  testutil.ListVersions,
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
			wantError: fmt.Sprintf("error uploading package: %s", testutil.Err.Error()),
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
			api: mock.API{
				CreateServiceFn: createServiceError,
			},
			manifest:  "name = \"package\"\n",
			wantError: fmt.Sprintf("error creating service: %s", testutil.Err.Error()),
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
			api: mock.API{
				GetServiceFn:    getServiceOK,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainError,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			manifest:  "name = \"package\"\n",
			wantError: fmt.Sprintf("error creating domain: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Creating service...",
				"Creating domain...",
			},
		},
		// The following test mocks the backend API call to fail, and we expect to
		// see a relevant error message related to that error.
		//
		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. We mock the service creation to be successful while we
		// mock the backend API call to fail, and we expect to see a relevant error
		// message related to that error.
		{
			name: "service backend error",
			args: []string{"compute", "deploy", "--token", "123"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendError,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			manifest:  "name = \"package\"\n",
			wantError: fmt.Sprintf("error creating backend: %s", testutil.Err.Error()),
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
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionError,
			},
			manifest:  "name = \"package\"\nservice_id = \"123\"\n",
			wantError: fmt.Sprintf("error activating version: %s", testutil.Err.Error()),
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
				ListVersionsFn: testutil.ListVersions,
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
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
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
				"Deployed package (service 123, version 3)",
			},
		},
		{
			name: "success with path",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123", "--version", "latest"},
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
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
				"Deployed package (service 123, version 3)",
			},
		},
		{
			name: "success with inactive version",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123"},
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
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
				"Deployed package (service 123, version 3)",
			},
		},
		{
			name: "success with specific locked version",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
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
				"Deployed package (service 123, version 4)",
			},
		},
		{
			name: "success with active version",
			args: []string{"compute", "deploy", "--token", "123", "-p", "pkg/package.tar.gz", "-s", "123", "--version", "active"},
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
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

			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, &stdout)
			ara.SetClientFactory(testcase.api)

			// we need to define stdin as the deploy process prompts the user multiple
			// times, but we don't need to provide any values as all our prompts will
			// fallback to default values if the input is unrecognised.
			ara.SetStdin(strings.NewReader(""))

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
		testutil.CopyFile(t, fromFilename, toFilename)
	}

	if manifestContent != "" {
		filename := filepath.Join(rootdir, compute.ManifestFilename)
		if err := os.WriteFile(filename, []byte(manifestContent), 0777); err != nil {
			t.Fatal(err)
		}
	}

	return rootdir
}
