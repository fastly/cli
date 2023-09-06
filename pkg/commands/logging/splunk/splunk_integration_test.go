package splunk_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestSplunkCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging splunk create --service-id 123 --version 1 --name log --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSplunkFn: createSplunkOK,
			},
			wantOutput: "Created Splunk logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging splunk create --service-id 123 --version 1 --name log --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateSplunkFn: createSplunkError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestSplunkList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging splunk list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSplunksFn:  listSplunksOK,
			},
			wantOutput: listSplunksShortOutput,
		},
		{
			args: args("logging splunk list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSplunksFn:  listSplunksOK,
			},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args: args("logging splunk list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSplunksFn:  listSplunksOK,
			},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args: args("logging splunk --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSplunksFn:  listSplunksOK,
			},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args: args("logging -v splunk list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSplunksFn:  listSplunksOK,
			},
			wantOutput: listSplunksVerboseOutput,
		},
		{
			args: args("logging splunk list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListSplunksFn:  listSplunksError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestSplunkDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging splunk describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging splunk describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSplunkFn:    getSplunkError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging splunk describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetSplunkFn:    getSplunkOK,
			},
			wantOutput: describeSplunkOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestSplunkUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging splunk update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging splunk update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSplunkFn: updateSplunkError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging splunk update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateSplunkFn: updateSplunkOK,
			},
			wantOutput: "Updated Splunk logging endpoint log (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestSplunkDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging splunk delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging splunk delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSplunkFn: deleteSplunkError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging splunk delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteSplunkFn: deleteSplunkOK,
			},
			wantOutput: "Deleted Splunk logging endpoint logs (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createSplunkOK(i *fastly.CreateSplunkInput) (*fastly.Splunk, error) {
	return &fastly.Splunk{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
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
Fastly API token provided via config file (profile: user)
Fastly API endpoint: https://api.fastly.com

Service ID (via --service-id): 123

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

var describeSplunkOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Response condition: Prevent default logging
Service ID: 123
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
TLS hostname: example.com
Token: tkn
URL: example.com
Version: 1
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
