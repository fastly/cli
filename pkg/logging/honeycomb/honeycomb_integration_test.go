package honeycomb_test

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

func TestHoneycombCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "honeycomb", "create", "--service-id", "123", "--version", "1", "--name", "log", "--dataset", "log"},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args:      []string{"logging", "honeycomb", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			wantError: "error parsing arguments: required flag --dataset not provided",
		},
		{
			args:       []string{"logging", "honeycomb", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc", "--dataset", "log"},
			api:        mock.API{CreateHoneycombFn: createHoneycombOK},
			wantOutput: "Created Honeycomb logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "honeycomb", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc", "--dataset", "log"},
			api:       mock.API{CreateHoneycombFn: createHoneycombError},
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

func TestHoneycombList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "honeycomb", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHoneycombsFn: listHoneycombsOK},
			wantOutput: listHoneycombsShortOutput,
		},
		{
			args:       []string{"logging", "honeycomb", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListHoneycombsFn: listHoneycombsOK},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args:       []string{"logging", "honeycomb", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListHoneycombsFn: listHoneycombsOK},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args:       []string{"logging", "honeycomb", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHoneycombsFn: listHoneycombsOK},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "honeycomb", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHoneycombsFn: listHoneycombsOK},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args:      []string{"logging", "honeycomb", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListHoneycombsFn: listHoneycombsError},
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

func TestHoneycombDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "honeycomb", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "honeycomb", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetHoneycombFn: getHoneycombError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "honeycomb", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetHoneycombFn: getHoneycombOK},
			wantOutput: describeHoneycombOutput,
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

func TestHoneycombUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "honeycomb", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "honeycomb", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHoneycombFn:    getHoneycombError,
				UpdateHoneycombFn: updateHoneycombOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "honeycomb", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHoneycombFn:    getHoneycombOK,
				UpdateHoneycombFn: updateHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "honeycomb", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHoneycombFn:    getHoneycombOK,
				UpdateHoneycombFn: updateHoneycombOK,
			},
			wantOutput: "Updated Honeycomb logging endpoint log (service 123 version 1)",
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

func TestHoneycombDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "honeycomb", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "honeycomb", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteHoneycombFn: deleteHoneycombError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "honeycomb", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteHoneycombFn: deleteHoneycombOK},
			wantOutput: "Deleted Honeycomb logging endpoint logs (service 123 version 1)",
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

func createHoneycombOK(i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	s := fastly.Honeycomb{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createHoneycombError(i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func listHoneycombsOK(i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return []*fastly.Honeycomb{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			Dataset:           "log",
			Token:             "tkn",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Dataset:           "log",
			Token:             "tkn",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listHoneycombsError(i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return nil, errTest
}

var listHoneycombsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHoneycombsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Honeycomb 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Dataset: log
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Honeycomb 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Dataset: log
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getHoneycombOK(i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Dataset:           "log",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getHoneycombError(i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

var describeHoneycombOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Dataset: log
Token: tkn
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateHoneycombOK(i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Dataset:           "log",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateHoneycombError(i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func deleteHoneycombOK(i *fastly.DeleteHoneycombInput) error {
	return nil
}

func deleteHoneycombError(i *fastly.DeleteHoneycombInput) error {
	return errTest
}
