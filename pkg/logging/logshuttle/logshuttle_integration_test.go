package logshuttle_test

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

func TestLogshuttleCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logshuttle", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args:      []string{"logging", "logshuttle", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args:       []string{"logging", "logshuttle", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com", "--auth-token", "abc"},
			api:        mock.API{CreateLogshuttleFn: createLogshuttleOK},
			wantOutput: "Created Logshuttle logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "logshuttle", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com", "--auth-token", "abc"},
			api:       mock.API{CreateLogshuttleFn: createLogshuttleError},
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

func TestLogshuttleList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "logshuttle", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListLogshuttlesFn: listLogshuttlesOK},
			wantOutput: listLogshuttlesShortOutput,
		},
		{
			args:       []string{"logging", "logshuttle", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListLogshuttlesFn: listLogshuttlesOK},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args:       []string{"logging", "logshuttle", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListLogshuttlesFn: listLogshuttlesOK},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args:       []string{"logging", "logshuttle", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListLogshuttlesFn: listLogshuttlesOK},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "logshuttle", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListLogshuttlesFn: listLogshuttlesOK},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args:      []string{"logging", "logshuttle", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListLogshuttlesFn: listLogshuttlesError},
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

func TestLogshuttleDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logshuttle", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "logshuttle", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetLogshuttleFn: getLogshuttleError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "logshuttle", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetLogshuttleFn: getLogshuttleOK},
			wantOutput: describeLogshuttleOutput,
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

func TestLogshuttleUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logshuttle", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "logshuttle", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetLogshuttleFn:    getLogshuttleError,
				UpdateLogshuttleFn: updateLogshuttleOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "logshuttle", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetLogshuttleFn:    getLogshuttleOK,
				UpdateLogshuttleFn: updateLogshuttleError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "logshuttle", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetLogshuttleFn:    getLogshuttleOK,
				UpdateLogshuttleFn: updateLogshuttleOK,
			},
			wantOutput: "Updated Logshuttle logging endpoint log (service 123 version 1)",
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

func TestLogshuttleDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "logshuttle", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "logshuttle", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteLogshuttleFn: deleteLogshuttleError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "logshuttle", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteLogshuttleFn: deleteLogshuttleOK},
			wantOutput: "Deleted Logshuttle logging endpoint logs (service 123 version 1)",
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

func createLogshuttleOK(i *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error) {
	s := fastly.Logshuttle{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createLogshuttleError(i *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error) {
	return nil, errTest
}

func listLogshuttlesOK(i *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error) {
	return []*fastly.Logshuttle{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			URL:               "example.com",
			Token:             "abc",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			URL:               "example.com",
			Token:             "abc",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listLogshuttlesError(i *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error) {
	return nil, errTest
}

var listLogshuttlesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listLogshuttlesVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Logshuttle 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Logshuttle 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: example.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getLogshuttleOK(i *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return &fastly.Logshuttle{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		URL:               "example.com",
		Token:             "abc",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getLogshuttleError(i *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return nil, errTest
}

var describeLogshuttleOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
URL: example.com
Token: abc
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateLogshuttleOK(i *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error) {
	return &fastly.Logshuttle{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		URL:               "example.com",
		Token:             "abc",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateLogshuttleError(i *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error) {
	return nil, errTest
}

func deleteLogshuttleOK(i *fastly.DeleteLogshuttleInput) error {
	return nil
}

func deleteLogshuttleError(i *fastly.DeleteLogshuttleInput) error {
	return errTest
}
