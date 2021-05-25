package logentries_test

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

func TestLogentriesCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "logentries", "create", "--service-id", "123", "--version", "2", "--name", "log", "--port", "20000", "--autoclone"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				CreateLogentriesFn: createLogentriesOK,
			},
			wantOutput: "Created Logentries logging endpoint log (service 123 version 3)",
		},
		{
			args: []string{"logging", "logentries", "create", "--service-id", "123", "--version", "2", "--name", "log", "--port", "20000"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				CreateLogentriesFn: createLogentriesError,
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

func TestLogentriesList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "logentries", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListLogentriesFn: listLogentriesOK,
			},
			wantOutput: listLogentriesShortOutput,
		},
		{
			args: []string{"logging", "logentries", "list", "--service-id", "123", "--version", "2", "--verbose"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListLogentriesFn: listLogentriesOK,
			},
			wantOutput: listLogentriesVerboseOutput,
		},
		{
			args: []string{"logging", "logentries", "list", "--service-id", "123", "--version", "2", "-v"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListLogentriesFn: listLogentriesOK,
			},
			wantOutput: listLogentriesVerboseOutput,
		},
		{
			args: []string{"logging", "logentries", "--verbose", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListLogentriesFn: listLogentriesOK,
			},
			wantOutput: listLogentriesVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "logentries", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListLogentriesFn: listLogentriesOK,
			},
			wantOutput: listLogentriesVerboseOutput,
		},
		{
			args: []string{"logging", "logentries", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListLogentriesFn: listLogentriesError,
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
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestLogentriesDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logentries", "describe", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "logentries", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:  listVersionsOK,
				GetVersionFn:    getVersionOK,
				GetLogentriesFn: getLogentriesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "logentries", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:  listVersionsOK,
				GetVersionFn:    getVersionOK,
				GetLogentriesFn: getLogentriesOK,
			},
			wantOutput: describeLogentriesOutput,
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

func TestLogentriesUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logentries", "update", "--service-id", "123", "--version", "2", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "logentries", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				UpdateLogentriesFn: updateLogentriesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "logentries", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				UpdateLogentriesFn: updateLogentriesOK,
			},
			wantOutput: "Updated Logentries logging endpoint log (service 123 version 3)",
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

func TestLogentriesDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logentries", "delete", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "logentries", "delete", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				DeleteLogentriesFn: deleteLogentriesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "logentries", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				DeleteLogentriesFn: deleteLogentriesOK,
			},
			wantOutput: "Deleted Logentries logging endpoint logs (service 123 version 3)",
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

func createLogentriesOK(i *fastly.CreateLogentriesInput) (*fastly.Logentries, error) {
	return &fastly.Logentries{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createLogentriesError(i *fastly.CreateLogentriesInput) (*fastly.Logentries, error) {
	return nil, errTest
}

func listLogentriesOK(i *fastly.ListLogentriesInput) ([]*fastly.Logentries, error) {
	return []*fastly.Logentries{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Port:              20000,
			UseTLS:            true,
			Token:             "tkn",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Port:              20001,
			UseTLS:            false,
			Token:             "tkn1",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listLogentriesError(i *fastly.ListLogentriesInput) ([]*fastly.Logentries, error) {
	return nil, errTest
}

var listLogentriesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      2        logs
123      2        analytics
`) + "\n"

var listLogentriesVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 2
	Logentries 1/2
		Service ID: 123
		Version: 2
		Name: logs
		Port: 20000
		Use TLS: true
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Logentries 2/2
		Service ID: 123
		Version: 2
		Name: analytics
		Port: 20001
		Use TLS: false
		Token: tkn1
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getLogentriesOK(i *fastly.GetLogentriesInput) (*fastly.Logentries, error) {
	return &fastly.Logentries{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Port:              20000,
		UseTLS:            true,
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getLogentriesError(i *fastly.GetLogentriesInput) (*fastly.Logentries, error) {
	return nil, errTest
}

var describeLogentriesOutput = strings.TrimSpace(`
Service ID: 123
Version: 2
Name: logs
Port: 20000
Use TLS: true
Token: tkn
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateLogentriesOK(i *fastly.UpdateLogentriesInput) (*fastly.Logentries, error) {
	return &fastly.Logentries{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Port:              20000,
		UseTLS:            true,
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateLogentriesError(i *fastly.UpdateLogentriesInput) (*fastly.Logentries, error) {
	return nil, errTest
}

func deleteLogentriesOK(i *fastly.DeleteLogentriesInput) error {
	return nil
}

func deleteLogentriesError(i *fastly.DeleteLogentriesInput) error {
	return errTest
}

func listVersionsOK(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func getVersionOK(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		ServiceID: i.ServiceID,
		Number:    2,
		Active:    true,
		UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}, nil
}

func cloneVersionOK(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion + 1}, nil
}
