package serviceversion_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestVersionClone(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service-version", "clone", "--version", "1"},
			wantError: "error reading service: no service ID found",
		},
		{
			args:      []string{"service-version", "clone", "--service-id", "123"},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: []string{"service-version", "clone", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantOutput: "Cloned service 123 version 1 to version 4",
		},
		{
			args: []string{"service-version", "clone", "--service-id", "456", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestVersionList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"service-version", "list", "--service-id", "123"},
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsShortOutput,
		},
		{
			args:       []string{"service-version", "list", "--service-id", "123", "--verbose"},
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       []string{"service-version", "list", "--service-id", "123", "-v"},
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       []string{"service-version", "--verbose", "list", "--service-id", "123"},
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       []string{"-v", "service-version", "list", "--service-id", "123"},
			api:        mock.API{ListVersionsFn: testutil.ListVersions},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:      []string{"service-version", "list", "--service-id", "123"},
			api:       mock.API{ListVersionsFn: testutil.ListVersionsError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestVersionUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"service-version", "update", "--service-id", "123", "--version", "1", "--comment", "foo", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionOK,
			},
			wantOutput: "Updated service 123 version 4",
		},
		{
			args: []string{"service-version", "update", "--service-id", "123", "--version", "1", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			args: []string{"service-version", "update", "--service-id", "123", "--version", "1", "--comment", "foo", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateVersionFn: updateVersionError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestVersionActivate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service-version", "activate", "--service-id", "123"},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: []string{"service-version", "activate", "--service-id", "123", "--version", "1", "--autoclone"},
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"service-version", "activate", "--service-id", "123", "--version", "1", "--autoclone"},
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				ActivateVersionFn: activateVersionOK,
			},
			wantOutput: "Activated service 123 version 4",
		},
		{
			args: []string{"service-version", "activate", "--service-id", "123", "--version", "3", "--autoclone"},
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ActivateVersionFn: activateVersionOK,
			},
			wantOutput: "Activated service 123 version 3",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestVersionDeactivate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service-version", "deactivate", "--service-id", "123"},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: []string{"service-version", "deactivate", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			wantOutput: "Deactivated service 123 version 1",
		},
		{
			args: []string{"service-version", "deactivate", "--service-id", "123", "--version", "3"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionOK,
			},
			wantOutput: "Deactivated service 123 version 3",
		},
		{
			args: []string{"service-version", "deactivate", "--service-id", "123", "--version", "3"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				DeactivateVersionFn: deactivateVersionError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestVersionLock(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"service-version", "lock", "--service-id", "123"},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args: []string{"service-version", "lock", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionOK,
			},
			wantOutput: "Locked service 123 version 1",
		},
		{
			args: []string{"service-version", "lock", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				LockVersionFn:  lockVersionError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

var listVersionsShortOutput = strings.TrimSpace(`
NUMBER  ACTIVE  LAST EDITED (UTC)
1       true    2000-01-01 01:00
2       false   2000-01-02 01:00
3       false   2000-01-03 01:00
`) + "\n"

var listVersionsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
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

func updateVersionError(i *fastly.UpdateVersionInput) (*fastly.Version, error) {
	return nil, errTest
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

func activateVersionError(i *fastly.ActivateVersionInput) (*fastly.Version, error) {
	return nil, errTest
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

func deactivateVersionError(i *fastly.DeactivateVersionInput) (*fastly.Version, error) {
	return nil, errTest
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

func lockVersionError(i *fastly.LockVersionInput) (*fastly.Version, error) {
	return nil, errTest
}
