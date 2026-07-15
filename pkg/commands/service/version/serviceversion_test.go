package version_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v16/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/version"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersionClone(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 1",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name: "validate successful clone",
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantOutput: "Cloned service 123 version 1 to version 4",
		},
		{
			Name: "validate successful clone json output",
			Args: "--service-id 123 --version 1 --json",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantOutput: cloneServiceVersionJSONOutput,
		},
		{
			Name: "validate error will be passed through if cloning fails",
			Args: "--service-id 456 --version 1",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "clone"}, scenarios)
}

func TestVersionList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:       "--service-id 123",
			API:        &mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsShortOutput,
		},
		{
			Args:       "--service-id 123 --verbose",
			API:        &mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:       "--service-id 123 -v",
			API:        &mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:       "--verbose --service-id 123",
			API:        &mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:       "-v --service-id 123",
			API:        &mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:      "--service-id 123",
			API:       &mock.API{ListVersionsFn: testutil.ListVersionsError},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestVersionUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --comment foo --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionOK,
			},
			WantOutput: "Updated service 123 version 4",
		},
		{
			Args: "--service-id 123 --version 1 --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			Args: "--service-id 123 --version 1 --comment foo --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestVersionActivate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1 --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 4",
		},
		{
			Args: "--service-id 123 --version 2 --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 4",
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ActivateVersionFn: func(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
					return nil, fmt.Errorf("Cannot activate version %d. Versions that have been activated cannot be activated", i.ServiceVersion)
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: "Cannot activate version 3. Versions that have been activated cannot be activated",
		},
		{
			Name: "validate API error when modifying locked version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ActivateVersionFn: func(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
					return nil, fmt.Errorf("Cannot activate version %d. Versions that have been locked cannot be activated", i.ServiceVersion)
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: "Cannot activate version 3. Versions that have been locked cannot be activated",
		},
		{
			Args: "--service-id 123 --version 3",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 3",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			Args:       "--service-id 123 --version 3 --autoclone",
			WantOutput: "Activated service 123 version 4",
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				ActivateVersionFn: func(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected activation on cloned version 4, got %d", i.ServiceVersion)
					}
					return &fastly.Version{
						Number:    fastly.ToPointer(i.ServiceVersion),
						ServiceID: fastly.ToPointer("123"),
						Active:    fastly.ToPointer(true),
						Deployed:  fastly.ToPointer(true),
						CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
						UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
					}, nil
				},
			},
			Args:       "--service-id 123 --version 3 --autoclone",
			WantOutput: "Activated service 123 version 4",
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				ActivateVersionFn: func(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected activation on cloned version 4, got %d", i.ServiceVersion)
					}
					return &fastly.Version{
						Number:    fastly.ToPointer(i.ServiceVersion),
						ServiceID: fastly.ToPointer("123"),
						Active:    fastly.ToPointer(true),
						Deployed:  fastly.ToPointer(true),
						CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
						UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
					}, nil
				},
			},
			Args:       "--service-id 123 --version 3 --autoclone",
			WantOutput: "Activated service 123 version 4",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "activate"}, scenarios)
}

func TestVersionDeactivate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: deactivateVersionOK,
			},
			WantOutput: "Deactivated service 123 version 1",
		},
		{
			Args: "--service-id 123 --version 3",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: deactivateVersionOK,
			},
			WantError: "service version 3 is not active",
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: deactivateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "deactivate"}, scenarios)
}

func TestVersionLock(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:  testutil.GetVersion,
				LockVersionFn: lockVersionOK,
			},
			WantOutput: "Locked service 123 version 1",
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:  testutil.GetVersion,
				LockVersionFn: lockVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "lock"}, scenarios)
}

func TestVersionStage(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name: "validate --autoclone stages a clone of an active version",
			Args: "--service-id 123 --version 1 --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 4",
		},
		{
			Name: "validate --autoclone stages a clone of a locked version",
			Args: "--service-id 123 --version 2 --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 4",
		},
		{
			Args: "--service-id 123 --version 3",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ActivateVersionFn: stageVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Args: "--service-id 123 --version 3",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 3",
		},
		{
			Args: "--service-id 123 --version 4",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 4",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "stage"}, scenarios)
}

func TestVersionUnstage(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: unstageVersionOK,
			},
			WantError: "service version 1 is not staged",
		},
		{
			Args: "--service-id 123 --version 3",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: unstageVersionError,
			},
			WantError: "service version 3 is not staged",
		},
		{
			Args: "--service-id 123 --version 4",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: unstageVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Args: "--service-id 123 --version 4",
			API: &mock.API{
				GetVersionFn:        testutil.GetVersion,
				DeactivateVersionFn: unstageVersionOK,
			},
			WantOutput: "Unstaged service 123 version 4",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "unstage"}, scenarios)
}

