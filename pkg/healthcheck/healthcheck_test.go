package healthcheck_test

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

func TestHealthCheckCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"healthcheck", "create", "--version", "1", "--service-id", "123"},
			api:       mock.API{CreateHealthCheckFn: createHealthCheckOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"healthcheck", "create", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:       mock.API{CreateHealthCheckFn: createHealthCheckError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"healthcheck", "create", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:        mock.API{CreateHealthCheckFn: createHealthCheckOK},
			wantOutput: "Created healthcheck www.test.com (service 123 version 1)",
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

func TestHealthCheckList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"healthcheck", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHealthChecksFn: listHealthChecksOK},
			wantOutput: listHealthChecksShortOutput,
		},
		{
			args:       []string{"healthcheck", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListHealthChecksFn: listHealthChecksOK},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args:       []string{"healthcheck", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListHealthChecksFn: listHealthChecksOK},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args:       []string{"healthcheck", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHealthChecksFn: listHealthChecksOK},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args:       []string{"-v", "healthcheck", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHealthChecksFn: listHealthChecksOK},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args:      []string{"healthcheck", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListHealthChecksFn: listHealthChecksError},
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

func TestHealthCheckDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"healthcheck", "describe", "--service-id", "123", "--version", "1"},
			api:       mock.API{GetHealthCheckFn: getHealthCheckOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"healthcheck", "describe", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:       mock.API{GetHealthCheckFn: getHealthCheckError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"healthcheck", "describe", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:        mock.API{GetHealthCheckFn: getHealthCheckOK},
			wantOutput: describeHealthCheckOutput,
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

func TestHealthCheckUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"healthcheck", "update", "--service-id", "123", "--version", "2", "--new-name", "www.test.com", "--comment", ""},
			api:       mock.API{UpdateHealthCheckFn: updateHealthCheckOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"healthcheck", "update", "--service-id", "123", "--version", "2", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				GetHealthCheckFn:    getHealthCheckError,
				UpdateHealthCheckFn: updateHealthCheckOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"healthcheck", "update", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				GetHealthCheckFn:    getHealthCheckError,
				UpdateHealthCheckFn: updateHealthCheckError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"healthcheck", "update", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				GetHealthCheckFn:    getHealthCheckOK,
				UpdateHealthCheckFn: updateHealthCheckOK,
			},
			wantOutput: "Updated healthcheck www.example.com (service 123 version 1)",
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

func TestHealthCheckDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"healthcheck", "delete", "--service-id", "123", "--version", "1"},
			api:       mock.API{DeleteHealthCheckFn: deleteHealthCheckOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"healthcheck", "delete", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:       mock.API{DeleteHealthCheckFn: deleteHealthCheckError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"healthcheck", "delete", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:        mock.API{DeleteHealthCheckFn: deleteHealthCheckOK},
			wantOutput: "Deleted healthcheck www.test.com (service 123 version 1)",
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

func createHealthCheckOK(i *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Comment:        i.Comment,
		Host:           "www.test.com",
		Path:           "/health",
	}, nil
}

func createHealthCheckError(i *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

func listHealthChecksOK(i *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return []*fastly.HealthCheck{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "test",
			Comment:        "test",
			Method:         "HEAD",
			Host:           "www.test.com",
			Path:           "/health",
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "example",
			Comment:        "example",
			Method:         "HEAD",
			Host:           "www.example.com",
			Path:           "/health",
		},
	}, nil
}

func listHealthChecksError(i *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return nil, errTest
}

var listHealthChecksShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME     METHOD  HOST             PATH
123      1        test     HEAD    www.test.com     /health
123      1        example  HEAD    www.example.com  /health
`) + "\n"

var listHealthChecksVerboseOutput = strings.Join([]string{
	"Fastly API token not provided",
	"Fastly API endpoint: https://api.fastly.com",
	"Service ID: 123",
	"Version: 1",
	"	Healthcheck 1/2",
	"		Name: test",
	"		Comment: test",
	"		Method: HEAD",
	"		Host: www.test.com",
	"		Path: /health",
	"		HTTP version: ",
	"		Timeout: 0",
	"		Check interval: 0",
	"		Expected response: 0",
	"		Window: 0",
	"		Threshold: 0",
	"		Initial: 0",
	"	Healthcheck 2/2",
	"		Name: example",
	"		Comment: example",
	"		Method: HEAD",
	"		Host: www.example.com",
	"		Path: /health",
	"		HTTP version: ",
	"		Timeout: 0",
	"		Check interval: 0",
	"		Expected response: 0",
	"		Window: 0",
	"		Threshold: 0",
	"		Initial: 0",
}, "\n") + "\n\n"

func getHealthCheckOK(i *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           "test",
		Method:         "HEAD",
		Host:           "www.test.com",
		Path:           "/healthcheck",
		Comment:        "test",
	}, nil
}

func getHealthCheckError(i *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

var describeHealthCheckOutput = strings.Join([]string{
	"Service ID: 123",
	"Version: 1",
	"Name: test",
	"Comment: test",
	"Method: HEAD",
	"Host: www.test.com",
	"Path: /healthcheck",
	"HTTP version: ",
	"Timeout: 0",
	"Check interval: 0",
	"Expected response: 0",
	"Window: 0",
	"Threshold: 0",
	"Initial: 0",
}, "\n") + "\n"

func updateHealthCheckOK(i *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.NewName,
		Comment:        *i.Comment,
	}, nil
}

func updateHealthCheckError(i *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

func deleteHealthCheckOK(i *fastly.DeleteHealthCheckInput) error {
	return nil
}

func deleteHealthCheckError(i *fastly.DeleteHealthCheckInput) error {
	return errTest
}
