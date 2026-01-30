package elasticsearch_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/elasticsearch"
)

func TestElasticsearchCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --index logs --url example.com --autoclone",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				CreateElasticsearchFn: createElasticsearchOK,
			},
			WantOutput: "Created Elasticsearch logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --index logs --url example.com --autoclone",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				CreateElasticsearchFn: createElasticsearchError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestElasticsearchList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			WantOutput: listElasticsearchsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			WantOutput: listElasticsearchsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			WantOutput: listElasticsearchsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestElasticsearchDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				GetElasticsearchFn: getElasticsearchError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				GetElasticsearchFn: getElasticsearchOK,
			},
			WantOutput: describeElasticsearchOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestElasticsearchUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				UpdateElasticsearchFn: updateElasticsearchError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				UpdateElasticsearchFn: updateElasticsearchOK,
			},
			WantOutput: "Updated Elasticsearch logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestElasticsearchDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				DeleteElasticsearchFn: deleteElasticsearchError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				DeleteElasticsearchFn: deleteElasticsearchOK,
			},
			WantOutput: "Deleted Elasticsearch logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createElasticsearchOK(_ context.Context, i *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return &fastly.Elasticsearch{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Index:             fastly.ToPointer("logs"),
		URL:               fastly.ToPointer("example.com"),
		Pipeline:          fastly.ToPointer("logs"),
		User:              fastly.ToPointer("user"),
		Password:          fastly.ToPointer("password"),
		RequestMaxEntries: fastly.ToPointer(2),
		RequestMaxBytes:   fastly.ToPointer(2),
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("example.com"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     fastly.ToPointer(2),
	}, nil
}

func createElasticsearchError(_ context.Context, _ *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

func listElasticsearchsOK(_ context.Context, i *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
	return []*fastly.Elasticsearch{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			Index:             fastly.ToPointer("logs"),
			URL:               fastly.ToPointer("example.com"),
			Pipeline:          fastly.ToPointer("logs"),
			User:              fastly.ToPointer("user"),
			Password:          fastly.ToPointer("password"),
			RequestMaxEntries: fastly.ToPointer(2),
			RequestMaxBytes:   fastly.ToPointer(2),
			Placement:         fastly.ToPointer("none"),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       fastly.ToPointer("example.com"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			FormatVersion:     fastly.ToPointer(2),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Index:             fastly.ToPointer("analytics"),
			URL:               fastly.ToPointer("example.com"),
			Pipeline:          fastly.ToPointer("analytics"),
			User:              fastly.ToPointer("user"),
			Password:          fastly.ToPointer("password"),
			RequestMaxEntries: fastly.ToPointer(2),
			RequestMaxBytes:   fastly.ToPointer(2),
			Placement:         fastly.ToPointer("none"),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       fastly.ToPointer("example.com"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listElasticsearchsError(_ context.Context, _ *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
	return nil, errTest
}

var listElasticsearchsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listElasticsearchsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Elasticsearch 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Index: logs
		URL: example.com
		Pipeline: logs
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		TLS hostname: example.com
		User: user
		Password: password
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
	Elasticsearch 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Index: analytics
		URL: example.com
		Pipeline: analytics
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		TLS hostname: example.com
		User: user
		Password: password
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
`) + "\n\n"

func getElasticsearchOK(_ context.Context, i *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
	return &fastly.Elasticsearch{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Index:             fastly.ToPointer("logs"),
		URL:               fastly.ToPointer("example.com"),
		Pipeline:          fastly.ToPointer("logs"),
		User:              fastly.ToPointer("user"),
		Password:          fastly.ToPointer("password"),
		RequestMaxEntries: fastly.ToPointer(2),
		RequestMaxBytes:   fastly.ToPointer(2),
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("example.com"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     fastly.ToPointer(2),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getElasticsearchError(_ context.Context, _ *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

var describeElasticsearchOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Index: logs
Name: logs
Password: password
Pipeline: logs
Placement: none
Processing region: us
Response condition: Prevent default logging
Service ID: 123
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
TLS hostname: example.com
URL: example.com
User: user
Version: 1
`) + "\n"

func updateElasticsearchOK(_ context.Context, i *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return &fastly.Elasticsearch{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Index:             fastly.ToPointer("logs"),
		URL:               fastly.ToPointer("example.com"),
		Pipeline:          fastly.ToPointer("logs"),
		User:              fastly.ToPointer("user"),
		Password:          fastly.ToPointer("password"),
		RequestMaxEntries: fastly.ToPointer(2),
		RequestMaxBytes:   fastly.ToPointer(2),
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("example.com"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     fastly.ToPointer(2),
	}, nil
}

func updateElasticsearchError(_ context.Context, _ *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

func deleteElasticsearchOK(_ context.Context, _ *fastly.DeleteElasticsearchInput) error {
	return nil
}

func deleteElasticsearchError(_ context.Context, _ *fastly.DeleteElasticsearchInput) error {
	return errTest
}
