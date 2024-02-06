package elasticsearch_test

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

func TestElasticsearchCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging elasticsearch create --service-id 123 --version 1 --name log --index logs --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				CreateElasticsearchFn: createElasticsearchOK,
			},
			wantOutput: "Created Elasticsearch logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging elasticsearch create --service-id 123 --version 1 --name log --index logs --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				CreateElasticsearchFn: createElasticsearchError,
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

func TestElasticsearchList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging elasticsearch list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsShortOutput,
		},
		{
			args: args("logging elasticsearch list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: args("logging elasticsearch list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: args("logging elasticsearch --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: args("logging -v elasticsearch list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: args("logging elasticsearch list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsError,
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

func TestElasticsearchDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging elasticsearch describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging elasticsearch describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				GetElasticsearchFn: getElasticsearchError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging elasticsearch describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				GetElasticsearchFn: getElasticsearchOK,
			},
			wantOutput: describeElasticsearchOutput,
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

func TestElasticsearchUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging elasticsearch update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging elasticsearch update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				UpdateElasticsearchFn: updateElasticsearchError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging elasticsearch update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				UpdateElasticsearchFn: updateElasticsearchOK,
			},
			wantOutput: "Updated Elasticsearch logging endpoint log (service 123 version 4)",
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

func TestElasticsearchDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging elasticsearch delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging elasticsearch delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				DeleteElasticsearchFn: deleteElasticsearchError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging elasticsearch delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				DeleteElasticsearchFn: deleteElasticsearchOK,
			},
			wantOutput: "Deleted Elasticsearch logging endpoint logs (service 123 version 4)",
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

func createElasticsearchOK(i *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
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

func createElasticsearchError(_ *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

func listElasticsearchsOK(i *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
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
		},
	}, nil
}

func listElasticsearchsError(_ *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
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
`) + "\n\n"

func getElasticsearchOK(i *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
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
	}, nil
}

func getElasticsearchError(_ *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
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

func updateElasticsearchOK(i *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
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

func updateElasticsearchError(_ *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

func deleteElasticsearchOK(_ *fastly.DeleteElasticsearchInput) error {
	return nil
}

func deleteElasticsearchError(_ *fastly.DeleteElasticsearchInput) error {
	return errTest
}
