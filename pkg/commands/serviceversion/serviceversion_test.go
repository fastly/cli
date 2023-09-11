package serviceversion_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersionClone(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-version clone --version 1"),
			wantError: "error reading service: no service ID found",
		},
		{
			args:      args("service-version clone --service-id 123"),
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: args("service-version clone --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantOutput: "Cloned service 123 version 1 to version 4",
		},
		{
			args: args("service-version clone --service-id 456 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionError,
			},
			wantError: testutil.Err.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestVersionList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       args("service-version list --service-id 123"),
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsShortOutput,
		},
		{
			args:       args("service-version list --service-id 123 --verbose"),
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       args("service-version list --service-id 123 -v"),
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       args("service-version --verbose list --service-id 123"),
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       args("-v service-version list --service-id 123"),
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:      args("service-version list --service-id 123"),
			api:       mock.API{ListVersionsFn: testutil.ListVersionsError},
			wantError: testutil.Err.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestVersionUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("service-version update --service-id 123 --version 1 --comment foo --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionOK,
			},
			wantOutput: "Updated service 123 version 4",
		},
		{
			args: args("service-version update --service-id 123 --version 1 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			args: args("service-version update --service-id 123 --version 1 --comment foo --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionError,
			},
			wantError: testutil.Err.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestVersionActivate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-version activate --service-id 123"),
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: args("service-version activate --service-id 123 --version 1 --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionError,
			},
			wantError: testutil.Err.Error(),
		},
		{
			args: args("service-version activate --service-id 123 --version 1 --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			wantOutput: "Activated service 123 version 4",
		},
		{
			args: args("service-version activate --service-id 123 --version 3 --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: activateVersionOK,
			},
			wantOutput: "Activated service 123 version 3",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestVersionDeactivate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-version deactivate --service-id 123"),
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: args("service-version deactivate --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			wantOutput: "Deactivated service 123 version 1",
		},
		{
			args: args("service-version deactivate --service-id 123 --version 3"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			wantOutput: "Deactivated service 123 version 3",
		},
		{
			args: args("service-version deactivate --service-id 123 --version 3"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionError,
			},
			wantError: testutil.Err.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestVersionLock(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-version lock --service-id 123"),
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: args("service-version lock --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionOK,
			},
			wantOutput: "Locked service 123 version 1",
		},
		{
			args: args("service-version lock --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionError,
			},
			wantError: testutil.Err.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
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
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Last edited (UTC): 2000-01-01 01:00
	Version 2/3
		Number: 2
		Service ID: 123
		Active: false
		Locked: true
		Deployed: false
		Staging: false
		Testing: false
		Last edited (UTC): 2000-01-02 01:00
	Version 3/3
		Number: 3
		Service ID: 123
		Active: false
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Last edited (UTC): 2000-01-03 01:00
`) + "\n\n"

func updateVersionOK(i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    i.ServiceVersion,
		ServiceID: "123",
		Active:    true,
		Deployed:  true,
		Comment:   "foo",
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func updateVersionError(_ *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func activateVersionOK(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    i.ServiceVersion,
		ServiceID: "123",
		Active:    true,
		Deployed:  true,
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func activateVersionError(_ *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func deactivateVersionOK(i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    i.ServiceVersion,
		ServiceID: "123",
		Active:    false,
		Deployed:  true,
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func deactivateVersionError(_ *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}

func lockVersionOK(i *fastly.LockVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    i.ServiceVersion,
		ServiceID: "123",
		Active:    false,
		Deployed:  true,
		Locked:    true,
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func lockVersionError(_ *fastly.LockVersionInput) (*fastly.Version, error) {
	return nil, testutil.Err
}
