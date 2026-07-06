package newrelicotlp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"

	serviceRoot "github.com/fastly/cli/pkg/commands/service"
	loggingRoot "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/newrelicotlp"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestNewRelicOTLPCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--key abc --name foo --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate CreateNewRelicOTLP API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateNewRelicOTLPFn: func(_ context.Context, _ *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--key abc --name foo --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateNewRelicOTLP API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateNewRelicOTLPFn: func(_ context.Context, i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--key abc --name foo --service-id 123 --version 3",
			WantOutput: "Created New Relic OTLP logging endpoint 'foo' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateNewRelicOTLPFn: func(_ context.Context, i *fastly.CreateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --key abc --name foo --service-id 123 --version 1",
			WantOutput: "Created New Relic OTLP logging endpoint 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{serviceRoot.CommandName, loggingRoot.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestNewRelicOTLPDelete(t *testing.T) {
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
			Name: "validate DeleteNewRelic API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteNewRelicOTLPFn: func(_ context.Context, _ *fastly.DeleteNewRelicOTLPInput) error {
					return testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteNewRelic API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteNewRelicOTLPFn: func(_ context.Context, _ *fastly.DeleteNewRelicOTLPInput) error {
					return nil
				},
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "Deleted New Relic OTLP logging endpoint 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteNewRelicOTLPFn: func(_ context.Context, i *fastly.DeleteNewRelicOTLPInput) error {
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
				DeleteNewRelicOTLPFn: func(_ context.Context, i *fastly.DeleteNewRelicOTLPInput) error {
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
				DeleteNewRelicOTLPFn: func(_ context.Context, _ *fastly.DeleteNewRelicOTLPInput) error {
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Deleted New Relic OTLP logging endpoint 'foo' (service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteNewRelicOTLPFn: func(_ context.Context, i *fastly.DeleteNewRelicOTLPInput) error {
					// Verify operation happens on the cloned version (4), not original (2)
					if i.ServiceVersion != 4 {
						return fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 2",
			WantOutput: "Deleted New Relic OTLP logging endpoint 'foo' (service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteNewRelicOTLPFn: func(_ context.Context, i *fastly.DeleteNewRelicOTLPInput) error {
					// Verify operation happens on the cloned version (4), not original (3)
					if i.ServiceVersion != 4 {
						return fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 3",
			WantOutput: "Deleted New Relic OTLP logging endpoint 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{serviceRoot.CommandName, loggingRoot.CommandName, sub.CommandName, "delete"}, scenarios)
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
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetNewRelicOTLPFn: func(_ context.Context, _ *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetNewRelic API success",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				GetNewRelicOTLPFn: getNewRelic,
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nProcessing region: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 3\nToken: abc\nURL: \nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				GetNewRelicOTLPFn: getNewRelic,
			},
			Args:       "--name foobar --service-id 123 --version 1",
			WantOutput: "\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nFormat: \nFormat Version: 0\nName: foobar\nPlacement: \nProcessing region: \nRegion: \nResponse Condition: \nService ID: 123\nService Version: 1\nToken: abc\nURL: \nUpdated at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{serviceRoot.CommandName, loggingRoot.CommandName, sub.CommandName, "describe"}, scenarios)
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
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListNewRelicOTLPFn: func(_ context.Context, _ *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListNewRelics API success",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				ListNewRelicOTLPFn: listNewRelic,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "SERVICE ID  VERSION  NAME\n123         3        foo\n123         3        bar\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				ListNewRelicOTLPFn: listNewRelic,
			},
			Args:       "--service-id 123 --version 1",
			WantOutput: "SERVICE ID  VERSION  NAME\n123         1        foo\n123         1        bar\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				ListNewRelicOTLPFn: listNewRelic,
			},
			Args:       "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (auth: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\n\nToken: \n\nFormat: \n\nFormat Version: 0\n\nPlacement: \n\nRegion: \n\nProcessing region: \n\nResponse Condition: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\n\nToken: \n\nFormat: \n\nFormat Version: 0\n\nPlacement: \n\nRegion: \n\nProcessing region: \n\nResponse Condition: \n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{serviceRoot.CommandName, loggingRoot.CommandName, sub.CommandName, "list"}, scenarios)
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
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate UpdateNewRelic API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateNewRelicOTLPFn: func(_ context.Context, _ *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateNewRelic API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateNewRelicOTLPFn: func(_ context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated New Relic OTLP logging endpoint 'beepboop' (previously: foobar, service: 123, version: 3)",
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateNewRelicOTLPFn: func(_ context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, fmt.Errorf("Cannot update version %d. Versions that have been activated cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been activated cannot be updated",
		},
		{
			Name: "validate API error when modifying locked version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateNewRelicOTLPFn: func(_ context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return nil, fmt.Errorf("Cannot update version %d. Versions that have been locked cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been locked cannot be updated",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateNewRelicOTLPFn: func(_ context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					return &fastly.NewRelicOTLP{
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 1",
			WantOutput: "Updated New Relic OTLP logging endpoint 'beepboop' (previously: foobar, service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateNewRelicOTLPFn: func(_ context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					// Verify operation happens on the cloned version (4), not original (2)
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return &fastly.NewRelicOTLP{
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 2",
			WantOutput: "Updated New Relic OTLP logging endpoint 'beepboop' (previously: foobar, service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateNewRelicOTLPFn: func(_ context.Context, i *fastly.UpdateNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
					// Verify operation happens on the cloned version (4), not original (3)
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return &fastly.NewRelicOTLP{
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated New Relic OTLP logging endpoint 'beepboop' (previously: foobar, service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{serviceRoot.CommandName, loggingRoot.CommandName, sub.CommandName, "update"}, scenarios)
}

func getNewRelic(_ context.Context, i *fastly.GetNewRelicOTLPInput) (*fastly.NewRelicOTLP, error) {
	t := testutil.Date

	return &fastly.NewRelicOTLP{
		Name:           fastly.ToPointer(i.Name),
		Token:          fastly.ToPointer("abc"),
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listNewRelic(_ context.Context, i *fastly.ListNewRelicOTLPInput) ([]*fastly.NewRelicOTLP, error) {
	t := testutil.Date
	vs := []*fastly.NewRelicOTLP{
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
