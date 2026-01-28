package https_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHTTPSCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("service logging https create --service-id 123 --version 1 --name log --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateHTTPSFn:  createHTTPSOK,
			},
			wantOutput: "Created HTTPS logging endpoint log (service 123 version 4)",
		},
		{
			args: args("service logging https create --service-id 123 --version 1 --name log --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateHTTPSFn:  createHTTPSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("service logging https create --service-id 123 --version 1 --name log --url example.com --compression-codec zstd --gzip-level 9 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
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

func TestHTTPSList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("service logging https list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			wantOutput: listHTTPSsShortOutput,
		},
		{
			args: args("service logging https list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args: args("service logging https list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args: args("service logging https --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args: args("service logging -v https list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsOK,
			},
			wantOutput: listHTTPSsVerboseOutput,
		},
		{
			args: args("service logging https list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHTTPSFn:    listHTTPSsError,
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

func TestHTTPSDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service logging https describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("service logging https describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHTTPSFn:     getHTTPSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("service logging https describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHTTPSFn:     getHTTPSOK,
			},
			wantOutput: describeHTTPSOutput,
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

func TestHTTPSUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service logging https update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("service logging https update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateHTTPSFn:  updateHTTPSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("service logging https update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateHTTPSFn:  updateHTTPSOK,
			},
			wantOutput: "Updated HTTPS logging endpoint log (service 123 version 4)",
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

func TestHTTPSDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service logging https delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("service logging https delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteHTTPSFn:  deleteHTTPSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("service logging https delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteHTTPSFn:  deleteHTTPSOK,
			},
			wantOutput: "Deleted HTTPS logging endpoint logs (service 123 version 4)",
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
