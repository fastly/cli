package splunk_test

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

func TestSplunkCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "splunk", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args:       []string{"logging", "splunk", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			api:        mock.API{CreateSplunkFn: createSplunkOK},
			wantOutput: "Created Splunk logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "splunk", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			api:       mock.API{CreateSplunkFn: createSplunkError},
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

func TestSplunkList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "splunk", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSplunksFn: listSplunksOK},
			wantOutput: listSplunksShortOutput,
		},
		{
			args:       []string{"logging", "splunk", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListSplunksFn: listSplunksOK},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args:       []string{"logging", "splunk", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListSplunksFn: listSplunksOK},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args:       []string{"logging", "splunk", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSplunksFn: listSplunksOK},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "splunk", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSplunksFn: listSplunksOK},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args:      []string{"logging", "splunk", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListSplunksFn: listSplunksError},
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

func TestSplunkDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "splunk", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "splunk", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetSplunkFn: getSplunkError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "splunk", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetSplunkFn: getSplunkOK},
			wantOutput: describeSplunkOutput,
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

func TestSplunkUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "splunk", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "splunk", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSplunkFn:    getSplunkError,
				UpdateSplunkFn: updateSplunkOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "splunk", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSplunkFn:    getSplunkOK,
				UpdateSplunkFn: updateSplunkError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "splunk", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSplunkFn:    getSplunkOK,
				UpdateSplunkFn: updateSplunkOK,
			},
			wantOutput: "Updated Splunk logging endpoint log (service 123 version 1)",
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

func TestSplunkDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "splunk", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "splunk", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteSplunkFn: deleteSplunkError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "splunk", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteSplunkFn: deleteSplunkOK},
			wantOutput: "Deleted Splunk logging endpoint logs (service 123 version 1)",
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

func createSplunkOK(i *fastly.CreateSplunkInput) (*fastly.Splunk, error) {
	return &fastly.Splunk{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createSplunkError(i *fastly.CreateSplunkInput) (*fastly.Splunk, error) {
	return nil, errTest
}

func listSplunksOK(i *fastly.ListSplunksInput) ([]*fastly.Splunk, error) {
	return []*fastly.Splunk{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			URL:               "example.com",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
			Token:             "tkn",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "example.com",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			URL:               "127.0.0.1",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
			Token:             "tkn1",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "example.com",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----qux",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----qux",
		},
	}, nil
}

func listSplunksError(i *fastly.ListSplunksInput) ([]*fastly.Splunk, error) {
	return nil, errTest
}

var listSplunksShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listSplunksVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Splunk 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Token: tkn
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS hostname: example.com
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Splunk 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: 127.0.0.1
		Token: tkn1
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS hostname: example.com
		TLS client certificate: -----BEGIN CERTIFICATE-----qux
		TLS client key: -----BEGIN PRIVATE KEY-----qux
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getSplunkOK(i *fastly.GetSplunkInput) (*fastly.Splunk, error) {
	return &fastly.Splunk{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		URL:               "example.com",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
		Token:             "tkn",
	}, nil
}

func getSplunkError(i *fastly.GetSplunkInput) (*fastly.Splunk, error) {
	return nil, errTest
}

var describeSplunkOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
URL: example.com
Token: tkn
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS hostname: example.com
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateSplunkOK(i *fastly.UpdateSplunkInput) (*fastly.Splunk, error) {
	return &fastly.Splunk{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		URL:               "example.com",
		Token:             "tkn",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateSplunkError(i *fastly.UpdateSplunkInput) (*fastly.Splunk, error) {
	return nil, errTest
}

func deleteSplunkOK(i *fastly.DeleteSplunkInput) error {
	return nil
}

func deleteSplunkError(i *fastly.DeleteSplunkInput) error {
	return errTest
}
