package custom_test

import (
	"bytes"
	"testing"
	"time"

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
			Args:      args("vcl custom describe --service-id 123 --version 3 --name foobar"),
			WantError: testutil.ErrAPI.Error(),
		},
		{
			Name: "validate GetVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("vcl custom describe --service-id 123 --version 3 --name foobar"),
			WantOutput: "Service ID: 123\nService Version: 3\nName: foobar\nMain: true\nContent: # some vcl content\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("vcl custom describe --service-id 123 --version 1 --name foobar"),
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
			Args:       args("vcl custom list --service-id 123 --version 1 --verbose"),
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
	//
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
