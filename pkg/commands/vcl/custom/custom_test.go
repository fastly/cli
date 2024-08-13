package custom_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/vcl"
	sub "github.com/fastly/cli/pkg/commands/vcl/custom"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVCLCustomCreate(t *testing.T) {
	var content string
	scenarios := []testutil.TestScenario{
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--content ./testdata/example.vcl --name foo --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--content ./testdata/example.vcl --name foo --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate CreateVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--content ./testdata/example.vcl --name foo --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateVCL API success for non-main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:             "--content ./testdata/example.vcl --name foo --service-id 123 --version 3",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate CreateVCL API success for main VCL",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:             "--content ./testdata/example.vcl --main --name foo --service-id 123 --version 3",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 3, main: true)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:             "--autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 4, main: false)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate CreateVCL API success with inline VCL content",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateVCLFn: func(i *fastly.CreateVCLInput) (*fastly.VCL, error) {
					// Track the contents parsed
					content = *i.Content
					if i.Content == nil {
						i.Content = fastly.ToPointer("")
					}
					if i.Main == nil {
						b := false
						i.Main = &b
					}
					if i.Name == nil {
						i.Name = fastly.ToPointer("")
					}
					return &fastly.VCL{
						Content:        i.Content,
						Main:           i.Main,
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:             "--content inline_vcl --name foo --service-id 123 --version 3",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestVCLCustomDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foobar --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foobar --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate DeleteVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return testutil.Err
				},
			},
			Arg:       "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteVCLFn: func(i *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Arg:        "--name foobar --service-id 123 --version 3",
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
			Arg:        "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Deleted custom VCL 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestVCLCustomDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn: func(i *fastly.GetVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetVCL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Arg:        "--name foobar --service-id 123 --version 3",
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Arg:        "--name foobar --service-id 123 --version 1",
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestVCLCustomList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListVCLs API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn: func(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListVCLs API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Arg:        "--service-id 123 --version 3",
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         3        foo   true\n123         3        bar   false\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Arg:        "--service-id 123 --version 1",
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         1        foo   true\n123         1        bar   false\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListVCLsFn:     listVCLs,
			},
			Arg:        "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nMain: false\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestVCLCustomUpdate(t *testing.T) {
	var content string
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foobar --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foobar --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate UpdateVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateVCL API success with --new-name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateVCLFn: func(i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        fastly.ToPointer("# untouched"),
						Main:           fastly.ToPointer(true),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:        "--name foobar --new-name beepboop --service-id 123 --version 3",
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
						Content:        i.Content,
						Main:           fastly.ToPointer(true),
						Name:           fastly.ToPointer(i.Name),
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:             "--content updated --name foobar --service-id 123 --version 3",
			WantOutput:      "Updated custom VCL 'foobar' (service: 123, version: 3)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
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
						Content:        i.Content,
						Main:           fastly.ToPointer(true),
						Name:           fastly.ToPointer(i.Name),
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:             "--autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1",
			WantOutput:      "Updated custom VCL 'foo' (service: 123, version: 4)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func getVCL(i *fastly.GetVCLInput) (*fastly.VCL, error) {
	t := testutil.Date

	return &fastly.VCL{
		Content:        fastly.ToPointer("# some vcl content"),
		Main:           fastly.ToPointer(true),
		Name:           fastly.ToPointer(i.Name),
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listVCLs(i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
	t := testutil.Date
	vs := []*fastly.VCL{
		{
			Content:        fastly.ToPointer("# some vcl content"),
			Main:           fastly.ToPointer(true),
			Name:           fastly.ToPointer("foo"),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			Content:        fastly.ToPointer("# some vcl content"),
			Main:           fastly.ToPointer(false),
			Name:           fastly.ToPointer("bar"),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
