package syslog_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/syslog"
)

func TestSyslogCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --address 127.0.0.1 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSyslogFn: createSyslogOK,
			},
			WantOutput: "Created Syslog logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --address 127.0.0.1 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSyslogFn: createSyslogError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestSyslogList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			WantOutput: listSyslogsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			WantOutput: listSyslogsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsOK,
			},
			WantOutput: listSyslogsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSyslogsFn:  listSyslogsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestSyslogDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSyslogFn:    getSyslogError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSyslogFn:    getSyslogOK,
			},
			WantOutput: describeSyslogOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestSyslogUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSyslogFn: updateSyslogError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSyslogFn: updateSyslogOK,
			},
			WantOutput: "Updated Syslog logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestSyslogDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSyslogFn: deleteSyslogError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSyslogFn: deleteSyslogOK,
			},
			WantOutput: "Deleted Syslog logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createSyslogOK(_ context.Context, i *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createSyslogError(_ context.Context, _ *fastly.CreateSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

func listSyslogsOK(_ context.Context, i *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
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
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
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
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listSyslogsError(_ context.Context, _ *fastly.ListSyslogsInput) ([]*fastly.Syslog, error) {
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
		Placement: none
		Processing region: us
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
		Placement: none
		Processing region: us
`) + "\n\n"

func getSyslogOK(_ context.Context, i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
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
		Placement:         fastly.ToPointer("none"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getSyslogError(_ context.Context, _ *fastly.GetSyslogInput) (*fastly.Syslog, error) {
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
Placement: none
Port: 514
Processing region: us
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

func updateSyslogOK(_ context.Context, i *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
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
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func updateSyslogError(_ context.Context, _ *fastly.UpdateSyslogInput) (*fastly.Syslog, error) {
	return nil, errTest
}

func deleteSyslogOK(_ context.Context, _ *fastly.DeleteSyslogInput) error {
	return nil
}

func deleteSyslogError(_ context.Context, _ *fastly.DeleteSyslogInput) error {
	return errTest
}
