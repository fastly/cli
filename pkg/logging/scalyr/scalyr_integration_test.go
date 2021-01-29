package scalyr_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	fsterrs "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestScalyrCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "scalyr", "create", "--service-id", "123", "--version", "1", "--auth-token", "abc"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "scalyr", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args:      []string{"logging", "scalyr", "create", "--name", "log", "--service-id", "", "--version", "1", "--auth-token", "abc"},
			wantError: fsterrs.ErrNoServiceID.Error(),
		},
		{
			args:       []string{"logging", "scalyr", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			api:        mock.API{CreateScalyrFn: createScalyrOK},
			wantOutput: "Created Scalyr logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "scalyr", "create", "--service-id", "123", "--version", "1", "--name", "log", "--auth-token", "abc"},
			api:       mock.API{CreateScalyrFn: createScalyrError},
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

func TestScalyrList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "scalyr", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListScalyrsFn: listScalyrsOK},
			wantOutput: listScalyrsShortOutput,
		},
		{
			args:       []string{"logging", "scalyr", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListScalyrsFn: listScalyrsOK},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args:       []string{"logging", "scalyr", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListScalyrsFn: listScalyrsOK},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args:       []string{"logging", "scalyr", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListScalyrsFn: listScalyrsOK},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "scalyr", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListScalyrsFn: listScalyrsOK},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args:      []string{"logging", "scalyr", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListScalyrsFn: listScalyrsError},
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

func TestScalyrDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "scalyr", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "scalyr", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetScalyrFn: getScalyrError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "scalyr", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetScalyrFn: getScalyrOK},
			wantOutput: describeScalyrOutput,
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

func TestScalyrUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "scalyr", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "scalyr", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetScalyrFn:    getScalyrError,
				UpdateScalyrFn: updateScalyrOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "scalyr", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetScalyrFn:    getScalyrOK,
				UpdateScalyrFn: updateScalyrError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "scalyr", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetScalyrFn:    getScalyrOK,
				UpdateScalyrFn: updateScalyrOK,
			},
			wantOutput: "Updated Scalyr logging endpoint log (service 123 version 1)",
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

func TestScalyrDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "scalyr", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "scalyr", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteScalyrFn: deleteScalyrError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "scalyr", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteScalyrFn: deleteScalyrOK},
			wantOutput: "Deleted Scalyr logging endpoint logs (service 123 version 1)",
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

func createScalyrOK(i *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	s := fastly.Scalyr{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	// Avoids null pointer dereference for test cases with missing required params.
	// If omitted, tests are guaranteed to panic.
	if i.Name != "" {
		s.Name = i.Name
	}

	if i.Token != "" {
		s.Token = i.Token
	}

	if i.Format != "" {
		s.Format = i.Format
	}

	if i.FormatVersion != 0 {
		s.FormatVersion = i.FormatVersion
	}

	if i.ResponseCondition != "" {
		s.ResponseCondition = i.ResponseCondition
	}

	if i.Placement != "" {
		s.Placement = i.Placement
	}

	return &s, nil
}

func createScalyrError(i *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

func listScalyrsOK(i *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
	return []*fastly.Scalyr{
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

func listScalyrsError(i *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
	return nil, errTest
}

var listScalyrsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listScalyrsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Scalyr 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Scalyr 2/2
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

func getScalyrOK(i *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return &fastly.Scalyr{
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

func getScalyrError(i *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

var describeScalyrOutput = strings.TrimSpace(`
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

func updateScalyrOK(i *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return &fastly.Scalyr{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Token:             "abc",
		Region:            "EU",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateScalyrError(i *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

func deleteScalyrOK(i *fastly.DeleteScalyrInput) error {
	return nil
}

func deleteScalyrError(i *fastly.DeleteScalyrInput) error {
	return errTest
}
