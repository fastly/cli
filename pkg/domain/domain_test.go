package domain_test

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

func TestDomainCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"domain", "create", "--version", "1", "--service-id", "123"},
			api:       mock.API{CreateDomainFn: createDomainOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       []string{"domain", "create", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:        mock.API{CreateDomainFn: createDomainOK},
			wantOutput: "Created domain www.test.com (service 123 version 1)",
		},
		{
			args:      []string{"domain", "create", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:       mock.API{CreateDomainFn: createDomainError},
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

func TestDomainList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"domain", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDomainsFn: listDomainsOK},
			wantOutput: listDomainsShortOutput,
		},
		{
			args:       []string{"domain", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListDomainsFn: listDomainsOK},
			wantOutput: listDomainsVerboseOutput,
		},
		{
			args:       []string{"domain", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListDomainsFn: listDomainsOK},
			wantOutput: listDomainsVerboseOutput,
		},
		{
			args:       []string{"domain", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDomainsFn: listDomainsOK},
			wantOutput: listDomainsVerboseOutput,
		},
		{
			args:       []string{"-v", "domain", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDomainsFn: listDomainsOK},
			wantOutput: listDomainsVerboseOutput,
		},
		{
			args:      []string{"domain", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListDomainsFn: listDomainsError},
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

func TestDomainDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"domain", "describe", "--service-id", "123", "--version", "1"},
			api:       mock.API{GetDomainFn: getDomainOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"domain", "describe", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:       mock.API{GetDomainFn: getDomainError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"domain", "describe", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:        mock.API{GetDomainFn: getDomainOK},
			wantOutput: describeDomainOutput,
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

func TestDomainUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"domain", "update", "--service-id", "123", "--version", "2", "--new-name", "www.test.com", "--comment", ""},
			api:       mock.API{UpdateDomainFn: updateDomainOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"domain", "update", "--service-id", "123", "--version", "2", "--name", "www.test.com"},
			api:       mock.API{UpdateDomainFn: updateDomainOK},
			wantError: "error parsing arguments: must provide either --new-name or --comment to update domain",
		},
		{
			args: []string{"domain", "update", "--service-id", "123", "--version", "2", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				GetDomainFn:    getDomainError,
				UpdateDomainFn: updateDomainOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"domain", "update", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				GetDomainFn:    getDomainError,
				UpdateDomainFn: updateDomainError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"domain", "update", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				GetDomainFn:    getDomainOK,
				UpdateDomainFn: updateDomainOK,
			},
			wantOutput: "Updated domain www.example.com (service 123 version 1)",
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

func TestDomainDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"domain", "delete", "--service-id", "123", "--version", "1"},
			api:       mock.API{DeleteDomainFn: deleteDomainOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"domain", "delete", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:       mock.API{DeleteDomainFn: deleteDomainError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"domain", "delete", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api:        mock.API{DeleteDomainFn: deleteDomainOK},
			wantOutput: "Deleted domain www.test.com (service 123 version 1)",
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

func createDomainOK(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Comment:        i.Comment,
	}, nil
}

func createDomainError(i *fastly.CreateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func listDomainsOK(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return []*fastly.Domain{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "www.test.com",
			Comment:        "test",
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "www.example.com",
			Comment:        "example",
		},
	}, nil
}

func listDomainsError(i *fastly.ListDomainsInput) ([]*fastly.Domain, error) {
	return nil, errTest
}

var listDomainsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME             COMMENT
123      1        www.test.com     test
123      1        www.example.com  example
`) + "\n"

var listDomainsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Domain 1/2
		Name: www.test.com
		Comment: test
	Domain 2/2
		Name: www.example.com
		Comment: example
`) + "\n\n"

func getDomainOK(i *fastly.GetDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Comment:        "test",
	}, nil
}

func getDomainError(i *fastly.GetDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

var describeDomainOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: www.test.com
Comment: test
`) + "\n"

func updateDomainOK(i *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return &fastly.Domain{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.NewName,
		Comment:        *i.Comment,
	}, nil
}

func updateDomainError(i *fastly.UpdateDomainInput) (*fastly.Domain, error) {
	return nil, errTest
}

func deleteDomainOK(i *fastly.DeleteDomainInput) error {
	return nil
}

func deleteDomainError(i *fastly.DeleteDomainInput) error {
	return errTest
}
