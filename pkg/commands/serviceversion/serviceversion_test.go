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
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--version 1",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name: "validate successful clone",
			Arg:  "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantOutput: "Cloned service 123 version 1 to version 4",
		},
		{
			Name: "validate error will be passed through if cloning fails",
			Arg:  "--service-id 456 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "clone"}, scenarios)
}

func TestVersionList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:        "--service-id 123",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsShortOutput,
		},
		{
			Arg:        "--service-id 123 --verbose",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Arg:        "--service-id 123 -v",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Arg:        "--verbose --service-id 123",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Arg:        "-v --service-id 123",
			API:        mock.API{ListVersionsFn: testutil.ListVersions},
			WantOutput: listVersionsVerboseOutput,
		},
		{
			Arg:       "--service-id 123",
			API:       mock.API{ListVersionsFn: testutil.ListVersionsError},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestVersionUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg: "--service-id 123 --version 1 --comment foo --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionOK,
			},
			WantOutput: "Updated service 123 version 4",
		},
		{
			Arg: "--service-id 123 --version 1 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			Arg: "--service-id 123 --version 1 --comment foo --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func TestVersionActivate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			WantError: "service version 1 is active",
		},
		{
			Arg: "--service-id 123 --version 1 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
		{
			Arg: "--service-id 123 --version 1 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 4",
		},
		{
			Arg: "--service-id 123 --version 2 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 4",
		},
		{
			Arg: "--service-id 123 --version 3 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: activateVersionOK,
			},
			WantOutput: "Activated service 123 version 3",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "activate"}, scenarios)
}

func TestVersionDeactivate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			WantOutput: "Deactivated service 123 version 1",
		},
		{
			Arg: "--service-id 123 --version 3",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			WantError: "service version 3 is not active",
		},
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "deactivate"}, scenarios)
}

func TestVersionLock(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Arg:       "--service-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionOK,
			},
			WantOutput: "Locked service 123 version 1",
		},
		{
			Arg: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionError,
			},
			WantError: testutil.Err.Error(),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "lock"}, scenarios)
}

var listVersionsShortOutput = strings.TrimSpace(`
NUMBER  ACTIVE  LAST EDITED (UTC)
1       true    2000-01-01 01:00
2       false   2000-01-02 01:00
3       false   2000-01-03 01:00
`) + "\n"

var listVersionsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Versions: 3
	Version 1/3
		Number: 1
		Service ID: 123
		Active: true
		Last edited (UTC): 2000-01-01 01:00
	Version 2/3
		Number: 2
		Service ID: 123
		Active: false
		Locked: true
		Last edited (UTC): 2000-01-02 01:00
	Version 3/3
		Number: 3
		Service ID: 123
		Active: false
		Last edited (UTC): 2000-01-03 01:00
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
