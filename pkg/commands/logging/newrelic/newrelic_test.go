package newrelic_test

import (
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/logging"
	sub "github.com/fastly/cli/pkg/commands/logging/newrelic"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestNewRelicCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--key abc --name foo --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--key abc --name foo --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--key abc --name foo --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate CreateNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateNewRelicFn: func(_ *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--key abc --name foo --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateNewRelicFn: func(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--key abc --name foo --service-id 123 --version 3",
			WantOutput: "Created New Relic logging endpoint 'foo' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateNewRelicFn: func(i *fastly.CreateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --key abc --name foo --service-id 123 --version 1",
			WantOutput: "Created New Relic logging endpoint 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestNewRelicDelete(t *testing.T) {
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
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--name foobar --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--name foobar --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate DeleteNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteNewRelicFn: func(_ *fastly.DeleteNewRelicInput) error {
					return testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteNewRelicFn: func(_ *fastly.DeleteNewRelicInput) error {
					return nil
				},
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "Deleted New Relic logging endpoint 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteNewRelicFn: func(_ *fastly.DeleteNewRelicInput) error {
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Deleted New Relic logging endpoint 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestNewRelicDescribe(t *testing.T) {
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
			Name: "validate GetNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicFn: func(_ *fastly.GetNewRelicInput) (*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicFn:  getNewRelic,
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nProcessing region: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 3\nToken: abc\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetNewRelicFn:  getNewRelic,
			},
			Args:       "--name foobar --service-id 123 --version 1",
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nProcessing region: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 1\nToken: abc\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestNewRelicList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
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
			Name: "validate ListNewRelics API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: func(_ *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListNewRelics API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: listNewRelic,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "SERVICE ID  VERSION  NAME\n123         3        foo\n123         3        bar\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: listNewRelic,
			},
			Args:       "--service-id 123 --version 1",
			WantOutput: "SERVICE ID  VERSION  NAME\n123         1        foo\n123         1        bar\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListNewRelicFn: listNewRelic,
			},
			Args:       "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\n\nToken: \n\nFormat: \n\nFormat Version: 0\n\nPlacement: \n\nRegion: \n\nProcessing region: \n\nResponse Condition: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\n\nToken: \n\nFormat: \n\nFormat Version: 0\n\nPlacement: \n\nRegion: \n\nProcessing region: \n\nResponse Condition: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestNewRelicUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--service-id 123 --version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar --service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--name foobar --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      "--name foobar --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate UpdateNewRelic API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateNewRelicFn: func(_ *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateNewRelic API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateNewRelicFn: func(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated New Relic logging endpoint 'beepboop' (previously: foobar, service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateNewRelicFn: func(i *fastly.UpdateNewRelicInput) (*fastly.NewRelic, error) {
					return &fastly.NewRelic{
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 1",
			WantOutput: "Updated New Relic logging endpoint 'beepboop' (previously: foobar, service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func getNewRelic(i *fastly.GetNewRelicInput) (*fastly.NewRelic, error) {
	t := testutil.Date

	return &fastly.NewRelic{
		Name:           fastly.ToPointer(i.Name),
		Token:          fastly.ToPointer("abc"),
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listNewRelic(i *fastly.ListNewRelicInput) ([]*fastly.NewRelic, error) {
	t := testutil.Date
	vs := []*fastly.NewRelic{
		{
			Name:           fastly.ToPointer("foo"),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
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
