package sumologic_test

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

func TestSumologicCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args:       []string{"logging", "sumologic", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			api:        mock.API{CreateSumologicFn: createSumologicOK},
			wantOutput: "Created Sumologic logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "sumologic", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			api:       mock.API{CreateSumologicFn: createSumologicError},
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

func TestSumologicList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSumologicsFn: listSumologicsOK},
			wantOutput: listSumologicsShortOutput,
		},
		{
			args:       []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListSumologicsFn: listSumologicsOK},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args:       []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListSumologicsFn: listSumologicsOK},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args:       []string{"logging", "sumologic", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSumologicsFn: listSumologicsOK},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "sumologic", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSumologicsFn: listSumologicsOK},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args:      []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListSumologicsFn: listSumologicsError},
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

func TestSumologicDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "sumologic", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetSumologicFn: getSumologicError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "sumologic", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetSumologicFn: getSumologicOK},
			wantOutput: describeSumologicOutput,
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

func TestSumologicUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSumologicFn:    getSumologicError,
				UpdateSumologicFn: updateSumologicOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSumologicFn:    getSumologicOK,
				UpdateSumologicFn: updateSumologicError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSumologicFn:    getSumologicOK,
				UpdateSumologicFn: updateSumologicOK,
			},
			wantOutput: "Updated Sumologic logging endpoint log (service 123 version 1)",
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

func TestSumologicDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "sumologic", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteSumologicFn: deleteSumologicError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "sumologic", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteSumologicFn: deleteSumologicOK},
			wantOutput: "Deleted Sumologic logging endpoint logs (service 123 version 1)",
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

func createSumologicOK(i *fastly.CreateSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createSumologicError(i *fastly.CreateSumologicInput) (*fastly.Sumologic, error) {
	return nil, errTest
}

func listSumologicsOK(i *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error) {
	return []*fastly.Sumologic{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			URL:               "example.com",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			URL:               "bar.com",
			Format:            `%h %l %u %t "%r" %>s %b`,
			ResponseCondition: "Prevent default logging",
			MessageType:       "classic",
			FormatVersion:     2,
			Placement:         "none",
		},
	}, nil
}

func listSumologicsError(i *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error) {
	return nil, errTest
}

var listSumologicsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listSumologicsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Sumologic 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
	Sumologic 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: bar.com
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
`) + "\n\n"

func getSumologicOK(i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		URL:               "example.com",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getSumologicError(i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return nil, errTest
}

var describeSumologicOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
URL: example.com
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Placement: none
`) + "\n"

func updateSumologicOK(i *fastly.UpdateSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		URL:               "example.com",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateSumologicError(i *fastly.UpdateSumologicInput) (*fastly.Sumologic, error) {
	return nil, errTest
}

func deleteSumologicOK(i *fastly.DeleteSumologicInput) error {
	return nil
}

func deleteSumologicError(i *fastly.DeleteSumologicInput) error {
	return errTest
}
