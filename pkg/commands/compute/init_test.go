package compute_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

func TestInit(t *testing.T) {
	args := testutil.SplitArgs
	if os.Getenv("TEST_COMPUTE_INIT") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_INIT to run this test")
	}

	skRust := []config.StarterKit{
		{
			Name:   "Default",
			Path:   "https://github.com/fastly/compute-starter-kit-rust-default",
			Branch: "main",
		},
	}
	skJS := []config.StarterKit{
		{
			Name:   "Default",
			Path:   "https://github.com/fastly/compute-starter-kit-javascript-default",
			Branch: "main",
		},
	}

	scenarios := []struct {
		name             string
		args             []string
		configFile       config.File
		httpClientRes    []*http.Response
		httpClientErr    []error
		manifest         string
		wantFiles        []string
		unwantedFiles    []string
		wantError        string
		wantOutput       []string
		manifestIncludes string
		manifestPath     string
		stdin            string
		setupSteps       func() error
	}{
		{
			name:      "broken endpoint",
			args:      args("compute init --from https://example.com/i-dont-exist"),
			wantError: "failed to get package 'https://example.com/i-dont-exist': Not Found",
			httpClientRes: []*http.Response{
				{
					Body:       io.NopCloser(strings.NewReader("")),
					Status:     http.StatusText(http.StatusNotFound),
					StatusCode: http.StatusNotFound,
				},
			},
			httpClientErr: []error{
				nil,
			},
		},
		{
			name: "name prompt",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			stdin: "foobar", // expect the first prompt to be for the package name.
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `name = "foobar"`,
		},
		{
			name: "description prompt empty",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `description = ""`, // expect this to be empty
		},
		{
			name: "with author",
			args: args("compute init --author test@example.com"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `authors = ["test@example.com"]`,
		},
		{
			name: "with multiple authors",
			args: args("compute init --author test1@example.com --author test2@example.com"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `authors = ["test1@example.com", "test2@example.com"]`,
		},
		{
			name: "with --from set to starter kit repository",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to starter kit repository with .git extension and branch",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default.git --branch main"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to zip archive",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default/archive/refs/heads/main.zip"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to tar.gz archive",
			args: args("compute init --from https://github.com/Integralist/devnull/files/7339887/compute-starter-kit-rust-default-main.tar.gz"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: []config.StarterKit{
						{
							Name: "Default",
							Path: "https://github.com/fastly/compute-starter-kit-rust-default.git",
						},
					},
				},
			},
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with existing fastly.toml",
			args: args("compute init --auto-yes"), // --force will ignore a directory that isn't empty
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			manifest: `
			manifest_version = 2
			service_id = 1234
			name = "test"
			language = "rust"
			description = "test"
			authors = ["test@fastly.com"]`,
			wantOutput: []string{
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name: "no args and no user profiles means no email set for author field",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantFiles: []string{
				"Cargo.toml",
				"fastly.toml",
				"src/main.rs",
			},
			unwantedFiles: []string{
				"SECURITY.md",
			},
			wantOutput: []string{
				"Author (email):",
				"Language:",
				"Fetching package template",
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name: "no args but email defaults to config.toml value in author field",
			args: args("compute init"),
			configFile: config.File{
				Profiles: config.Profiles{
					"user": &config.Profile{
						Email:   "test@example.com",
						Default: true,
					},
					"non_default": &config.Profile{
						Email: "no-default@example.com",
					},
				},
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
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
				"Fetching package template",
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name: "non empty directory",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			wantError: "project directory not empty",
			manifest: `
			manifest_version = 2
			name = "test"`,
		},
		{
			name: "with default name inferred from directory",
			args: args("compute init"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name: "with directory name inferred from --directory",
			args: args("compute init --directory ./foo"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					Rust: skRust,
				},
			},
			stdin:            "Y",
			manifest:         `manifest_version = 2`,
			manifestPath:     "foo",
			manifestIncludes: `name = "foo`,
		},
		{
			name: "with JavaScript language",
			args: args("compute init --language javascript"),
			configFile: config.File{
				StarterKits: config.StarterKitLanguages{
					JavaScript: skJS,
				},
			},
			manifestIncludes: `name = "fastly-temp`,
		},
		// NOTE: This test verifies that we don't fetch a remote project.
		// Whether that be a starter kit or custom project template.
		// This is because "other" indicates an unsupported platform language.
		{
			name:             "with pre-compiled Wasm binary",
			args:             args("compute init --language other"),
			manifestIncludes: `language = "other"`,
			wantOutput: []string{
				"Initialized package",
				"To package a pre-compiled Wasm binary for deployment",
				"SUCCESS: Initialized package",
			},
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to an init environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			manifestPath := filepath.Join(testcase.manifestPath, manifest.Filename)

			// Create test environment
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Write: []testutil.FileIO{
					{Src: testcase.manifest, Dst: manifestPath},
				},
			})
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the init environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(pwd)
			}()

			// Before running the test, run some steps to initialize the environment.
			if testcase.setupSteps != nil {
				err := testcase.setupSteps()
				if err != nil {
					panic(err)
				}
			}

			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.Config = testcase.configFile

				if testcase.httpClientRes != nil || testcase.httpClientErr != nil {
					opts.HTTPClient = mock.HTMLClient(testcase.httpClientRes, testcase.httpClientErr)
				}

				// we need to define stdin as the init process prompts the user multiple
				// times, but we don't need to provide any values as all our prompts will
				// fallback to default values if the input is unrecognised.
				opts.Input = strings.NewReader(testcase.stdin)
				return opts, nil
			}
			err = app.Run(testcase.args, nil)

			t.Log(stdout.String())

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
				testutil.AssertStringContains(t, stdout.String(), s)
			}
			if testcase.manifestIncludes != "" {
				content, err := os.ReadFile(filepath.Join(rootdir, manifestPath))
				if err != nil {
					t.Fatal(err)
				}
				testutil.AssertStringContains(t, string(content), testcase.manifestIncludes)
			}
		})
	}
}

