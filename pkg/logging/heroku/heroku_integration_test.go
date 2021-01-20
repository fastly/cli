package heroku_test

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

func TestHerokuCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "heroku", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args:      []string{"logging", "heroku", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args:       []string{"logging", "heroku", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc", "--url", "example.com"},
			api:        mock.API{CreateHerokuFn: createHerokuOK},
			wantOutput: "Created Heroku logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "heroku", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc", "--url", "example.com"},
			api:       mock.API{CreateHerokuFn: createHerokuError},
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

func TestHerokuList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "heroku", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHerokusFn: listHerokusOK},
			wantOutput: listHerokusShortOutput,
		},
		{
			args:       []string{"logging", "heroku", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListHerokusFn: listHerokusOK},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args:       []string{"logging", "heroku", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListHerokusFn: listHerokusOK},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args:       []string{"logging", "heroku", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHerokusFn: listHerokusOK},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "heroku", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHerokusFn: listHerokusOK},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args:      []string{"logging", "heroku", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListHerokusFn: listHerokusError},
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

func TestHerokuDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "heroku", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "heroku", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetHerokuFn: getHerokuError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "heroku", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetHerokuFn: getHerokuOK},
			wantOutput: describeHerokuOutput,
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

func TestHerokuUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "heroku", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "heroku", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHerokuFn:    getHerokuError,
				UpdateHerokuFn: updateHerokuOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "heroku", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHerokuFn:    getHerokuOK,
				UpdateHerokuFn: updateHerokuError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "heroku", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHerokuFn:    getHerokuOK,
				UpdateHerokuFn: updateHerokuOK,
			},
			wantOutput: "Updated Heroku logging endpoint log (service 123 version 1)",
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

func TestHerokuDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "heroku", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "heroku", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteHerokuFn: deleteHerokuError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "heroku", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteHerokuFn: deleteHerokuOK},
			wantOutput: "Deleted Heroku logging endpoint logs (service 123 version 1)",
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

func createHerokuOK(i *fastly.CreateHerokuInput) (*fastly.Heroku, error) {
	s := fastly.Heroku{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createHerokuError(i *fastly.CreateHerokuInput) (*fastly.Heroku, error) {
	return nil, errTest
}

func listHerokusOK(i *fastly.ListHerokusInput) ([]*fastly.Heroku, error) {
	return []*fastly.Heroku{
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
			URL:               "bar.com",
			Token:             "abc",
			Format:            `%h %l %u %t "%r" %>s %b`,
			ResponseCondition: "Prevent default logging",
			FormatVersion:     2,
			Placement:         "none",
		},
	}, nil
}

func listHerokusError(i *fastly.ListHerokusInput) ([]*fastly.Heroku, error) {
	return nil, errTest
}

var listHerokusShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHerokusVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Heroku 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Heroku 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: bar.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getHerokuOK(i *fastly.GetHerokuInput) (*fastly.Heroku, error) {
	return &fastly.Heroku{
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

func getHerokuError(i *fastly.GetHerokuInput) (*fastly.Heroku, error) {
	return nil, errTest
}

var describeHerokuOutput = strings.TrimSpace(`
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

func updateHerokuOK(i *fastly.UpdateHerokuInput) (*fastly.Heroku, error) {
	return &fastly.Heroku{
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

func updateHerokuError(i *fastly.UpdateHerokuInput) (*fastly.Heroku, error) {
	return nil, errTest
}

func deleteHerokuOK(i *fastly.DeleteHerokuInput) error {
	return nil
}

func deleteHerokuError(i *fastly.DeleteHerokuInput) error {
	return errTest
}
