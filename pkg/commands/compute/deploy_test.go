package compute_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestDeploy(t *testing.T) {
	// NOTE: Some tests don't provide a Service ID via any mechanism (e.g. flag
	// or manifest) and if one is provided the test will fail due to a specific
	// API call not being mocked. Be careful not to add a Service ID to all tests
	// without first checking whether the Service ID is expected as the user flow
	// for when no Service ID is provided is to create a new service.

	// We're going to chdir to a deploy environment,
	// so save the PWD to return to, afterwards.
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// Create test environment
	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T: t,
		Copy: []testutil.FileIO{
			{
				Src: filepath.Join("testdata", "deploy", "pkg", "package.tar.gz"),
				Dst: filepath.Join("pkg", "package.tar.gz"),
			},
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

	args := testutil.Args
	for _, testcase := range []struct {
		name             string
		args             []string
		api              mock.API
		wantError        string
		wantOutput       []string
		manifestIncludes string
		noManifest       bool
	}{
		{
			name:      "no token",
			args:      args("compute deploy"),
			wantError: "no token provided",
		},
		{
			name:       "no fastly.toml manifest",
			args:       args("compute deploy --token 123"),
			wantError:  "error reading package manifest",
			noManifest: true,
		},
		{
			// If no Service ID defined via flag or manifest, then the expectation is
			// for the service to be created via the API and for the returned ID to
			// be stored into the manifest.
			//
			// Additionally it validates that the specified path (files generated by
			// the testutil.NewEnv()) cause no issues.
			name: "path with no service ID",
			args: args("compute deploy --token 123 -v -p pkg/package.tar.gz"),
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
		// Same validation as above with the exception that we use the default path
		// parsing logic (i.e. we don't explicitly pass a path via `-p` flag).
		{
			name: "empty service ID",
			args: args("compute deploy --token 123 -v"),
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
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: testutil.ListVersionsError,
			},
			wantError: fmt.Sprintf("error listing service versions: %s", testutil.Err.Error()),
		},
		{
			name: "service version is active, clone version error",
			args: args("compute deploy --service-id 123 --token 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionError,
			},
			wantError: fmt.Sprintf("error cloning service version: %s", testutil.Err.Error()),
		},
		{
			name: "list domains error",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsError,
			},
			wantError: fmt.Sprintf("error fetching service domains: %s", testutil.Err.Error()),
		},
		{
			name: "list backends error",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOk,
				ListBackendsFn: listBackendsError,
			},
			wantError: fmt.Sprintf("error fetching service backends: %s", testutil.Err.Error()),
		},
		// The following test doesn't just validate the package API error behaviour
		// but as a side effect it validates that when deleting the created
		// service, the Service ID is also cleared out from the manifest.
		{
			name: "package API error",
			args: args("compute deploy --token 123"),
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
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn: createServiceError,
			},
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
			args: args("compute deploy --token 123"),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainError,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			wantError: fmt.Sprintf("error creating domain: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Creating service...",
				"Creating domain...",
			},
		},
		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. We mock the service creation to be successful while we
		// mock the backend API call to fail.
		{
			name: "service backend error",
			args: args("compute deploy --token 123"),
			api: mock.API{
				GetServiceFn:    getServiceOK,
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendError,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
			},
			wantError: fmt.Sprintf("error creating backend: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Creating service...",
				"Creating domain...",
				"Creating backend...",
			},
		},
		// The following test validates that the undoStack is executed as expected
		// e.g. the backend and domain resources are deleted.
		{
			name: "activate error",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionError,
			},
			wantError: fmt.Sprintf("error activating version: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
			},
		},
		{
			name: "identical package",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				GetServiceFn:   getServiceOK,
				ListVersionsFn: testutil.ListVersions,
				ListDomainsFn:  listDomainsOk,
				ListBackendsFn: listBackendsOk,
				GetPackageFn:   getPackageIdentical,
			},
			wantOutput: []string{
				"Skipping package deployment",
			},
		},
		{
			name: "success",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
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
			args: args("compute deploy --service-id 123 --token 123 -p pkg/package.tar.gz --version latest"),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
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
			args: args("compute deploy --service-id 123 --token 123 -p pkg/package.tar.gz"),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 3)",
			},
		},
		{
			name: "success with specific locked version",
			args: args("compute deploy --service-id 123 --token 123 -p pkg/package.tar.gz --version 2"),
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
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 4)",
			},
		},
		{
			name: "success with active version",
			args: args("compute deploy --service-id 123 --token 123 -p pkg/package.tar.gz --version active"),
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
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 4)",
			},
		},
		{
			name: "success with comment",
			args: args("compute deploy --service-id 123 --token 123 -p pkg/package.tar.gz --version 2 --comment foo"),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				UpdateVersionFn:   updateVersionOk,
			},
			wantOutput: []string{
				"Uploading package...",
				"Activating version...",
				"Deployed package (service 123, version 4)",
			},
		},
		{
			name: "success with --backend and no --backend-port",
			args: args("compute deploy --backend host.com --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendExpect("host.com", 80, "", ""),
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
		},
		{
			name: "success with --backend and --backend-port",
			args: args("compute deploy --backend host.com --backend-port 443 --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendExpect("host.com", 443, "", ""),
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
		},
		{
			name: "success with --backend and --override-host",
			args: args("compute deploy --backend host.com --override-host otherhost.com --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendExpect("host.com", 80, "otherhost.com", ""),
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
		},
		{
			name: "success with --backend and --ssl-sni-hostname",
			args: args("compute deploy --backend host.com --ssl-sni-hostname anotherhost.com --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendExpect("host.com", 80, "", "anotherhost.com"),
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
		},
		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. Our fastly.toml is configured with a [setup] section so
		// we expect to see the appropriate messaging in the output.
		{
			name: "success with setup configuration",
			args: args("compute deploy --token 123"),
			api: mock.API{
				GetServiceFn:      getServiceOK,
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				DeleteBackendFn:   deleteBackendOK,
				DeleteDomainFn:    deleteDomainOK,
				DeleteServiceFn:   deleteServiceOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
			},
			manifest: `
			name = "package"
			manifest_version = 1
			language = "rust"

			[setup]
				[[setup.backends]]
					name = "backend_name"
					prompt = "Backend 1"
					address = "developer.fastly.com"
					port = 443
				[[setup.backends]]
					name = "other_backend_name"
					prompt = "Backend 2"
					address = "httpbin.org"
					port = 443
			`,
			wantOutput: []string{
				"Initializing...",
				"Creating service...",
				"Creating domain...",
				"Creating backend 'developer.fastly.com'...",
				"Creating backend 'httpbin.org'...",
				"Uploading package...",
				"Activating version...",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
		},
		{
			name: "error with setup configuration -- missing setup.backends.name",
			args: args("compute deploy --token 123"),
			api: mock.API{
				GetServiceFn: getServiceOK,
			},
			manifest: `
			name = "package"
			manifest_version = 1
			language = "rust"

			[setup]
				[[setup.backends]]
					prompt = "Backend 1"
					address = "developer.fastly.com"
					port = 443
				[[setup.backends]]
					prompt = "Backend 2"
					address = "httpbin.org"
					port = 443
			`,
			wantError: "error parsing the [[setup.backends]] configuration",
		},
		// The 'name' field should be a string, not an integer
		{
			name: "error with setup configuration -- invalid setup.backends.name",
			args: args("compute deploy --token 123"),
			api: mock.API{
				GetServiceFn: getServiceOK,
			},
			manifest: `
			name = "package"
			manifest_version = 1
			language = "rust"

			[setup]
				[[setup.backends]]
				  name = 123
					prompt = "Backend 1"
					address = "developer.fastly.com"
					port = 443
				[[setup.backends]]
				  name = 456
					prompt = "Backend 2"
					address = "httpbin.org"
					port = 443
			`,
			wantError: "error parsing the [[setup.backends]] configuration",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// Because the manifest can be mutated on each test scenario, we recreate
			// the file each time.
			if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(`name = "package"`), 0777); err != nil {
				t.Fatal(err)
			}

			// For any test scenario that expects no manifest to exist, then instead
			// of deleting the manifest and having to recreate it, we'll simply
			// rename it, and then rename it back once the specific test scenario has
			// finished running.
			if testcase.noManifest {
				old := filepath.Join(rootdir, manifest.Filename)
				tmp := filepath.Join(rootdir, manifest.Filename+"Tmp")
				if err := os.Rename(old, tmp); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := os.Rename(tmp, old); err != nil {
						t.Fatal(err)
					}
				}()
			}

			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)

			// we need to define stdin as the deploy process prompts the user multiple
			// times, but we don't need to provide any values as all our prompts will
			// fallback to default values if the input is unrecognised.
			opts.Stdin = strings.NewReader("")

			err = app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)

			for _, s := range testcase.wantOutput {
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

func createServiceOK(i *fastly.CreateServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:   "12345",
		Name: i.Name,
		Type: i.Type,
	}, nil
}

func createServiceError(*fastly.CreateServiceInput) (*fastly.Service, error) {
	return nil, testutil.Err
}

func deleteServiceOK(i *fastly.DeleteServiceInput) error {
	return nil
}

func createDomainError(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return nil, testutil.Err
}

func deleteDomainOK(i *fastly.DeleteDomainInput) error {
	return nil
}

func createBackendError(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return nil, testutil.Err
}

func deleteBackendOK(i *fastly.DeleteBackendInput) error {
	return nil
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

func activateVersionError(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func listDomainsError(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, testutil.Err
}

func listBackendsError(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return nil, testutil.Err
}

func createBackendExpect(address string, port uint, overrideHost string, sslSNIHostname string) func(*fastly.CreateBackendInput) (*fastly.Backend, error) {
	return func(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
		if address != i.Address || port != i.Port || i.OverrideHost != overrideHost || i.SSLSNIHostname != sslSNIHostname {
			return nil, testutil.Err
		}
		return &fastly.Backend{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           i.Name,
		}, nil
	}
}