func TestInit_ExistingService(t *testing.T) {
	serviceID := fastly.NullString("LsyQ2UXDGk6d4ENjvgqTN4")
	customerID := fastly.NullString("YflD2HKQTx6q4RAwitdGA4")
	packageID := fastly.NullString("4AGdtiwAR4q6xTQKH2DlfY")

	scenarios := []struct {
		name              string
		args              []string
		getServiceDetails func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error)
		getPackage        func(*fastly.GetPackageInput) (*fastly.Package, error)
		expectInOutput    []string
		expectInManifest  []string
		expectNoManifest  bool
		expectInError     string
		suppresBeacon     bool
	}{
		{
			name: "when the service exists",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(gsi *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				if gsi.ServiceID != *serviceID {
					return nil, &fastly.HTTPError{
						StatusCode: http.StatusNotFound,
					}
				}
				return &fastly.ServiceDetail{
					ServiceID:  serviceID,
					CustomerID: customerID,
					Comment:    fastly.ToPointer(""),
					Name:       fastly.ToPointer("example service"),
					Type:       fastly.ToPointer("wasm"),
					ActiveVersion: &fastly.Version{
						Number: fastly.ToPointer(1),
					},
				}, nil
			},
			getPackage: func(gpi *fastly.GetPackageInput) (*fastly.Package, error) {
				if gpi.ServiceID != *serviceID || gpi.ServiceVersion != 1 {
					return nil, &fastly.HTTPError{
						StatusCode: http.StatusNotFound,
					}
				}
				return &fastly.Package{
					PackageID: packageID,
					ServiceID: serviceID,
					Metadata: &fastly.PackageMetadata{
						Authors:     []string{"author@example.com"},
						Description: fastly.NullString("a description"),
						Name:        fastly.NullString("test-package"),
						Language:    fastly.NullString("rust"),
					},
				}, nil
			},
			expectInOutput: []string{
				"Initializing Compute project from service LsyQ2UXDGk6d4ENjvgqTN4.",
				"SUCCESS: Initialized package test-package",
			},
			expectInManifest: []string{
				`name = "test-package"`,
				`authors = ["author@example.com"]`,
				`description = "a description"`,
				`language = "rust"`,
				`service_id = "LsyQ2UXDGk6d4ENjvgqTN4"`,
			},
		},
		{
			name: "when the service doesn't exist",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				return nil, &fastly.HTTPError{
					StatusCode: http.StatusNotFound,
				}
			},
			expectInOutput: []string{
				"Initializing Compute project from service LsyQ2UXDGk6d4ENjvgqTN4.",
			},
			expectInError:    "the service LsyQ2UXDGk6d4ENjvgqTN4 could not be found",
			expectNoManifest: true,
		},
		{
			name: "service has no versions that include package metadata",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ *fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				return &fastly.ServiceDetail{
					ServiceID:     serviceID,
					Name:          fastly.NullString("test-service"),
					Comment:       fastly.NullString(""),
					Type:          fastly.NullString("wasm"),
					ActiveVersion: nil,
					Versions: []*fastly.Version{
						{
							Active:   fastly.ToPointer(false),
							Deployed: fastly.ToPointer(false),
							Locked:   fastly.ToPointer(false),
							Number:   fastly.ToPointer(1),
						},
					},
				}, nil
			},
			getPackage: func(_ *fastly.GetPackageInput) (*fastly.Package, error) {
				return nil, &fastly.HTTPError{
					StatusCode: http.StatusNotFound,
				}
			},
			expectInError: "unable to find any version of service LsyQ2UXDGk6d4ENjvgqTN4 with an existing package",
		},
		{
			name: "service is vcl",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				return &fastly.ServiceDetail{
					ServiceID: serviceID,
					Type:      fastly.NullString("vcl"),
				}, nil
			},
			expectInError:    "service LsyQ2UXDGk6d4ENjvgqTN4 is not a Compute service",
			expectNoManifest: true,
		},
		{
			name:          "service id does not look like a Fastly ID",
			args:          testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4EN"),
			expectInError: "--from url seems invalid",
			// Not a valid URL OR Service ID
			suppresBeacon: true,
		},
		{
			name: "service has a cloned_from value",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				return &fastly.ServiceDetail{
					ServiceID: serviceID,
					Name:      fastly.NullString("cloned-service"),
					Comment:   fastly.NullString(""),
					Type:      fastly.NullString("wasm"),
					ActiveVersion: &fastly.Version{
						Number: fastly.ToPointer(1),
					},
				}, nil
			},
			getPackage: func(*fastly.GetPackageInput) (*fastly.Package, error) {
				return &fastly.Package{
					ServiceID: serviceID,
					PackageID: fastly.NullString("hVPTrHgswnF5KFwFKoQz1f"),
					Metadata: &fastly.PackageMetadata{
						ClonedFrom: fastly.ToPointer("https://github.com/fastly/compute-starter-kit-rust-empty"),
						Language:   fastly.ToPointer("rust"),
					},
				}, nil
			},
			expectInOutput: []string{"Initializing file structure from selected starter kit..."},
		},
		{
			name: "service has an unreachable cloned_from value",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				return &fastly.ServiceDetail{
					ServiceID: serviceID,
					Name:      fastly.NullString("cloned-service"),
					Comment:   fastly.NullString(""),
					Type:      fastly.NullString("wasm"),
					ActiveVersion: &fastly.Version{
						Number: fastly.ToPointer(1),
					},
				}, nil
			},
			getPackage: func(*fastly.GetPackageInput) (*fastly.Package, error) {
				return &fastly.Package{
					ServiceID: serviceID,
					PackageID: fastly.NullString("hVPTrHgswnF5KFwFKoQz1f"),
					Metadata: &fastly.PackageMetadata{
						ClonedFrom: fastly.ToPointer("https://github.com/fastly/fake-template"),
						Language:   fastly.ToPointer("rust"),
					},
				}, nil
			},
			expectInError: "could not fetch original source code",
		},
		{
			name: "service has active version greater than 1",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(*fastly.GetServiceInput) (*fastly.ServiceDetail, error) {
				return &fastly.ServiceDetail{
					ServiceID: serviceID,
					Name:      fastly.NullString("cloned-service"),
					Comment:   fastly.NullString(""),
					Type:      fastly.NullString("wasm"),
					ActiveVersion: &fastly.Version{
						Number: fastly.ToPointer(2),
					},
				}, nil
			},
			getPackage: func(*fastly.GetPackageInput) (*fastly.Package, error) {
				return &fastly.Package{
					ServiceID: serviceID,
					PackageID: fastly.NullString("hVPTrHgswnF5KFwFKoQz1f"),
					Metadata: &fastly.PackageMetadata{
						ClonedFrom: fastly.ToPointer("https://github.com/fastly/fake-template"),
						Language:   fastly.ToPointer("rust"),
					},
				}, nil
			},
			expectInOutput: []string{"not fetching starter kit source"},
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			// We're going to chdir to an init environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(pwd)
			}()

			// Create test environment
			rootdir := testutil.NewEnv(testutil.EnvOpts{
				T: t,
				Write: []testutil.FileIO{
					{Src: "", Dst: manifest.Filename},
				},
			})
			defer os.RemoveAll(rootdir)

			manifestPath := filepath.Join(rootdir, manifest.Filename)

			// Before running the test, chdir into the init environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably assert file structure.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}

			httpClient := &mock.HTTPClient{
				Responses: []*http.Response{
					// The body is closed by beacon.Notify.
					//nolint: bodyclose
					mock.NewHTTPResponse(http.StatusNoContent, nil, nil),
				},
				Errors: []error{
					nil,
				},
				Index:        -1,
				SaveRequests: true,
			}

			stdout := &threadsafe.Buffer{}
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, stdout)
				opts.APIClientFactory = mock.APIClient(mock.API{
					GetServiceDetailsFn: testcase.getServiceDetails,
					GetPackageFn:        testcase.getPackage,
				})
				opts.Input = strings.NewReader("")
				opts.HTTPClient = httpClient

				return opts, nil
			}

			err = app.Run(testcase.args, nil)

			if testcase.expectInError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err == nil {
					t.Log("expected an error and did not get one")
					t.Fail()
				}
				testutil.AssertErrorContains(t, err, testcase.expectInError)
			}

			t.Log(stdout.String())

			if testcase.suppresBeacon {
				testutil.AssertLength(t, 0, httpClient.Requests)
			} else {
				testutil.AssertLength(t, 1, httpClient.Requests)
				beaconReq := httpClient.Requests[0]
				testutil.AssertEqual(t, "fastly-notification-relay.edgecompute.app", beaconReq.URL.Hostname())
			}

			for _, s := range testcase.expectInOutput {
				testutil.AssertStringContains(t, stdout.String(), s)
			}

			if testcase.expectNoManifest {
				_, err = os.Stat(manifestPath)
				if err == nil {
					t.Log("found unexpected manifest file", manifestPath)
					t.Fail()
				}
			}

			if len(testcase.expectInManifest) > 0 {
				mfContentBytes, err := os.ReadFile(manifestPath)
				if err != nil {
					t.Fatal(err)
				}
				mfContent := string(mfContentBytes)
				for _, s := range testcase.expectInManifest {
					testutil.AssertStringContains(t, mfContent, s)
				}
			}
		})
	}
}
