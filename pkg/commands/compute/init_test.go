package compute_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/starterkit"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

// buildTestTarballGz returns the bytes of a minimal tar.gz archive
// containing a single file, for use as a fake starter-kit tarball response.
func buildTestTarballGz(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	content := []byte("name = \"empty\"\nlanguage = \"rust\"\nmanifest_version = 3\n")
	if err := tw.WriteHeader(&tar.Header{
		Name: "fastly.toml",
		Mode: 0o644,
		Size: int64(len(content)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestInit(t *testing.T) {
	args := testutil.SplitArgs
	if os.Getenv("TEST_COMPUTE_INIT") == "" {
		t.Log("skipping test")
		t.Skip("Set TEST_COMPUTE_INIT to run this test")
	}

	// NOTE: Starter kits are no longer sourced from a local config.File
	// fixture -- the interactive prompt now fetches them live from the
	// starter-kit edge service (see pkg/starterkit), so scenarios that rely
	// on the default (option "1") kit selection exercise the real service.

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
			name:  "name prompt",
			args:  args("compute init"),
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
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `description = ""`, // expect this to be empty
		},
		{
			name: "with author",
			args: args("compute init --author test@example.com"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `authors = ["test@example.com"]`,
		},
		{
			name: "with multiple authors",
			args: args("compute init --author test1@example.com --author test2@example.com"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
			},
			manifestIncludes: `authors = ["test1@example.com", "test2@example.com"]`,
		},
		{
			name: "with --from set to starter kit repository",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to starter kit repository when dir with same name exists in pwd",
			args: args("compute init --auto-yes --from https://github.com/fastly/compute-starter-kit-rust-default"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
			setupSteps: func() error {
				return os.MkdirAll("compute-starter-kit-rust-default", 0o755)
			},
		},
		{
			name: "with --from set to starter kit repository with .git extension and branch",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default.git --branch main"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to starter kit repository with .git extension and branch when dir with same name exists in pwd",
			args: args("compute init --auto-yes --from https://github.com/fastly/compute-starter-kit-rust-default.git --branch main"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
			setupSteps: func() error {
				return os.MkdirAll("compute-starter-kit-rust-default.git", 0o755)
			},
		},
		{
			name: "with --from set to zip archive",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-rust-default/archive/refs/heads/main.zip"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to zip archive when file with same name exists in pwd",
			args: args("compute init --auto-yes --from https://github.com/fastly/compute-starter-kit-rust-default/archive/refs/heads/main.zip"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
			setupSteps: func() error {
				file, err := os.Create("main.zip")
				if file != nil {
					defer file.Close()
				}
				return err
			},
		},
		{
			name: "with --from set to tar.gz archive",
			args: args("compute init --from https://github.com/Integralist/devnull/files/7339887/compute-starter-kit-rust-default-main.tar.gz"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to starter-kit/<lang>/<name> reference",
			args: args("compute init --from starter-kit/javascript/typescript-default"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with --from set to a legacy repo carrying a .starter-kit-id redirect",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-typescript-default"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
		},
		{
			name: "with existing fastly.toml",
			args: args("compute init --auto-yes"), // --force will ignore a directory that isn't empty
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
			wantFiles: []string{
				"Cargo.toml",
				"fastly.toml",
				"src/main.rs",
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
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Email: "test@example.com",
						},
						"non_default": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Email: "no-default@example.com",
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
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"Saving manifest changes",
				"Initializing package",
			},
		},
		{
			name:      "non empty directory",
			args:      args("compute init"),
			wantError: "project directory not empty",
			manifest: `
			manifest_version = 2
			name = "test"`,
		},
		{
			name:             "with default name inferred from directory",
			args:             args("compute init"),
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name:             "with directory name inferred from --directory",
			args:             args("compute init --directory ./foo"),
			stdin:            "Y",
			manifest:         `manifest_version = 2`,
			manifestPath:     "foo",
			manifestIncludes: `name = "foo`,
		},
		{
			name:             "with JavaScript language",
			args:             args("compute init --language javascript"),
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name:             "with C++ language",
			args:             args("compute init --language cpp"),
			manifestIncludes: `name = "fastly-temp`,
		},
		{
			name: "with --from set to C++ empty starter kit",
			args: args("compute init --from https://github.com/fastly/compute-starter-kit-cpp-empty"),
			wantOutput: []string{
				"Fetching package template",
				"Reading fastly.toml",
				"SUCCESS: Initialized package",
			},
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
				if err := testcase.setupSteps(); err != nil {
					t.Fatal(err)
				}
			}

			var stdout threadsafe.Buffer
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
		getServiceDetails func(context.Context, *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error)
		getPackage        func(context.Context, *fastly.GetPackageInput) (*fastly.Package, error)
		expectInOutput    []string
		expectInManifest  []string
		expectNoManifest  bool
		expectInError     string
		suppressBeacon    bool
		// starterKitIDCheck is true when ClonedFrom is a fastly-org GitHub URL,
		// which now triggers an extra HTTP call (the .starter-kit-id redirect
		// lookup) before falling back to git-clone, in addition to the beacon
		// notification call.
		starterKitIDCheck bool
		// tarballFromMonorepo is true when ClonedFrom is a compute-starter-kits
		// monorepo URL (see starterkit.ParseSourceURL), which fetches a kit
		// tarball directly instead of falling back to git-clone, in addition
		// to the beacon notification call.
		tarballFromMonorepo bool
	}{
		{
			name: "when the service exists",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ context.Context, gsi *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getPackage: func(_ context.Context, gpi *fastly.GetPackageInput) (*fastly.Package, error) {
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
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getPackage: func(_ context.Context, _ *fastly.GetPackageInput) (*fastly.Package, error) {
				return nil, &fastly.HTTPError{
					StatusCode: http.StatusNotFound,
				}
			},
			expectInError: "unable to find any version of service LsyQ2UXDGk6d4ENjvgqTN4 with an existing package",
		},
		{
			name: "service is vcl",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			suppressBeacon: true,
		},
		{
			name: "service has a cloned_from value",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getPackage: func(_ context.Context, _ *fastly.GetPackageInput) (*fastly.Package, error) {
				return &fastly.Package{
					ServiceID: serviceID,
					PackageID: fastly.NullString("hVPTrHgswnF5KFwFKoQz1f"),
					Metadata: &fastly.PackageMetadata{
						ClonedFrom: fastly.ToPointer("https://github.com/fastly/compute-starter-kit-rust-empty"),
						Language:   fastly.ToPointer("rust"),
					},
				}, nil
			},
			expectInOutput:    []string{"Initializing file structure from selected starter kit..."},
			starterKitIDCheck: true,
		},
		{
			name: "service has a cloned_from value pointing into the compute-starter-kits monorepo",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getPackage: func(_ context.Context, _ *fastly.GetPackageInput) (*fastly.Package, error) {
				return &fastly.Package{
					ServiceID: serviceID,
					PackageID: fastly.NullString("hVPTrHgswnF5KFwFKoQz1f"),
					Metadata: &fastly.PackageMetadata{
						ClonedFrom: fastly.ToPointer("https://github.com/fastly/compute-starter-kits/tree/main/starter-kits/rust/empty"),
						Language:   fastly.ToPointer("rust"),
					},
				}, nil
			},
			expectInOutput:      []string{"Initializing file structure from selected starter kit..."},
			tarballFromMonorepo: true,
		},
		{
			name: "service has an unreachable cloned_from value",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getPackage: func(_ context.Context, _ *fastly.GetPackageInput) (*fastly.Package, error) {
				return &fastly.Package{
					ServiceID: serviceID,
					PackageID: fastly.NullString("hVPTrHgswnF5KFwFKoQz1f"),
					Metadata: &fastly.PackageMetadata{
						ClonedFrom: fastly.ToPointer("https://github.com/fastly/fake-template"),
						Language:   fastly.ToPointer("rust"),
					},
				}, nil
			},
			expectInError:     "could not fetch original source code",
			starterKitIDCheck: true,
		},
		{
			name: "service has active version greater than 1",
			args: testutil.SplitArgs("compute init --from LsyQ2UXDGk6d4ENjvgqTN4"),
			getServiceDetails: func(_ context.Context, _ *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
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
			getPackage: func(_ context.Context, _ *fastly.GetPackageInput) (*fastly.Package, error) {
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

			// The body is closed by beacon.Notify.
			//nolint: bodyclose
			responses := []*http.Response{mock.NewHTTPResponse(http.StatusNoContent, nil, nil)}
			errs := []error{nil}
			if testcase.starterKitIDCheck {
				// ClonePackageFromEndpoint now checks for a .starter-kit-id
				// redirect marker before cloning a fastly-org repo; simulate
				// "not found" so it falls through to the existing git-clone
				// behavior these scenarios expect.
				//nolint: bodyclose
				responses = append([]*http.Response{mock.NewHTTPResponse(http.StatusNotFound, nil, io.NopCloser(strings.NewReader("")))}, responses...)
				errs = append([]error{nil}, errs...)
			}
			if testcase.tarballFromMonorepo {
				// ClonePackageFromEndpoint recognizes the cloned_from URL as a
				// compute-starter-kits monorepo link and fetches the kit
				// tarball directly instead of git-cloning.
				tarball := buildTestTarballGz(t)
				//nolint: bodyclose
				responses = append([]*http.Response{mock.NewHTTPResponse(http.StatusOK, map[string]string{"Content-Type": "application/gzip"}, io.NopCloser(bytes.NewReader(tarball)))}, responses...)
				errs = append([]error{nil}, errs...)
			}

			httpClient := &mock.HTTPClient{
				Responses:    responses,
				Errors:       errs,
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

			if testcase.suppressBeacon {
				testutil.AssertLength(t, 0, httpClient.Requests)
			} else {
				wantRequests := 1
				if testcase.starterKitIDCheck || testcase.tarballFromMonorepo {
					wantRequests = 2
				}
				testutil.AssertLength(t, wantRequests, httpClient.Requests)
				beaconReq := httpClient.Requests[len(httpClient.Requests)-1]
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

// TestPromptForStarterKitBounds verifies that bounds checks are applied to
// starter kit selection in the interactive mode prompt.
func TestPromptForStarterKitBounds(t *testing.T) {
	kits := []starterkit.Kit{
		{
			ID:       "rust-default",
			Name:     "Default",
			Language: "rust",
		},
		{
			ID:       "rust-empty",
			Name:     "Empty",
			Language: "rust",
		},
	}

	scenarios := []struct {
		name string
		// stdin is the input given at the starter kit prompt. An invalid entry
		// is rejected and the prompt repeats, so those cases supply a valid
		// follow-up value.
		stdin    string
		wantFrom string
		// wantRejected asserts the validation message was shown to the user.
		wantRejected bool
	}{
		{
			name:     "first option",
			stdin:    "1\n",
			wantFrom: kits[0].FromValue(),
		},
		{
			name:     "last option",
			stdin:    "2\n",
			wantFrom: kits[1].FromValue(),
		},
		{
			name:     "no input defaults to the first option",
			stdin:    "\n",
			wantFrom: kits[0].FromValue(),
		},
		{
			name:     "git URL is passed through",
			stdin:    "https://github.com/fastly/compute-starter-kit-rust-websockets\n",
			wantFrom: "https://github.com/fastly/compute-starter-kit-rust-websockets",
		},
		{
			name:     "starter kit reference is passed through",
			stdin:    "starter-kit/rust/websockets\n",
			wantFrom: "starter-kit/rust/websockets",
		},
		{
			// Without the lower bound this indexed kits[-1] and panicked.
			name:         "zero is rejected",
			stdin:        "0\n1\n",
			wantFrom:     kits[0].FromValue(),
			wantRejected: true,
		},
		{
			// Without the lower bound this indexed kits[-2] and panicked.
			name:         "negative is rejected",
			stdin:        "-1\n2\n",
			wantFrom:     kits[1].FromValue(),
			wantRejected: true,
		},
		{
			name:         "above the upper bound is rejected",
			stdin:        "3\n1\n",
			wantFrom:     kits[0].FromValue(),
			wantRejected: true,
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			var stdout threadsafe.Buffer
			c := compute.InitCommand{
				Base: argparser.Base{
					Globals: testutil.MockGlobalData(testutil.SplitArgs("compute init"), &stdout),
				},
			}

			// The starter-kit edge service has no concept of git refs, so the
			// branch/tag returned are always empty.
			from, branch, tag, err := c.PromptForStarterKit(kits, strings.NewReader(testcase.stdin), &stdout)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			testutil.AssertEqual(t, testcase.wantFrom, from)
			testutil.AssertEqual(t, "", branch)
			testutil.AssertEqual(t, "", tag)

			if testcase.wantRejected {
				testutil.AssertStringContains(t, stdout.String(), "must be a valid option, git URL, or starter-kit/<lang>/<name> reference")
			}
		})
	}
}

// TestPromptForStarterKitBoundsNonInteractive verifies that bounds checks are
// applied to starter kit selection when either the AcceptDefaults flag or the
// NonInteractive flag are true, which skips the prompt and prompt validation.
func TestPromptForStarterKitBoundsNonInteractive(t *testing.T) {
	var stdout threadsafe.Buffer
	g := testutil.MockGlobalData(testutil.SplitArgs("compute init"), &stdout)
	g.Flags.AcceptDefaults = true

	c := compute.InitCommand{Base: argparser.Base{Globals: g}}

	// With defaults accepted and no kits available, the option would otherwise
	// fall back to "1" with an empty slice, which is out of range.
	_, _, _, err := c.PromptForStarterKit([]starterkit.Kit{}, strings.NewReader(""), &stdout)
	testutil.AssertErrorContains(t, err, "no default starter kits configured for this language")
}