var cloneServiceVersionJSONOutput = strings.TrimSpace(`
{
  "Active": null,
  "Comment": null,
  "CreatedAt": null,
  "DeletedAt": null,
  "Deployed": null,
  "Locked": null,
  "Number": 4,
  "ServiceID": "123",
  "Staging": null,
  "Testing": null,
  "UpdatedAt": null,
  "Environments": null
}
`) + "\n"

var listVersionsShortOutput = strings.TrimSpace(`
NUMBER  ACTIVE  STAGED  LAST EDITED (UTC)
4       false   true    2000-01-04 01:00
3       false   false   2000-01-03 01:00
2       false   false   2000-01-02 01:00
1       true    false   2000-01-01 01:00
`) + "\n"

var listVersionsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (auth: user)

Service ID (via --service-id): 123

Versions: 4
	Version 1/4
		Number: 4
		Service ID: 123
		Staged: true
		Last edited (UTC): 2000-01-04 01:00
	Version 2/4
		Number: 3
		Service ID: 123
		Last edited (UTC): 2000-01-03 01:00
	Version 3/4
		Number: 2
		Service ID: 123
		Locked: true
		Last edited (UTC): 2000-01-02 01:00
	Version 4/4
		Number: 1
		Service ID: 123
		Active: true
		Last edited (UTC): 2000-01-01 01:00
`) + "\n\n"

func updateVersionOK(_ context.Context, i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(true),
		Deployed:  fastly.ToPointer(true),
		Comment:   fastly.ToPointer("foo"),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func updateVersionError(_ context.Context, _ *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func activateVersionOK(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(true),
		Deployed:  fastly.ToPointer(true),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func activateVersionError(_ context.Context, _ *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func deactivateVersionOK(_ context.Context, i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(false),
		Deployed:  fastly.ToPointer(true),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func deactivateVersionError(_ context.Context, _ *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func stageVersionOK(_ context.Context, i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(true),
		Deployed:  fastly.ToPointer(true),
		Staging:   fastly.ToPointer(true),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func stageVersionError(_ context.Context, _ *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func unstageVersionOK(_ context.Context, i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(false),
		Deployed:  fastly.ToPointer(true),
		Staging:   fastly.ToPointer(false),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func unstageVersionError(_ context.Context, _ *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func lockVersionOK(_ context.Context, i *fastly.LockVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(false),
		Deployed:  fastly.ToPointer(true),
		Locked:    fastly.ToPointer(true),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func lockVersionError(_ context.Context, _ *fastly.LockVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func TestVersionValidate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 1",
			WantError: "error parsing arguments: required flag --service-id not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name: "validate successful - valid version without message",
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionValid(""),
			},
			WantOutput: "Service 123 version 1 is valid",
		},
		{
			Name: "validate successful - valid version with message",
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionValid("All checks passed"),
			},
			WantOutput: "Service 123 version 1 is valid: All checks passed",
		},
		{
			Name: "validate successful - invalid version without message",
			Args: "--service-id 123 --version 2",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionInvalid(""),
			},
			WantOutput: "Service 123 version 2 is not valid",
		},
		{
			Name: "validate successful - invalid version with message",
			Args: "--service-id 123 --version 2",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionInvalid("Missing required backend"),
			},
			WantOutput: "Service 123 version 2 is not valid: Missing required backend",
		},
		{
			Name: "validate with json output - valid version",
			Args: "--service-id 123 --version 1 --json",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionValid("All checks passed"),
			},
			WantOutput: validateVersionValidJSONOutput,
		},
		{
			Name: "validate with json output - invalid version",
			Args: "--service-id 123 --version 2 --json",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionInvalid("Missing required backend"),
			},
			WantOutput: validateVersionInvalidJSONOutput,
		},
		{
			Name: "validate error from API",
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ValidateVersionFn: validateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "validate"}, scenarios)
}

func validateVersionValid(message string) func(context.Context, *fastly.ValidateVersionInput) (bool, string, error) {
	return func(_ context.Context, _ *fastly.ValidateVersionInput) (bool, string, error) {
		return true, message, nil
	}
}

func validateVersionInvalid(message string) func(context.Context, *fastly.ValidateVersionInput) (bool, string, error) {
	return func(_ context.Context, _ *fastly.ValidateVersionInput) (bool, string, error) {
		return false, message, nil
	}
}

func validateVersionError(_ context.Context, _ *fastly.ValidateVersionInput) (bool, string, error) {
	return false, "", testutil.Err
}

var validateVersionValidJSONOutput = strings.TrimSpace(`
{
  "message": "All checks passed",
  "valid": true
}
`) + "\n"

var validateVersionInvalidJSONOutput = strings.TrimSpace(`
{
  "message": "Missing required backend",
  "valid": false
}
`) + "\n"
