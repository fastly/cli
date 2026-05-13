package custom_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"

	top "github.com/fastly/cli/pkg/commands/service"
	root "github.com/fastly/cli/pkg/commands/service/vcl"
	sub "github.com/fastly/cli/pkg/commands/service/vcl/custom"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVCLCustomCreate(t *testing.T) {
	var content string
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate CreateVCL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateVCLFn: func(_ context.Context, _ *fastly.CreateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--content ./testdata/example.vcl --name foo --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateVCL API success for non-main VCL",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateVCLFn: func(_ context.Context, i *fastly.CreateVCLInput) (*fastly.VCL, error) {
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
			Args:            "--content ./testdata/example.vcl --name foo --service-id 123 --version 3",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate CreateVCL API success for main VCL",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateVCLFn: func(_ context.Context, i *fastly.CreateVCLInput) (*fastly.VCL, error) {
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
			Args:            "--content ./testdata/example.vcl --main --name foo --service-id 123 --version 3",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 3, main: true)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateVCLFn: func(_ context.Context, i *fastly.CreateVCLInput) (*fastly.VCL, error) {
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
			Args:            "--autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 4, main: false)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate CreateVCL API success with inline VCL content",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateVCLFn: func(_ context.Context, i *fastly.CreateVCLInput) (*fastly.VCL, error) {
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
			Args:            "--content inline_vcl --name foo --service-id 123 --version 3",
			WantOutput:      "Created custom VCL 'foo' (service: 123, version: 3, main: false)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
	}

	testutil.RunCLIScenarios(t, []string{top.CommandName, root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestVCLCustomDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate DeleteVCL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteVCLFn: func(_ context.Context, _ *fastly.DeleteVCLInput) error {
					return testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteVCL API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteVCLFn: func(_ context.Context, _ *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "Deleted custom VCL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteVCLFn: func(_ context.Context, i *fastly.DeleteVCLInput) error {
					return fmt.Errorf("Cannot update version %d. Versions that have been activated cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been activated cannot be updated",
		},
		{
			Name: "validate API error when modifying locked version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteVCLFn: func(_ context.Context, i *fastly.DeleteVCLInput) error {
					return fmt.Errorf("Cannot update version %d. Versions that have been locked cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been locked cannot be updated",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteVCLFn: func(_ context.Context, _ *fastly.DeleteVCLInput) error {
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Deleted custom VCL 'foo' (service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteVCLFn: func(_ context.Context, i *fastly.DeleteVCLInput) error {
					// Verify operation happens on the cloned version (4), not original (2)
					if i.ServiceVersion != 4 {
						return fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 2",
			WantOutput: "Deleted custom VCL 'foo' (service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteVCLFn: func(_ context.Context, i *fastly.DeleteVCLInput) error {
					// Verify operation happens on the cloned version (4), not original (3)
					if i.ServiceVersion != 4 {
						return fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 3",
			WantOutput: "Deleted custom VCL 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{top.CommandName, root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestVCLCustomDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetVCL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetVCLFn: func(_ context.Context, _ *fastly.GetVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetVCL API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetVCLFn:     getVCL,
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetVCLFn:     getVCL,
			},
			Args:       "--name foobar --service-id 123 --version 1",
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{top.CommandName, root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestVCLCustomList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListVCLs API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListVCLsFn: func(_ context.Context, _ *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListVCLs API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListVCLsFn:   listVCLs,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         3        foo   true\n123         3        bar   false\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListVCLsFn:   listVCLs,
			},
			Args:       "--service-id 123 --version 1",
			WantOutput: "SERVICE ID  VERSION  NAME  MAIN\n123         1        foo   true\n123         1        bar   false\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListVCLsFn:   listVCLs,
			},
			Args:       "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (auth: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nMain: true\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nMain: false\nContent: \n# some vcl content\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{top.CommandName, root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestVCLCustomUpdate(t *testing.T) {
	var content string
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate UpdateVCL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateVCLFn: func(_ context.Context, _ *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateVCL API success with --new-name",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return &fastly.VCL{
						Content:        fastly.ToPointer("# untouched"),
						Main:           fastly.ToPointer(true),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated custom VCL 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate UpdateVCL API success with --content",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
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
			Args:            "--content updated --name foobar --service-id 123 --version 3",
			WantOutput:      "Updated custom VCL 'foobar' (service: 123, version: 3)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return nil, fmt.Errorf("Cannot update version %d. Versions that have been activated cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--content updated --name foobar --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been activated cannot be updated",
		},
		{
			Name: "validate API error when modifying locked version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					return nil, fmt.Errorf("Cannot update version %d. Versions that have been locked cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--content updated --name foobar --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been locked cannot be updated",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
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
			Args:            "--autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 1",
			WantOutput:      "Updated custom VCL 'foo' (service: 123, version: 4)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					// Verify operation happens on the cloned version (4), not original (2)
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
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
			Args:            "--autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 2",
			WantOutput:      "Updated custom VCL 'foo' (service: 123, version: 4)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateVCLFn: func(_ context.Context, i *fastly.UpdateVCLInput) (*fastly.VCL, error) {
					// Verify operation happens on the cloned version (4), not original (3)
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
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
			Args:            "--autoclone --content ./testdata/example.vcl --name foo --service-id 123 --version 3",
			WantOutput:      "Updated custom VCL 'foo' (service: 123, version: 4)",
			PathContentFlag: &testutil.PathContentFlag{Flag: "content", Fixture: "example.vcl", Content: func() string { return content }},
		},
	}

	testutil.RunCLIScenarios(t, []string{top.CommandName, root.CommandName, sub.CommandName, "update"}, scenarios)
}

func getVCL(_ context.Context, i *fastly.GetVCLInput) (*fastly.VCL, error) {
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

func listVCLs(_ context.Context, i *fastly.ListVCLsInput) ([]*fastly.VCL, error) {
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
