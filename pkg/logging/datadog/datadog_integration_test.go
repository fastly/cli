package datadog_test

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

func TestDatadogCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "datadog", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args:       []string{"logging", "datadog", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			api:        mock.API{CreateDatadogFn: createDatadogOK},
			wantOutput: "Created Datadog logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "datadog", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			api:       mock.API{CreateDatadogFn: createDatadogError},
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

func TestDatadogList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "datadog", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDatadogFn: listDatadogsOK},
			wantOutput: listDatadogsShortOutput,
		},
		{
			args:       []string{"logging", "datadog", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListDatadogFn: listDatadogsOK},
			wantOutput: listDatadogsVerboseOutput,
		},
		{
			args:       []string{"logging", "datadog", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListDatadogFn: listDatadogsOK},
			wantOutput: listDatadogsVerboseOutput,
		},
		{
			args:       []string{"logging", "datadog", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDatadogFn: listDatadogsOK},
			wantOutput: listDatadogsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "datadog", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDatadogFn: listDatadogsOK},
			wantOutput: listDatadogsVerboseOutput,
		},
		{
			args:      []string{"logging", "datadog", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListDatadogFn: listDatadogsError},
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

func TestDatadogDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "datadog", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "datadog", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetDatadogFn: getDatadogError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "datadog", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetDatadogFn: getDatadogOK},
			wantOutput: describeDatadogOutput,
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

func TestDatadogUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "datadog", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "datadog", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetDatadogFn:    getDatadogError,
				UpdateDatadogFn: updateDatadogOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "datadog", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetDatadogFn:    getDatadogOK,
				UpdateDatadogFn: updateDatadogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "datadog", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetDatadogFn:    getDatadogOK,
				UpdateDatadogFn: updateDatadogOK,
			},
			wantOutput: "Updated Datadog logging endpoint log (service 123 version 1)",
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

func TestDatadogDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "datadog", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "datadog", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteDatadogFn: deleteDatadogError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "datadog", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteDatadogFn: deleteDatadogOK},
			wantOutput: "Deleted Datadog logging endpoint logs (service 123 version 1)",
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

func createDatadogOK(i *fastly.CreateDatadogInput) (*fastly.Datadog, error) {
	s := fastly.Datadog{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createDatadogError(i *fastly.CreateDatadogInput) (*fastly.Datadog, error) {
	return nil, errTest
}

func listDatadogsOK(i *fastly.ListDatadogInput) ([]*fastly.Datadog, error) {
	return []*fastly.Datadog{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Token:             "abc",
			Region:            "US",
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
			Region:            "US",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listDatadogsError(i *fastly.ListDatadogInput) ([]*fastly.Datadog, error) {
	return nil, errTest
}

var listDatadogsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listDatadogsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Datadog 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Datadog 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getDatadogOK(i *fastly.GetDatadogInput) (*fastly.Datadog, error) {
	return &fastly.Datadog{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Token:             "abc",
		Region:            "US",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getDatadogError(i *fastly.GetDatadogInput) (*fastly.Datadog, error) {
	return nil, errTest
}

var describeDatadogOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Token: abc
Region: US
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateDatadogOK(i *fastly.UpdateDatadogInput) (*fastly.Datadog, error) {
	return &fastly.Datadog{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Token:             "abc",
		Region:            "US",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
	}, nil
}

func updateDatadogError(i *fastly.UpdateDatadogInput) (*fastly.Datadog, error) {
	return nil, errTest
}

func deleteDatadogOK(i *fastly.DeleteDatadogInput) error {
	return nil
}

func deleteDatadogError(i *fastly.DeleteDatadogInput) error {
	return errTest
}
