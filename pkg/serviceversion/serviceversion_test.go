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
			api:       mock.API{CloneVersionFn: cloneVersionOK},
			wantError: "error reading service: no service ID found",
		},
		{
			args:      []string{"service-version", "clone", "--service-id", "123"},
			api:       mock.API{CloneVersionFn: cloneVersionOK},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args:       []string{"service-version", "clone", "--service-id", "123", "--version", "1"},
			api:        mock.API{CloneVersionFn: cloneVersionOK},
			wantOutput: "Cloned service 123 version 1 to version 2",
		},
		{
			args:      []string{"service-version", "clone", "--service-id", "123", "--version", "1"},
			api:       mock.API{CloneVersionFn: cloneVersionError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
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
			api:        mock.API{ListVersionsFn: listVersionsOK},
			wantOutput: listVersionsShortOutput,
		},
		{
			args:       []string{"service-version", "list", "--service-id", "123", "--verbose"},
			api:        mock.API{ListVersionsFn: listVersionsOK},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       []string{"service-version", "list", "--service-id", "123", "-v"},
			api:        mock.API{ListVersionsFn: listVersionsOK},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       []string{"service-version", "--verbose", "list", "--service-id", "123"},
			api:        mock.API{ListVersionsFn: listVersionsOK},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:       []string{"-v", "service-version", "list", "--service-id", "123"},
			api:        mock.API{ListVersionsFn: listVersionsOK},
			wantOutput: listVersionsVerboseOutput,
		},
		{
			args:      []string{"service-version", "list", "--service-id", "123"},
			api:       mock.API{ListVersionsFn: listVersionsError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
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
			args: []string{"service-version", "update", "--service-id", "123", "--version", "1", "--comment", "foo"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateVersionFn: updateVersionOK,
			},
			wantOutput: "Updated service 123 version 1",
		},
		{
			args:      []string{"service-version", "update", "--service-id", "123", "--version", "2"},
			api:       mock.API{UpdateVersionFn: updateVersionOK},
			wantError: "error parsing arguments: required flag --comment not provided",
		},
		{
			args: []string{"service-version", "update", "--service-id", "123", "--version", "1", "--comment", "foo"},
			api: mock.API{
				GetServiceFn:    getServiceOK,
				UpdateVersionFn: updateVersionError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
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
			api:       mock.API{ActivateVersionFn: activateVersionOK},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args:       []string{"service-version", "activate", "--service-id", "123", "--version", "1"},
			api:        mock.API{ActivateVersionFn: activateVersionOK},
			wantOutput: "Activated service 123 version 1",
		},
		{
			args:      []string{"service-version", "activate", "--service-id", "123", "--version", "1"},
			api:       mock.API{ActivateVersionFn: activateVersionError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
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
			api:       mock.API{DeactivateVersionFn: deactivateVersionOK},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args:       []string{"service-version", "deactivate", "--service-id", "123", "--version", "1"},
			api:        mock.API{DeactivateVersionFn: deactivateVersionOK},
			wantOutput: "Deactivated service 123 version 1",
		},
		{
			args:      []string{"service-version", "deactivate", "--service-id", "123", "--version", "1"},
			api:       mock.API{DeactivateVersionFn: deactivateVersionError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
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
			api:       mock.API{LockVersionFn: lockVersionOK},
			wantError: "error parsing arguments: required flag --version not provided",
		},
		{
			args:       []string{"service-version", "lock", "--service-id", "123", "--version", "1"},
			api:        mock.API{LockVersionFn: lockVersionOK},
			wantOutput: "Locked service 123 version 1",
		},
		{
			args:      []string{"service-version", "lock", "--service-id", "123", "--version", "1"},
			api:       mock.API{LockVersionFn: lockVersionError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func cloneVersionOK(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		Number:    i.ServiceVersion + 1,
		ServiceID: "123",
		Active:    true,
		Deployed:  true,
		CreatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
	}, nil
}

func cloneVersionError(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return nil, errTest
}

func listVersionsOK(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			Number:    1,
			Comment:   "a",
			ServiceID: "b",
			CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt: testutil.MustParseTimeRFC3339("2010-11-15T19:01:02Z"),
		},
		{
			Number:    2,
			Comment:   "c",
			ServiceID: "b",
			Active:    true,
			Deployed:  true,
			CreatedAt: testutil.MustParseTimeRFC3339("2001-03-03T04:05:06Z"),
			UpdatedAt: testutil.MustParseTimeRFC3339("2015-03-14T12:59:59Z"),
		},
	}, nil
}

func listVersionsError(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return nil, errTest
}

var listVersionsShortOutput = strings.TrimSpace(`
NUMBER  ACTIVE  LAST EDITED (UTC)
1       false   2010-11-15 19:01
2       true    2015-03-14 12:59
`) + "\n"

var listVersionsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Versions: 2
	Version 1/2
		Number: 1
		Comment: a
		Service ID: b
		Active: false
		Locked: false
		Deployed: false
		Staging: false
		Testing: false
		Created (UTC): 2001-02-03 04:05
		Last edited (UTC): 2010-11-15 19:01
	Version 2/2
		Number: 2
		Comment: c
		Service ID: b
		Active: true
		Locked: false
		Deployed: true
		Staging: false
		Testing: false
		Created (UTC): 2001-03-03 04:05
		Last edited (UTC): 2015-03-14 12:59
`) + "\n\n"

func getServiceOK(i *fastly.GetServiceInput) (*fastly.Service, error) {
	return &fastly.Service{
		ID:      "12345",
		Name:    "Foo",
		Comment: "Bar",
	}, nil
}

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
