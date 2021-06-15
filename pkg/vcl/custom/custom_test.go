package custom_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestVCLCustomCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --file flag",
			Args:      args("vcl custom create --version 3"),
			WantError: "error parsing arguments: required flag --file not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("vcl custom create --file /path/to/main.vcl"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("vcl custom create --version 3 --file /path/to/example.vcl"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("vcl custom create --service-id 123 --version 1 --file /path/to/example.vcl"),
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
			Args:      args("vcl custom create --service-id 123 --version 3 --file /path/to/example.vcl"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate CreateVCL API success for non-main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --service-id 123 --version 3 --file /path/to/example.vcl"),
			WantOutput: "Created custom VCL 'example' (service: 123, version: 3, main: false)",
		},
		{
			Name: "validate CreateVCL API success for main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --service-id 123 --version 3 --file /path/to/example.vcl --main"),
			WantOutput: "Created custom VCL 'example' (service: 123, version: 3, main: true)",
		},
		{
			Name: "validate CreateVCL API success for main VCL with custom name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("vcl custom create --service-id 123 --version 3 --file /path/to/example.vcl --main --name foobar"),
			WantOutput: "Created custom VCL 'foobar' (service: 123, version: 3, main: true)",
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
			Args:      args("vcl custom delete --service-id 123 --version 1 --name foobar"),
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
			Args:      args("vcl custom delete --service-id 123 --version 3 --name foobar"),
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
			Args:       args("vcl custom delete --service-id 123 --version 3 --name foobar"),
			WantOutput: "Deleted custom VCL 'foobar' (service: 123, version: 3)",
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
	//
}

func TestVCLCustomList(t *testing.T) {
	//
}

func TestVCLCustomUpdate(t *testing.T) {
	//
}
