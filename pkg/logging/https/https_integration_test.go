package https_test

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

func TestHTTPSCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "https", "create", "--service-id", "123", "--version", "1", "--name", "log"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args:       []string{"logging", "https", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			api:        mock.API{CreateHTTPSFn: createHTTPSOK},
			wantOutput: "Created HTTPS logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "https", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com"},
			api:       mock.API{CreateHTTPSFn: createHTTPSError},
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

func TestHTTPSList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "https", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHTTPSFn: listHTTPSsOK},
			wantOutput: listHTTPSsShortOutput,
		},
		{
			args:       []string{"logging", "https", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListHTTPSFn: listHTTPSsOK},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args:       []string{"logging", "https", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListHTTPSFn: listHTTPSsOK},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args:       []string{"logging", "https", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHTTPSFn: listHTTPSsOK},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "https", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListHTTPSFn: listHTTPSsOK},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args:      []string{"logging", "https", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListHTTPSFn: listHTTPSsError},
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

func TestHTTPSDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "https", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "https", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetHTTPSFn: getHTTPSError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "https", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetHTTPSFn: getHTTPSOK},
			wantOutput: describeHTTPSOutput,
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

func TestHTTPSUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "https", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "https", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHTTPSFn:    getHTTPSError,
				UpdateHTTPSFn: updateHTTPSOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "https", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHTTPSFn:    getHTTPSOK,
				UpdateHTTPSFn: updateHTTPSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "https", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetHTTPSFn:    getHTTPSOK,
				UpdateHTTPSFn: updateHTTPSOK,
			},
			wantOutput: "Updated HTTPS logging endpoint log (service 123 version 1)",
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

func TestHTTPSDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "https", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "https", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteHTTPSFn: deleteHTTPSError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "https", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteHTTPSFn: deleteHTTPSOK},
			wantOutput: "Deleted HTTPS logging endpoint logs (service 123 version 1)",
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

func createHTTPSOK(i *fastly.CreateHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		URL:               "example.com",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		ContentType:       "application/json",
		HeaderName:        "name",
		HeaderValue:       "value",
		Method:            "GET",
		JSONFormat:        "1",
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		TLSHostname:       "example.com",
		MessageType:       "classic",
		FormatVersion:     2,
	}, nil
}

func createHTTPSError(i *fastly.CreateHTTPSInput) (*fastly.HTTPS, error) {
	return nil, errTest
}

func listHTTPSsOK(i *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error) {
	return []*fastly.HTTPS{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			URL:               "example.com",
			RequestMaxEntries: 2,
			RequestMaxBytes:   2,
			ContentType:       "application/json",
			HeaderName:        "name",
			HeaderValue:       "value",
			Method:            "GET",
			JSONFormat:        "1",
			Placement:         "none",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			TLSHostname:       "example.com",
			MessageType:       "classic",
			FormatVersion:     2,
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			URL:               "analytics.example.com",
			RequestMaxEntries: 2,
			RequestMaxBytes:   2,
			ContentType:       "application/json",
			HeaderName:        "name",
			HeaderValue:       "value",
			Method:            "GET",
			JSONFormat:        "1",
			Placement:         "none",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			TLSHostname:       "example.com",
			MessageType:       "classic",
			FormatVersion:     2,
		},
	}, nil
}

func listHTTPSsError(i *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error) {
	return nil, errTest
}

var listHTTPSsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHTTPSsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	HTTPS 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Content type: application/json
		Header name: name
		Header value: value
		Method: GET
		JSON format: 1
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		TLS hostname: example.com
		Request max entries: 2
		Request max bytes: 2
		Message type: classic
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	HTTPS 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: analytics.example.com
		Content type: application/json
		Header name: name
		Header value: value
		Method: GET
		JSON format: 1
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		TLS hostname: example.com
		Request max entries: 2
		Request max bytes: 2
		Message type: classic
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getHTTPSOK(i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		URL:               "example.com",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		ContentType:       "application/json",
		HeaderName:        "name",
		HeaderValue:       "value",
		Method:            "GET",
		JSONFormat:        "1",
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		TLSHostname:       "example.com",
		MessageType:       "classic",
		FormatVersion:     2,
	}, nil
}

func getHTTPSError(i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return nil, errTest
}

var describeHTTPSOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: log
URL: example.com
Content type: application/json
Header name: name
Header value: value
Method: GET
JSON format: 1
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
TLS hostname: example.com
Request max entries: 2
Request max bytes: 2
Message type: classic
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateHTTPSOK(i *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		URL:               "example.com",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		ContentType:       "application/json",
		HeaderName:        "name",
		HeaderValue:       "value",
		Method:            "GET",
		JSONFormat:        "1",
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		TLSHostname:       "example.com",
		MessageType:       "classic",
		FormatVersion:     2,
	}, nil
}

func updateHTTPSError(i *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error) {
	return nil, errTest
}

func deleteHTTPSOK(i *fastly.DeleteHTTPSInput) error {
	return nil
}

func deleteHTTPSError(i *fastly.DeleteHTTPSInput) error {
	return errTest
}
