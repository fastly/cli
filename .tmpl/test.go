package ${CLI_PACKAGE}_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

const (
	baseCommand = "${CLI_COMMAND}"
)

func TestCreate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--service-id 123 --version 1",
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate Create${CLI_API} API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Create${CLI_API}Fn: func(i *fastly.Create${CLI_API}Input) (*fastly.${CLI_API}, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Create${CLI_API} API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Create${CLI_API}Fn: func(i *fastly.Create${CLI_API}Input) (*fastly.${CLI_API}, error) {
					return &fastly.${CLI_API}{
						ServiceID: i.ServiceID,
					}, nil
				},
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "Created <...> '456' (service: 123)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				Create${CLI_API}Fn: func(i *fastly.Create${CLI_API}Input) (*fastly.${CLI_API}, error) {
					return &fastly.VCL{
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       "--autoclone --service-id 123 --version 1",
			WantOutput: "Created <...> 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand, "create"}, scenarios)
}

func TestDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 1",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--service-id 123 --version 1",
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate Delete${CLI_API} API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Delete${CLI_API}Fn: func(i *fastly.Delete${CLI_API}Input) error {
					return testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Delete${CLI_API} API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Delete${CLI_API}Fn: func(i *fastly.Delete${CLI_API}Input) error {
					return nil
				},
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "Deleted <...> '456' (service: 123)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				Delete${CLI_API}Fn: func(i *fastly.Delete${CLI_API}Input) error {
					return nil
				},
			},
			Args:       "--autoclone --service-id 123 --version 1",
			WantOutput: "Deleted <...> 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand, "delete"}, scenarios)
}

func TestDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate Get${CLI_API} API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Get${CLI_API}Fn: func(i *fastly.Get${CLI_API}Input) (*fastly.${CLI_API}, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Get${CLI_API} API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Get${CLI_API}Fn: get${CLI_API},
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "<...>",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand, "describe"}, scenarios)
}

func TestList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate List${CLI_API}s API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				List${CLI_API}sFn: func(i *fastly.List${CLI_API}sInput) ([]*fastly.${CLI_API}, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate List${CLI_API}s API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				List${CLI_API}sFn: list${CLI_API}s,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "<...>",
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				List${CLI_API}sFn: list${CLI_API}s,
			},
			Args:       "--service-id 123 --version 3 --verbose",
			WantOutput: "<...>",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand, "list"}, scenarios)
}

func TestUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
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
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--service-id 123 --version 1",
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate Update${CLI_API} API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Update${CLI_API}Fn: func(i *fastly.Update${CLI_API}Input) (*fastly.${CLI_API}, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate Update${CLI_API} API success with --new-name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				Update${CLI_API}Fn: func(i *fastly.Update${CLI_API}Input) (*fastly.${CLI_API}, error) {
					return &fastly.${CLI_API}{
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated <...> 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
	}

	testutil.RunScenarios(t, []string{baseCommand, "update"}, scenarios)
}

func get${CLI_API}(i *fastly.Get${CLI_API}Input) (*fastly.${CLI_API}, error) {
	t := testutil.Date

	return &fastly.${CLI_API}{
		ServiceID: i.ServiceID,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func list${CLI_API}s(i *fastly.List${CLI_API}sInput) ([]*fastly.${CLI_API}, error) {
	t := testutil.Date
	vs := []*fastly.${CLI_API}{
		{
			ServiceID: i.ServiceID,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			ServiceID: i.ServiceID,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
