package papertrail_test

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

func TestPapertrailCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --address not provided",
		},
		{
			args:       []string{"logging", "papertrail", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com:123"},
			api:        mock.API{CreatePapertrailFn: createPapertrailOK},
			wantOutput: "Created Papertrail logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "papertrail", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com:123"},
			api:       mock.API{CreatePapertrailFn: createPapertrailError},
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

func TestPapertrailList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListPapertrailsFn: listPapertrailsOK},
			wantOutput: listPapertrailsShortOutput,
		},
		{
			args:       []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListPapertrailsFn: listPapertrailsOK},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args:       []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListPapertrailsFn: listPapertrailsOK},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args:       []string{"logging", "papertrail", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListPapertrailsFn: listPapertrailsOK},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "papertrail", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListPapertrailsFn: listPapertrailsOK},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args:      []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListPapertrailsFn: listPapertrailsError},
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

func TestPapertrailDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "papertrail", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetPapertrailFn: getPapertrailError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "papertrail", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetPapertrailFn: getPapertrailOK},
			wantOutput: describePapertrailOutput,
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

func TestPapertrailUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetPapertrailFn:    getPapertrailError,
				UpdatePapertrailFn: updatePapertrailOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetPapertrailFn:    getPapertrailOK,
				UpdatePapertrailFn: updatePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetPapertrailFn:    getPapertrailOK,
				UpdatePapertrailFn: updatePapertrailOK,
			},
			wantOutput: "Updated Papertrail logging endpoint log (service 123 version 1)",
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

func TestPapertrailDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "papertrail", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeletePapertrailFn: deletePapertrailError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "papertrail", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeletePapertrailFn: deletePapertrailOK},
			wantOutput: "Deleted Papertrail logging endpoint logs (service 123 version 1)",
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

func createPapertrailOK(i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createPapertrailError(i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func listPapertrailsOK(i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return []*fastly.Papertrail{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Address:           "example.com:123",
			Port:              123,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Address:           "127.0.0.1:456",
			Port:              456,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listPapertrailsError(i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return nil, errTest
}

var listPapertrailsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listPapertrailsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Papertrail 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Address: example.com:123
		Port: 123
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Papertrail 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: 127.0.0.1:456
		Port: 456
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getPapertrailOK(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com:123",
		Port:              123,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getPapertrailError(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

var describePapertrailOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Address: example.com:123
Port: 123
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updatePapertrailOK(i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Address:           "example.com:123",
		Port:              123,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updatePapertrailError(i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func deletePapertrailOK(i *fastly.DeletePapertrailInput) error {
	return nil
}

func deletePapertrailError(i *fastly.DeletePapertrailInput) error {
	return errTest
}
