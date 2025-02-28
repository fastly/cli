package syslog_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestSyslogCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging syslog create --service-id 123 --version 1 --name log --address 127.0.0.1 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSyslogFn: createSyslogOK,
			},
			wantOutput: "Created Syslog logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging syslog create --service-id 123 --version 1 --name log --address 127.0.0.1 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSyslogFn: createSyslogError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestSyslogList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging syslog list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsShortOutput,
		},
		{
			args: args("logging syslog list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: args("logging syslog list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: args("logging syslog --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: args("logging -v syslog list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			wantOutput: listSyslogsVerboseOutput,
		},
		{
			args: args("logging syslog list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestSyslogDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging syslog describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging syslog describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSyslogFn:    getSyslogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging syslog describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSyslogFn:    getSyslogOK,
			},
			wantOutput: describeSyslogOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestSyslogUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging syslog update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging syslog update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSyslogFn: updateSyslogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging syslog update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSyslogFn: updateSyslogOK,
			},
			wantOutput: "Updated Syslog logging endpoint log (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestSyslogDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging syslog delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging syslog delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSyslogFn: deleteSyslogError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging syslog delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSyslogFn: deleteSyslogOK,
			},
			wantOutput: "Deleted Syslog logging endpoint logs (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createSyslogOK(i *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createSyslogError(_ *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

func listSyslogsOK(i *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
	return []*fastly.Syslog{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Address:           fastly.ToPointer("127.0.0.1"),
			Hostname:          fastly.ToPointer("127.0.0.1"),
			Port:              fastly.ToPointer(514),
			UseTLS:            fastly.ToPointer(false),
			IPV4:              fastly.ToPointer("127.0.0.1"),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       fastly.ToPointer("example.com"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			Token:             fastly.ToPointer("tkn"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Address:           fastly.ToPointer("example.com"),
			Hostname:          fastly.ToPointer("example.com"),
			Port:              fastly.ToPointer(789),
			UseTLS:            fastly.ToPointer(true),
			IPV4:              fastly.ToPointer("127.0.0.1"),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----baz"),
			TLSHostname:       fastly.ToPointer("example.com"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----qux"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----qux"),
			Token:             fastly.ToPointer("tkn"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
		},
	}, nil
}

func listSyslogsError(_ *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
	return nil, errTest
}

var listSyslogsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listSyslogsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Syslog 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Address: 127.0.0.1
		Hostname: 127.0.0.1
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
	Syslog 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: example.com
		Hostname: example.com
		Port: 789
		Use TLS: true
		IPV4: 127.0.0.1
		TLS CA certificate: -----BEGIN CERTIFICATE-----baz
		TLS hostname: example.com
		TLS client certificate: -----BEGIN CERTIFICATE-----qux
		TLS client key: -----BEGIN PRIVATE KEY-----qux
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Message type: classic
		Response condition: Prevent default logging
`) + "\n\n"

func getSyslogOK(i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Address:           fastly.ToPointer("example.com"),
		Hostname:          fastly.ToPointer("example.com"),
		Port:              fastly.ToPointer(514),
		UseTLS:            fastly.ToPointer(true),
		IPV4:              fastly.ToPointer(""),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("example.com"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		Token:             fastly.ToPointer("tkn"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		MessageType:       fastly.ToPointer("classic"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
	}, nil
}

func getSyslogError(_ *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

var describeSyslogOutput = `
Address: example.com
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Hostname: example.com
IPV4: ` + `
Message type: classic
Name: logs
Port: 514
Response condition: Prevent default logging
Service ID: 123
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
TLS hostname: example.com
Token: tkn
Use TLS: true
Version: 1
`

func updateSyslogOK(i *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Address:           fastly.ToPointer("example.com"),
		Hostname:          fastly.ToPointer("example.com"),
		Port:              fastly.ToPointer(514),
		UseTLS:            fastly.ToPointer(true),
		IPV4:              fastly.ToPointer(""),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("example.com"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		Token:             fastly.ToPointer("tkn"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		MessageType:       fastly.ToPointer("classic"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
	}, nil
}

func updateSyslogError(_ *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

func deleteSyslogOK(_ *fastly.DeleteSyslogInput) error {
	return nil
}

func deleteSyslogError(_ *fastly.DeleteSyslogInput) error {
	return errTest
}
