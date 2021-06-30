package elasticsearch_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestElasticsearchCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "elasticsearch", "create", "--service-id", "123", "--version", "1", "--name", "log", "--index", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args: []string{"logging", "elasticsearch", "create", "--service-id", "123", "--version", "1", "--name", "log", "--url", "example.com", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --index not provided",
		},
		{
			args: []string{"logging", "elasticsearch", "create", "--service-id", "123", "--version", "1", "--name", "log", "--index", "logs", "--url", "example.com", "--autoclone"},
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				CreateElasticsearchFn: createElasticsearchOK,
			},
			wantOutput: "Created Elasticsearch logging endpoint log (service 123 version 4)",
		},
		{
			args: []string{"logging", "elasticsearch", "create", "--service-id", "123", "--version", "1", "--name", "log", "--index", "logs", "--url", "example.com", "--autoclone"},
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				CreateElasticsearchFn: createElasticsearchError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, testcase.api, &stdout)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestElasticsearchList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "elasticsearch", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsShortOutput,
		},
		{
			args: []string{"logging", "elasticsearch", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: []string{"logging", "elasticsearch", "list", "--service-id", "123", "--version", "1", "-v"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: []string{"logging", "elasticsearch", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "elasticsearch", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsOK,
			},
			wantOutput: listElasticsearchsVerboseOutput,
		},
		{
			args: []string{"logging", "elasticsearch", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListElasticsearchFn: listElasticsearchsError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, testcase.api, &stdout)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestElasticsearchDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "elasticsearch", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "elasticsearch", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				GetElasticsearchFn: getElasticsearchError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "elasticsearch", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				GetElasticsearchFn: getElasticsearchOK,
			},
			wantOutput: describeElasticsearchOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, testcase.api, &stdout)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestElasticsearchUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "elasticsearch", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "elasticsearch", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				UpdateElasticsearchFn: updateElasticsearchError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "elasticsearch", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				UpdateElasticsearchFn: updateElasticsearchOK,
			},
			wantOutput: "Updated Elasticsearch logging endpoint log (service 123 version 4)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, testcase.api, &stdout)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestElasticsearchDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "elasticsearch", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "elasticsearch", "delete", "--service-id", "123", "--version", "1", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				DeleteElasticsearchFn: deleteElasticsearchError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "elasticsearch", "delete", "--service-id", "123", "--version", "1", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				DeleteElasticsearchFn: deleteElasticsearchOK,
			},
			wantOutput: "Deleted Elasticsearch logging endpoint logs (service 123 version 4)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			ara := testutil.NewAppRunArgs(testcase.args, testcase.api, &stdout)
			err := app.Run(ara.Args, ara.Env, ara.File, ara.AppConfigFile, ara.ClientFactory, ara.HTTPClient, ara.CLIVersioner, ara.In, ara.Out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createElasticsearchOK(i *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return &fastly.Elasticsearch{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Index:             "logs",
		URL:               "example.com",
		Pipeline:          "logs",
		User:              "user",
		Password:          "password",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
	}, nil
}

func createElasticsearchError(i *fastly.CreateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

func listElasticsearchsOK(i *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
	return []*fastly.Elasticsearch{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			Index:             "logs",
			URL:               "example.com",
			Pipeline:          "logs",
			User:              "user",
			Password:          "password",
			RequestMaxEntries: 2,
			RequestMaxBytes:   2,
			Placement:         "none",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "example.com",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			FormatVersion:     2,
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Index:             "analytics",
			URL:               "example.com",
			Pipeline:          "analytics",
			User:              "user",
			Password:          "password",
			RequestMaxEntries: 2,
			RequestMaxBytes:   2,
			Placement:         "none",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "example.com",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
		},
	}, nil
}

func listElasticsearchsError(i *fastly.ListElasticsearchInput) ([]*fastly.Elasticsearch, error) {
	return nil, errTest
}

var listElasticsearchsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listElasticsearchsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
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
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Index:             "logs",
		URL:               "example.com",
		Pipeline:          "logs",
		User:              "user",
		Password:          "password",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
	}, nil
}

func getElasticsearchError(i *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

var describeElasticsearchOutput = strings.TrimSpace(`
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
`) + "\n"

func updateElasticsearchOK(i *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return &fastly.Elasticsearch{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Index:             "logs",
		URL:               "example.com",
		Pipeline:          "logs",
		User:              "user",
		Password:          "password",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
	}, nil
}

func updateElasticsearchError(i *fastly.UpdateElasticsearchInput) (*fastly.Elasticsearch, error) {
	return nil, errTest
}

func deleteElasticsearchOK(i *fastly.DeleteElasticsearchInput) error {
	return nil
}

func deleteElasticsearchError(i *fastly.DeleteElasticsearchInput) error {
	return errTest
}
