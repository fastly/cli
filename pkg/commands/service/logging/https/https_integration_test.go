package https_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/https"
)

func TestHTTPSCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --url example.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateHTTPSFn:  createHTTPSOK,
			},
			WantOutput: "Created HTTPS logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --url example.com --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateHTTPSFn:  createHTTPSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name log --url example.com --compression-codec zstd --gzip-level 9 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestHTTPSList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			WantOutput: listHTTPSsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			WantOutput: listHTTPSsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			WantOutput: listHTTPSsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestHTTPSDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHTTPSFn:     getHTTPSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHTTPSFn:     getHTTPSOK,
			},
			WantOutput: describeHTTPSOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestHTTPSUpdate(t *testing.T) {
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
				UpdateHTTPSFn:  updateHTTPSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateHTTPSFn:  updateHTTPSOK,
			},
			WantOutput: "Updated HTTPS logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestHTTPSDelete(t *testing.T) {
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
				DeleteHTTPSFn:  deleteHTTPSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteHTTPSFn:  deleteHTTPSOK,
			},
			WantOutput: "Deleted HTTPS logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createHTTPSOK(_ context.Context, i *fastly.CreateHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		URL:               fastly.ToPointer("example.com"),
		RequestMaxEntries: fastly.ToPointer(2),
		RequestMaxBytes:   fastly.ToPointer(2),
		CompressionCodec:  fastly.ToPointer(""),
		ContentType:       fastly.ToPointer("application/json"),
		GzipLevel:         fastly.ToPointer(0),
		HeaderName:        fastly.ToPointer("name"),
		HeaderValue:       fastly.ToPointer("value"),
		Method:            fastly.ToPointer(http.MethodGet),
		JSONFormat:        fastly.ToPointer("1"),
		Period:            fastly.ToPointer(0),
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		TLSHostname:       fastly.ToPointer("example.com"),
		MessageType:       fastly.ToPointer("classic"),
		FormatVersion:     fastly.ToPointer(2),
	}, nil
}

func createHTTPSError(_ context.Context, _ *fastly.CreateHTTPSInput) (*fastly.HTTPS, error) {
	return nil, errTest
}

func listHTTPSsOK(_ context.Context, i *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error) {
	return []*fastly.HTTPS{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			URL:               fastly.ToPointer("example.com"),
			RequestMaxEntries: fastly.ToPointer(2),
			RequestMaxBytes:   fastly.ToPointer(2),
			CompressionCodec:  fastly.ToPointer(""),
			ContentType:       fastly.ToPointer("application/json"),
			GzipLevel:         fastly.ToPointer(0),
			HeaderName:        fastly.ToPointer("name"),
			HeaderValue:       fastly.ToPointer("value"),
			Method:            fastly.ToPointer(http.MethodGet),
			JSONFormat:        fastly.ToPointer("1"),
			Period:            fastly.ToPointer(0),
			Placement:         fastly.ToPointer("none"),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			TLSHostname:       fastly.ToPointer("example.com"),
			MessageType:       fastly.ToPointer("classic"),
			FormatVersion:     fastly.ToPointer(2),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			URL:               fastly.ToPointer("analytics.example.com"),
			RequestMaxEntries: fastly.ToPointer(2),
			RequestMaxBytes:   fastly.ToPointer(2),
			CompressionCodec:  fastly.ToPointer(""),
			ContentType:       fastly.ToPointer("application/json"),
			GzipLevel:         fastly.ToPointer(0),
			HeaderName:        fastly.ToPointer("name"),
			HeaderValue:       fastly.ToPointer("value"),
			Method:            fastly.ToPointer(http.MethodGet),
			JSONFormat:        fastly.ToPointer("1"),
			Period:            fastly.ToPointer(0),
			Placement:         fastly.ToPointer("none"),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			TLSHostname:       fastly.ToPointer("example.com"),
			MessageType:       fastly.ToPointer("classic"),
			FormatVersion:     fastly.ToPointer(2),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listHTTPSsError(_ context.Context, _ *fastly.ListHTTPSInput) ([]*fastly.HTTPS, error) {
	return nil, errTest
}

var listHTTPSsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHTTPSsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	HTTPS 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Compression codec: 
		Content type: application/json
		GZip level: 0
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
		Period: 0
		Placement: none
		Processing region: us
	HTTPS 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: analytics.example.com
		Compression codec: 
		Content type: application/json
		GZip level: 0
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
		Period: 0
		Placement: none
		Processing region: us
`) + "\n\n"

func getHTTPSOK(_ context.Context, i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		URL:               fastly.ToPointer("example.com"),
		RequestMaxEntries: fastly.ToPointer(2),
		RequestMaxBytes:   fastly.ToPointer(2),
		CompressionCodec:  fastly.ToPointer(""),
		ContentType:       fastly.ToPointer("application/json"),
		GzipLevel:         fastly.ToPointer(0),
		HeaderName:        fastly.ToPointer("name"),
		HeaderValue:       fastly.ToPointer("value"),
		Method:            fastly.ToPointer(http.MethodGet),
		JSONFormat:        fastly.ToPointer("1"),
		Period:            fastly.ToPointer(0),
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		TLSHostname:       fastly.ToPointer("example.com"),
		MessageType:       fastly.ToPointer("classic"),
		FormatVersion:     fastly.ToPointer(2),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getHTTPSError(_ context.Context, _ *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return nil, errTest
}

var describeHTTPSOutput = "\n" + strings.TrimSpace(`
Compression codec: 
Content type: application/json
Format: %h %l %u %t "%r" %>s %b
Format version: 2
GZip level: 0
Header name: name
Header value: value
JSON format: 1
Message type: classic
Method: GET
Name: log
Period: 0
Placement: none
Processing region: us
Request max bytes: 2
Request max entries: 2
Response condition: Prevent default logging
Service ID: 123
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
TLS hostname: example.com
URL: example.com
Version: 1
`) + "\n"

func updateHTTPSOK(_ context.Context, i *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		URL:               fastly.ToPointer("example.com"),
		RequestMaxEntries: fastly.ToPointer(2),
		RequestMaxBytes:   fastly.ToPointer(2),
		CompressionCodec:  fastly.ToPointer(""),
		ContentType:       fastly.ToPointer("application/json"),
		GzipLevel:         fastly.ToPointer(7),
		HeaderName:        fastly.ToPointer("name"),
		HeaderValue:       fastly.ToPointer("value"),
		Method:            fastly.ToPointer(http.MethodGet),
		JSONFormat:        fastly.ToPointer("1"),
		Period:            fastly.ToPointer(0),
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		TLSHostname:       fastly.ToPointer("example.com"),
		MessageType:       fastly.ToPointer("classic"),
		FormatVersion:     fastly.ToPointer(2),
	}, nil
}

func updateHTTPSError(_ context.Context, _ *fastly.UpdateHTTPSInput) (*fastly.HTTPS, error) {
	return nil, errTest
}

func deleteHTTPSOK(_ context.Context, _ *fastly.DeleteHTTPSInput) error {
	return nil
}

func deleteHTTPSError(_ context.Context, _ *fastly.DeleteHTTPSInput) error {
	return errTest
}
