package serviceversion_test

import (
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/serviceversion"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersionClone(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 1",
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
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantOutput: "Cloned service 123 version 1 to version 4",
		},
		{
			Name: "validate error will be passed through if cloning fails",
			Args: "--service-id 456 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "clone"}, scenarios)
}

func TestVersionList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:       "--service-id 123",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsShortOutput,
		},
		{
			Args:       "--service-id 123 --verbose",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:       "--service-id 123 -v",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:       "--verbose --service-id 123",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:       "-v --service-id 123",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Args:      "--service-id 123",
			API:       mock.API{ListVersionsFn: testutil.ListVersionsError},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestVersionUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --comment foo --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionOK,
			},
			WantOutput: "Updated service 123 version 4",
		},
		{
			Args: "--service-id 123 --version 1 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			Args: "--service-id 123 --version 1 --comment foo --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func TestVersionActivate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			WantError: "service version 1 is active",
		},
		{
			Args: "--service-id 123 --version 1 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 4",
		},
		{
			Args: "--service-id 123 --version 2 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 4",
		},
		{
			Args: "--service-id 123 --version 3 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 3",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "activate"}, scenarios)
}

func TestVersionDeactivate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			WantOutput: "Deactivated service 123 version 1",
		},
		{
			Args: "--service-id 123 --version 3",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			WantError: "service version 3 is not active",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "deactivate"}, scenarios)
}

func TestVersionLock(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionOK,
			},
			WantOutput: "Locked service 123 version 1",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "lock"}, scenarios)
}

func TestVersionStage(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: stageVersionOK,
			},
			WantError: "service version 1 is active",
		},
		{
			Args: "--service-id 123 --version 2",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: stageVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Args: "--service-id 123 --version 2",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 2",
		},
		{
			Args: "--service-id 123 --version 3",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 3",
		},
		{
			Args: "--service-id 123 --version 4",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: stageVersionOK,
			},
			WantOutput: "Staged service 123 version 4",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "stage"}, scenarios)
}

func TestVersionUnstage(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: unstageVersionOK,
			},
			WantError: "service version 1 is not staged",
		},
		{
			Args: "--service-id 123 --version 3",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: unstageVersionError,
			},
			WantError: "service version 3 is not staged",
		},
		{
			Args: "--service-id 123 --version 4",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: unstageVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Args: "--service-id 123 --version 4",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: unstageVersionOK,
			},
			WantOutput: "Unstaged service 123 version 4",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "unstage"}, scenarios)
}

var listVersionsShortOutput = strings.TrimSpace(`
NUMBER  ACTIVE  STAGING  LAST EDITED (UTC)
1       true    false    2000-01-01 01:00
2       false   false    2000-01-02 01:00
3       false   false    2000-01-03 01:00
4       false   true     2000-01-04 01:00
`) + "\n"

var listVersionsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Versions: 4
	Version 1/4
		Number: 1
		Service ID: 123
		Active: true
		Last edited (UTC): 2000-01-01 01:00
	Version 2/4
		Number: 2
		Service ID: 123
		Locked: true
		Last edited (UTC): 2000-01-02 01:00
	Version 3/4
		Number: 3
		Service ID: 123
		Last edited (UTC): 2000-01-03 01:00
	Version 4/4
		Number: 4
		Service ID: 123
		Staging: true
		Last edited (UTC): 2000-01-04 01:00
`) + "\n\n"

func updateVersionOK(i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(true),
		Deployed:  fastly.ToPointer(true),
		Comment:   fastly.ToPointer("foo"),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func updateVersionError(_ *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func activateVersionOK(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(true),
		Deployed:  fastly.ToPointer(true),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func activateVersionError(_ *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func deactivateVersionOK(i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    fastly.ToPointer(i.ServiceVersion),
		ServiceID: fastly.ToPointer("123"),
		Active:    fastly.ToPointer(false),
		Deployed:  fastly.ToPointer(true),
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func deactivateVersionError(_ *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func stageVersionOK(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
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

func stageVersionError(_ *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func unstageVersionOK(i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
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

func unstageVersionError(_ *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func lockVersionOK(i *fastly.LockVersionInput) (*fastly.Version, error) {
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

func lockVersionError(_ *fastly.LockVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}
