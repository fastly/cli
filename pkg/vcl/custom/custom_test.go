package custom_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/vcl/custom"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestVCLCustomCreate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --content flag",
			Args:      args("vcl custom create --name foo --version 3"),
			WantError: "error parsing arguments: required flag --content not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom create --content /path/to/example.vcl --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom create --content /path/to/example.vcl --name foo"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom create --content /path/to/example.vcl --name foo --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom create --content ./testdata/example.vcl --name foo --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl custom create --content ./testdata/example.vcl --name foo --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate CreateVCL API success for non-main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --service-id 123 --version 3 --name foo --content ./testdata/example.vcl"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
		},
		{
			Name: "validate CreateVCL API success for main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --content ./testdata/example.vcl --main --name foo --service-id 123 --version 3"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 3, main: true)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 4, main: false)",
		},
		{
			Name: "validate CreateVCL API success with inline VCL content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = i.Content

					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --content inline_vcl --name foo --service-id 123 --version 3"),
			WantOutput: "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
			validateContent(testcase.WantError, testcase.Args, content, t)
		})
	}
}

func TestVCLCustomDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom delete --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom delete --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom delete --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom delete --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return testutil.ErrAPI
				},
			},
			Args:      args("vcl custom delete --name foobar --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate DeleteVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Args:       args("vcl custom delete --name foobar --service-id 123 --version 3"),
			WantOutput: "Deleted custom VCL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Args:       args("vcl custom delete --autoclone --name foo --service-id 123 --version 1"),
			WantOutput: "Deleted custom VCL 'foo' (service: 123, version: 4)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}

func TestVCLCustomDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom describe --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom describe --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom describe --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn: func(i *fastly.GetVCLInput) (*fastly.VCL, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl custom describe --name foobar --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate GetVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("vcl custom describe --name foobar --service-id 123 --version 3"),
			WantOutput: "Service ID: 123\nService Version: 3\nName: foobar\nMain: true\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("vcl custom describe --name foobar --service-id 123 --version 1"),
			WantOutput: "Service ID: 123\nService Version: 1\nName: foobar\nMain: true\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}

func TestVCLCustomList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListVCLs API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn: func(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl custom list --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate ListVCLs API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Args:       args("vcl custom list --service-id 123 --version 3"),
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         3        foo   true\n123         3        bar   false\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Args:       args("vcl custom list --service-id 123 --version 1"),
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         1        foo   true\n123         1        bar   false\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Args:       args("vcl custom list --service-id 123 --verbose --version 1"),
			WantOutput: "Fastly API token not provided\nFastly API endpoint: https://api.fastly.com\nService ID: 123\nService Version: 1\nName: foo\nMain: true\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nName: bar\nMain: false\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
		})
	}
}

func TestVCLCustomUpdate(t *testing.T) {
	var content string
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Args:      args("vcl custom update --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom update --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom update --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom update --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate UpdateVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.ErrAPI
				},
			},
			Args:      args("vcl custom update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate UpdateVCL API success with --new-name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        "# untouched",
						Main:           true,
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated custom VCL 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate UpdateVCL API success with --content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.VCL{
						Content:        *i.Content,
						Main:           true,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom update --content updated --name foobar --service-id 123 --version 3"),
			WantOutput: "Updated custom VCL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content

					return &fastly.VCL{
						Content:        *i.Content,
						Main:           true,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom update --autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1"),
			WantOutput: "Updated custom VCL 'foo' (service: 123, version: 4)",
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var buf bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.Args, testcase.API, &buf)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, buf.String(), testcase.WantOutput)
			validateContent(testcase.WantError, testcase.Args, content, t)
		})
	}
}

func getVCL(i *fastly.GetVCLInput) (*fastly.VCL, error) {
	t := time.Date(2021, time.June, 15, 23, 0, 0, 0, time.UTC)

	return &fastly.VCL{
		Content:        "# some vcl content",
		Main:           true,
		Name:           i.Name,
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		CreatedAt:      &t,
		UpdatedAt:      &t,
		DeletedAt:      &t,
	}, nil
}

func listVCLs(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
	t := time.Date(2021, time.June, 15, 23, 0, 0, 0, time.UTC)
	vs := []*fastly.VCL{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "foo",
			Main:           true,
			Content:        "# some vcl content",
			CreatedAt:      &t,
			UpdatedAt:      &t,
			DeletedAt:      &t,
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "bar",
			Main:           false,
			Content:        "# some vcl content",
			CreatedAt:      &t,
			UpdatedAt:      &t,
			DeletedAt:      &t,
		},
	}
	return vs, nil
}

// When dealing with successful test scenarios: validate the --content value was
// parsed as expected.
//
// Example: if --content is passed a file path, then we expect the
// testdata/example.vcl to have been read, otherwise we expect the given flag
// value to have been used as is.
func validateContent(wantError string, args []string, content string, t *testing.T) {
	if wantError == "" {
		for i, a := range args {
			if a == "--content" {
				want := args[i+1]
				if want == "./testdata/example.vcl" {
					want = custom.Content(want)
				}
				if content != want {
					t.Errorf("wanted %s, have %s", want, content)
				}
				break
			}
		}
	}
}
