package syslog_test

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

func TestSyslogCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "syslog", "create", "--service-id", "123", "--version", "2", "--name", "log"},
			wantError: "error parsing arguments: required flag --address not provided",
		},
		{
			args: []string{"logging", "syslog", "create", "--service-id", "123", "--version", "2", "--name", "log", "--address", "127.0.0.1", "--autoclone"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				CreateSyslogFn: createSyslogOK,
			},
			wantOutput: "Created Syslog logging endpoint log (service 123 version 3)",
		},
		{
			args: []string{"logging", "syslog", "create", "--service-id", "123", "--version", "2", "--name", "log", "--address", "127.0.0.1"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				CreateSyslogFn: createSyslogError,
			},
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestSyslogList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "syslog", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsShortOutput,
		},
		{
			args: []string{"logging", "syslog", "list", "--service-id", "123", "--version", "2", "--verbose"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: []string{"logging", "syslog", "list", "--service-id", "123", "--version", "2", "-v"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: []string{"logging", "syslog", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "syslog", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: []string{"logging", "syslog", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				ListSyslogsFn:  listSyslogsError,
			},
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestSyslogDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "syslog", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "syslog", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				GetSyslogFn:    getSyslogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "syslog", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				GetSyslogFn:    getSyslogOK,
			},
			wantOutput: describeSyslogOutput,
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestSyslogUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "syslog", "update", "--service-id", "123", "--version", "2", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "syslog", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				UpdateSyslogFn: updateSyslogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "syslog", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				UpdateSyslogFn: updateSyslogOK,
			},
			wantOutput: "Updated Syslog logging endpoint log (service 123 version 3)",
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestSyslogDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "syslog", "delete", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "syslog", "delete", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				DeleteSyslogFn: deleteSyslogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "syslog", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn: listVersionsActiveOk,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				DeleteSyslogFn: deleteSyslogOK,
			},
			wantOutput: "Deleted Syslog logging endpoint logs (service 123 version 3)",
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createSyslogOK(i *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createSyslogError(i *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

func listSyslogsOK(i *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
	return []*fastly.Syslog{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Address:           "127.0.0.1",
			Hostname:          "",
			Port:              514,
			UseTLS:            false,
			IPV4:              "127.0.0.1",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "example.com",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			Token:             "tkn",
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
			Address:           "example.com",
			Hostname:          "example.com",
			Port:              789,
			UseTLS:            true,
			IPV4:              "",
			TLSCACert:         "-----BEGIN CERTIFICATE-----baz",
			TLSHostname:       "example.com",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----qux",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----qux",
			Token:             "tkn",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listSyslogsError(i *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
	return nil, errTest
}

var listSyslogsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      2        logs
123      2        analytics
`) + "\n"

var listSyslogsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 2
	Syslog 1/2
		Service ID: 123
		Version: 2
		Name: logs
		Address: 127.0.0.1
		Hostname: 
		Port: 514
		Use TLS: false
		IPV4: 127.0.0.1
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS hostname: example.com
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Message type: classic
		Response condition: Prevent default logging
		Placement: none
	Syslog 2/2
		Service ID: 123
		Version: 2
		Name: analytics
		Address: example.com
		Hostname: example.com
		Port: 789
		Use TLS: true
		IPV4: 
		TLS CA certificate: -----BEGIN CERTIFICATE-----baz
		TLS hostname: example.com
		TLS client certificate: -----BEGIN CERTIFICATE-----qux
		TLS client key: -----BEGIN PRIVATE KEY-----qux
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Message type: classic
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getSyslogOK(i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com",
		Hostname:          "example.com",
		Port:              514,
		UseTLS:            true,
		IPV4:              "",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getSyslogError(i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

var describeSyslogOutput = strings.TrimSpace(`
Service ID: 123
Version: 2
Name: logs
Address: example.com
Hostname: example.com
Port: 514
Use TLS: true
IPV4: 
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS hostname: example.com
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
Token: tkn
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Message type: classic
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateSyslogOK(i *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Address:           "example.com",
		Hostname:          "example.com",
		Port:              514,
		UseTLS:            true,
		IPV4:              "",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateSyslogError(i *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

func deleteSyslogOK(i *fastly.DeleteSyslogInput) error {
	return nil
}

func deleteSyslogError(i *fastly.DeleteSyslogInput) error {
	return errTest
}

func listVersionsActiveOk(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func getVersionOK(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		ServiceID: i.ServiceID,
		Number:    2,
		Active:    true,
		UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}, nil
}

func cloneVersionOK(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion + 1}, nil
}
