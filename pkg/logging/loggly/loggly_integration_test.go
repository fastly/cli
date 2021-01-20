package loggly_test

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

func TestLogglyCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "loggly", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args:       []string{"logging", "loggly", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			api:        mock.API{CreateLogglyFn: createLogglyOK},
			wantOutput: "Created Loggly logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "loggly", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			api:       mock.API{CreateLogglyFn: createLogglyError},
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

func TestLogglyList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "loggly", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListLogglyFn: listLogglysOK},
			wantOutput: listLogglysShortOutput,
		},
		{
			args:       []string{"logging", "loggly", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListLogglyFn: listLogglysOK},
			wantOutput: listLogglysVerboseOutput,
		},
		{
			args:       []string{"logging", "loggly", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListLogglyFn: listLogglysOK},
			wantOutput: listLogglysVerboseOutput,
		},
		{
			args:       []string{"logging", "loggly", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListLogglyFn: listLogglysOK},
			wantOutput: listLogglysVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "loggly", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListLogglyFn: listLogglysOK},
			wantOutput: listLogglysVerboseOutput,
		},
		{
			args:      []string{"logging", "loggly", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListLogglyFn: listLogglysError},
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

func TestLogglyDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "loggly", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "loggly", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetLogglyFn: getLogglyError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "loggly", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetLogglyFn: getLogglyOK},
			wantOutput: describeLogglyOutput,
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

func TestLogglyUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "loggly", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "loggly", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetLogglyFn:    getLogglyError,
				UpdateLogglyFn: updateLogglyOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "loggly", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetLogglyFn:    getLogglyOK,
				UpdateLogglyFn: updateLogglyError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "loggly", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetLogglyFn:    getLogglyOK,
				UpdateLogglyFn: updateLogglyOK,
			},
			wantOutput: "Updated Loggly logging endpoint log (service 123 version 1)",
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

func TestLogglyDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "loggly", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "loggly", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteLogglyFn: deleteLogglyError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "loggly", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteLogglyFn: deleteLogglyOK},
			wantOutput: "Deleted Loggly logging endpoint logs (service 123 version 1)",
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

func createLogglyOK(i *fastly.CreateLogglyInput) (*fastly.Loggly, error) {
	s := fastly.Loggly{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createLogglyError(i *fastly.CreateLogglyInput) (*fastly.Loggly, error) {
	return nil, errTest
}

func listLogglysOK(i *fastly.ListLogglyInput) ([]*fastly.Loggly, error) {
	return []*fastly.Loggly{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Token:             "abc",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Token:             "abc",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listLogglysError(i *fastly.ListLogglyInput) ([]*fastly.Loggly, error) {
	return nil, errTest
}

var listLogglysShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listLogglysVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Loggly 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Loggly 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getLogglyOK(i *fastly.GetLogglyInput) (*fastly.Loggly, error) {
	return &fastly.Loggly{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Token:             "abc",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getLogglyError(i *fastly.GetLogglyInput) (*fastly.Loggly, error) {
	return nil, errTest
}

var describeLogglyOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Token: abc
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateLogglyOK(i *fastly.UpdateLogglyInput) (*fastly.Loggly, error) {
	return &fastly.Loggly{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Token:             "abc",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
	}, nil
}

func updateLogglyError(i *fastly.UpdateLogglyInput) (*fastly.Loggly, error) {
	return nil, errTest
}

func deleteLogglyOK(i *fastly.DeleteLogglyInput) error {
	return nil
}

func deleteLogglyError(i *fastly.DeleteLogglyInput) error {
	return errTest
}
