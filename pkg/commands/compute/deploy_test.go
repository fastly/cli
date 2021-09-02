package compute_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NOTE: Some tests don't provide a Service ID via any mechanism (e.g. flag
// or manifest) and if one is provided the test will fail due to a specific
// API call not being mocked. Be careful not to add a Service ID to all tests
// without first checking whether the Service ID is expected as the user flow
// for when no Service ID is provided is to create a new service.
//
// Additionally, stdin can be mocked in one of two ways...
//
// 1. Provide a single value.
// 2. Provide multiple values (one for each prompt expected).
//
// In the first case, the first prompt given to the user will get the value you
// defined in the testcase.stdin field, all other prompts will get an empty
// value. This has worked fine for the most part as the prompts have
// historically provided default values when an empty value is encountered.
//
// The second case is to address running the test code successfully as the
// business logic has changed over time to now 'require' values to be provided
// for some prompts, this means an empty string will break the test flow. If
// that's what you're encountering, then you should add multiple values for the
// testcase.stdin field so that there is a value provided for every prompt your
// testcase user flow expects to encounter.
func TestDeploy(t *testing.T) {
	if os.Getenv("TEST_COMPUTE_DEPLOY") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_DEPLOY to run this test")
	}

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
		api                  mock.API
		args                 []string
		dontWantOutput       []string
		manifest             string
		name                 string
		noManifest           bool
		reduceSizeLimit      bool
		stdin                []string
		wantError            string
		wantRemediationError string
		wantOutput           []string
	}{
		{
			name:      "no token",
			args:      args("compute deploy"),
			wantError: "no token provided",
		},
		{
			name:                 "package size too large",
			args:                 args("compute deploy -p pkg/package.tar.gz --token 123"),
			reduceSizeLimit:      true,
			wantError:            "package size is too large",
			wantRemediationError: errors.PackageSizeRemediation,
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
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantOutput: []string{
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
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantOutput: []string{
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
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendOK,
				GetPackageFn:    getPackageOk,
				UpdatePackageFn: updatePackageError,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
				ListDomainsFn:   listDomainsOk,
				ListBackendsFn:  listBackendsOk,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantError: fmt.Sprintf("error uploading package: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Uploading package...",
			},
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
			stdin: []string{
				"Y", // when prompted to create a new service
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
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainError,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
				ListDomainsFn:   listDomainsNone,
				ListBackendsFn:  listBackendsOk,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
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
				CreateServiceFn: createServiceOK,
				CreateDomainFn:  createDomainOK,
				CreateBackendFn: createBackendError,
				DeleteBackendFn: deleteBackendOK,
				DeleteDomainFn:  deleteDomainOK,
				DeleteServiceFn: deleteServiceOK,
				ListDomainsFn:   listDomainsOk,
				ListBackendsFn:  listBackendsNone,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantError: fmt.Sprintf("error configuring the service: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Creating service...",
			},
			dontWantOutput: []string{
				"Creating domain...",
			},
		},
		// The following test validates that the undoStack is executed as expected
		// e.g. the backend and domain resources are deleted.
		{
			name: "activate error",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetServiceFn:      getServiceOK,
				ListDomainsFn:     listDomainsNone,
				ListBackendsFn:    listBackendsOk,
				CreateDomainFn:    createDomainOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionError,
				DeleteDomainFn:    deleteDomainOK,
			},
			wantError: fmt.Sprintf("error activating version: %s", testutil.Err.Error()),
			wantOutput: []string{
				"Creating domain...",
				"Uploading package...",
				"Activating version...",
			},
		},
		{
			name: "identical package",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetServiceFn:   getServiceOK,
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
				ListVersionsFn:    testutil.ListVersions,
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
				ListVersionsFn:    testutil.ListVersions,
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
				ListVersionsFn:    testutil.ListVersions,
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
		// The following test doesn't provide a Service ID by either a flag nor the
		// manifest, so this will result in the deploy script attempting to create
		// a new service. Our fastly.toml is configured with a [setup] section so
		// we expect to see the appropriate messaging in the output.
		{
			name: "success with setup configuration",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
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
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantOutput: []string{
				"Backend 1: [developer.fastly.com]",
				"Backend port number: [443]",
				"Backend 2: [httpbin.org]",
				"Backend port number: [443]",
				"Creating service...",
				"Creating backend 'backend_name' (host: developer.fastly.com, port: 443)...",
				"Creating backend 'other_backend_name' (host: httpbin.org, port: 443)...",
				"Uploading package...",
				"Activating version...",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
		},
		// The following [setup] configuration doesn't define any prompts, nor any
		// ports, so we validate that the user prompts match our default expectations.
		{
			name: "success with setup configuration and no prompts or ports defined",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
			},
			manifest: `
			name = "package"
			manifest_version = 1
			language = "rust"

			[setup]
				[[setup.backends]]
					name = "foo_backend"
					address = "developer.fastly.com"
				[[setup.backends]]
					name = "bar_backend"
					address = "httpbin.org"
			`,
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantOutput: []string{
				"Backend for 'foo_backend': [developer.fastly.com]",
				"Backend port number: [80]",
				"Backend for 'bar_backend': [httpbin.org]",
				"Backend port number: [80]",
				"Creating service...",
				"Creating backend 'foo_backend' (host: developer.fastly.com, port: 80)...",
				"Creating backend 'bar_backend' (host: httpbin.org, port: 80)...",
				"Uploading package...",
				"Activating version...",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
			dontWantOutput: []string{
				"Creating domain...",
			},
		},
		// The following test validates no prompts are displayed to the user due to
		// the use of the --accept-defaults flag.
		{
			name: "success with setup configuration and accept-defaults",
			args: args("compute deploy --accept-defaults --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsOk,
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
				"Creating backend 'backend_name' (host: developer.fastly.com, port: 443)...",
				"Creating backend 'other_backend_name' (host: httpbin.org, port: 443)...",
				"Uploading package...",
				"Activating version...",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
			dontWantOutput: []string{
				"Backend 1: [developer.fastly.com]",
				"Backend port number: [443]",
				"Backend 2: [httpbin.org]",
				"Backend port number: [443]",
				"Domain: [",
			},
		},
		// The follow test validates the setup.backends.address field is a required
		// field. This is because we need an address to generate a name (if no name
		// was provided by the user).
		{
			name: "error with setup configuration and missing required fields",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn: createServiceOK,
				ListDomainsFn:   listDomainsOk,
				ListBackendsFn:  listBackendsNone,

				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			manifest: `
			name = "package"
			manifest_version = 1
			language = "rust"

			[setup]
				[[setup.backends]]
					prompt = "Backend 1"
					port = 443
				[[setup.backends]]
					prompt = "Backend 2"
					port = 443
			`,
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantError: "error parsing the [[setup.backends]] configuration",
		},
		// The following test validates the setup.backends.name field should be a
		// string, not an integer.
		{
			name: "error with setup configuration -- invalid setup.backends.name",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn: createServiceOK,
				ListDomainsFn:   listDomainsOk,
				ListBackendsFn:  listBackendsOk,
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
			stdin: []string{
				"Y", // when prompted to create a new service
			},
			wantError: "error parsing the [[setup.backends]] configuration",
		},
		// The following test validates that a new 'originless' backend is created
		// when the user has no [setup] configuration and they also pass the
		// --accept-defaults flag. This is done by ensuring we DON'T see the
		// standard 'Creating backend' output because we want to conceal the fact
		// that we require a backend for compute services because it's a temporary
		// implementation detail.
		{
			name: "success with no setup configuration and --accept-defaults for new service creation",
			args: args("compute deploy --accept-defaults --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsNone,
			},
			wantOutput: []string{
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
			dontWantOutput: []string{
				"Creating backend", // expect originless creation to be hidden
			},
		},
		{
			name: "success with no setup configuration and single backend entered at prompt for new service",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsNone,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
				"fastly.com",
				"443",
				"my_backend_name",
				"", // this stops prompting for backends
			},
			wantOutput: []string{
				"Backend (hostname or IP address, or leave blank to stop adding backends):",
				"Backend port number: [80]",
				"Backend name:",
				"Creating backend 'my_backend_name' (host: fastly.com, port: 443)...",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
		},
		// This is the same test as above but when prompted it will provide two
		// backends instead of one, and will also allow the code to generate the
		// backend name using its predefined formula.
		{
			name: "success with no setup configuration and multiple backends entered at prompt for new service",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsNone,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
				"fastly.com",
				"443",
				"", // this is so we generate a backend name using a built-in formula
				"google.com",
				"123",
				"", // this is so we generate a backend name using a built-in formula
				"", // this stops prompting for backends
			},
			wantOutput: []string{
				"Backend (hostname or IP address, or leave blank to stop adding backends):",
				"Backend port number: [80]",
				"Backend name:",
				"Creating backend 'backend_1' (host: fastly.com, port: 443)...",
				"Creating backend 'backend_2' (host: google.com, port: 123)...",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
		},
		// The following test validates that when prompting the user for backends
		// that we'll default to creating an 'originless' backend if no value
		// provided at the prompt.
		{
			name: "success with no setup configuration and defaulting to originless",
			args: args("compute deploy --token 123"),
			api: mock.API{
				CreateServiceFn:   createServiceOK,
				CreateDomainFn:    createDomainOK,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsNone,
			},
			stdin: []string{
				"Y", // when prompted to create a new service
				"",  // this stops prompting for backends
			},
			wantOutput: []string{
				"Backend (hostname or IP address, or leave blank to stop adding backends):",
				"SUCCESS: Deployed package (service 12345, version 1)",
			},
			dontWantOutput: []string{
				"Creating backend", // expect originless creation to be hidden
			},
		},
		// The following test validates that when dealing with an existing service,
		// if there are no backends, then we'll prompt the user for backends.
		{
			name: "success with no setup configuration and multiple backends prompted for existing service with no backends",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetServiceFn:      getServiceOK,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsNone,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			stdin: []string{
				"fastly.com",
				"443",
				"", // this is so we generate a backend name using a built-in formula
				"google.com",
				"123",
				"", // this is so we generate a backend name using a built-in formula
				"", // this stops prompting for backends
			},
			wantOutput: []string{
				"Backend (hostname or IP address, or leave blank to stop adding backends):",
				"Creating backend 'backend_1' (host: fastly.com, port: 443)...",
				"Creating backend 'backend_2' (host: google.com, port: 123)...",
				"SUCCESS: Deployed package (service 123, version 3)",
			},
		},
		// The following test is the same setup as above, but if the user provides
		// the --accept-defaults flag we won't prompt for any backends.
		{
			name: "success with no setup configuration and use of --accept-defaults for existing service",
			args: args("compute deploy --accept-defaults --service-id 123 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetServiceFn:      getServiceOK,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsNone,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			wantOutput: []string{
				"SUCCESS: Deployed package (service 123, version 3)",
			},
			dontWantOutput: []string{
				"Creating backend", // expect originless creation to be hidden
			},
		},
		// The following test validates that when dealing with an existing service,
		// if there are some backends that exist from the given [setup], then only
		// prompt the user for backends that don't exist.
		{
			name: "success with setup configuration and only prompting missing backends for existing service",
			args: args("compute deploy --service-id 123 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetServiceFn:      getServiceOK,
				ListDomainsFn:     listDomainsOk,
				ListBackendsFn:    listBackendsSome,
				CreateBackendFn:   createBackendOK,
				GetPackageFn:      getPackageOk,
				UpdatePackageFn:   updatePackageOk,
				ActivateVersionFn: activateVersionOk,
			},
			manifest: `
			name = "package"
			manifest_version = 1
			language = "rust"

			[setup]
				[[setup.backends]]
					name = "fastly"
					prompt = "Backend 1"
					address = "fastly.com"
					port = 443
				[[setup.backends]]
					name = "google"
					prompt = "Backend 2"
					address = "google.com"
					port = 443
				[[setup.backends]]
					name = "facebook"
					prompt = "Backend 3"
					address = "facebook.com"
					port = 443
			`,
			stdin: []string{
				"beep.com",
				"123",
				"boop.com",
				"456",
			},
			wantOutput: []string{
				"Creating backend 'google' (host: beep.com, port: 123)",
				"Creating backend 'facebook' (host: boop.com, port: 456)",
				"SUCCESS: Deployed package (service 123, version 3)",
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// Because the manifest can be mutated on each test scenario, we recreate
			// the file each time.
			manifestContent := `name = "package"`
			if testcase.manifest != "" {
				manifestContent = testcase.manifest
			}
			if err := os.WriteFile(filepath.Join(rootdir, manifest.Filename), []byte(manifestContent), 0777); err != nil {
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

			if testcase.reduceSizeLimit {
				compute.PackageSizeLimit = 1000000 // 1mb (our test package should above this)
			}

			if len(testcase.stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Stdin = stdin

				// Wait for user input and write it to the prompt
				inputc := make(chan string)
				go func() {
					for input := range inputc {
						fmt.Fprintln(prompt, input)
					}
				}()

				// We need a channel so we wait for `run()` to complete
				done := make(chan bool)

				// Call `app.Run()` and wait for response
				go func() {
					err = app.Run(opts)
					done <- true
				}()

				// User provides input
				//
				// NOTE: Must provide as much input as is expected to be waited on by `run()`.
				//       For example, if `run()` calls `input()` twice, then provide two messages.
				//       Otherwise the select statement will trigger the timeout error.
				for _, input := range testcase.stdin {
					inputc <- input
				}

				select {
				case <-done:
					// Wait for app.Run() to finish
				case <-time.After(time.Second):
					t.Fatalf("unexpected timeout waiting for mocked prompt inputs to be processed")
				}
			} else {
				stdin := ""
				if len(testcase.stdin) > 0 {
					stdin = testcase.stdin[0]
				}
				opts.Stdin = strings.NewReader(stdin)
				err = app.Run(opts)
			}

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

func listBackendsOk(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{
		{Name: "foo"},
		{Name: "bar"},
	}, nil
}

func listBackendsError(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return nil, testutil.Err
}

func listBackendsNone(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{}, nil
}

func listBackendsSome(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{
		{Address: "fastly.com", Name: "fastly", Port: 443},
	}, nil
}

func listDomainsNone(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{}, nil
}
