package bigquery_test

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

func TestBigQueryCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "bigquery", "create", "--service-id", "123", "--version", "1", "--name", "log", "--project-id", "project123", "--dataset", "logs", "--table", "logs", "--user", "user@domain.com"},
			api:       mock.API{CreateBigQueryFn: createBigQueryOK},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args:       []string{"logging", "bigquery", "create", "--service-id", "123", "--version", "1", "--name", "log", "--project-id", "project123", "--dataset", "logs", "--table", "logs", "--user", "user@domain.com", "--secret-key", `"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"`},
			api:        mock.API{CreateBigQueryFn: createBigQueryOK},
			wantOutput: "Created BigQuery logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "bigquery", "create", "--service-id", "123", "--version", "1", "--name", "log", "--project-id", "project123", "--dataset", "logs", "--table", "logs", "--user", "user@domain.com", "--secret-key", `"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"`},
			api:       mock.API{CreateBigQueryFn: createBigQueryError},
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

func TestBigQueryList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "bigquery", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListBigQueriesFn: listBigQueriesOK},
			wantOutput: listBigQueriesShortOutput,
		},
		{
			args:       []string{"logging", "bigquery", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListBigQueriesFn: listBigQueriesOK},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args:       []string{"logging", "bigquery", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListBigQueriesFn: listBigQueriesOK},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args:       []string{"logging", "bigquery", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListBigQueriesFn: listBigQueriesOK},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "bigquery", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListBigQueriesFn: listBigQueriesOK},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args:      []string{"logging", "bigquery", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListBigQueriesFn: listBigQueriesError},
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

func TestBigQueryDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "bigquery", "describe", "--service-id", "123", "--version", "1"},
			api:       mock.API{GetBigQueryFn: getBigQueryOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "bigquery", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetBigQueryFn: getBigQueryError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "bigquery", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetBigQueryFn: getBigQueryOK},
			wantOutput: describeBigQueryOutput,
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

func TestBigQueryUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "bigquery", "update", "--service-id", "123", "--version", "1", "--new-name", "log", "--project-id", "project123", "--dataset", "logs", "--table", "logs", "--user", "user@domain.com", "--secret-key", `"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"`},
			api:       mock.API{UpdateBigQueryFn: updateBigQueryOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "bigquery", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetBigQueryFn:    getBigQueryError,
				UpdateBigQueryFn: updateBigQueryOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "bigquery", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetBigQueryFn:    getBigQueryOK,
				UpdateBigQueryFn: updateBigQueryError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "bigquery", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetBigQueryFn:    getBigQueryOK,
				UpdateBigQueryFn: updateBigQueryOK,
			},
			wantOutput: "Updated BigQuery logging endpoint log (service 123 version 1)",
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

func TestBigQueryDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "bigquery", "delete", "--service-id", "123", "--version", "1"},
			api:       mock.API{DeleteBigQueryFn: deleteBigQueryOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "bigquery", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteBigQueryFn: deleteBigQueryError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "bigquery", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteBigQueryFn: deleteBigQueryOK},
			wantOutput: "Deleted BigQuery logging endpoint logs (service 123 version 1)",
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

func createBigQueryOK(i *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createBigQueryError(i *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

func listBigQueriesOK(i *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
	return []*fastly.BigQuery{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			ProjectID:         "my-project",
			Dataset:           "raw-logs",
			Table:             "logs",
			User:              "service-account@domain.com",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Format:            `%h %l %u %t "%r" %>s %b`,
			Template:          "%Y%m%d",
			Placement:         "none",
			ResponseCondition: "Prevent default logging",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			ProjectID:         "my-project",
			Dataset:           "analytics",
			Table:             "logs",
			User:              "service-account@domain.com",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Format:            `%h %l %u %t "%r" %>s %b`,
			Template:          "%Y%m%d",
			Placement:         "none",
			ResponseCondition: "Prevent default logging",
		},
	}, nil
}

func listBigQueriesError(i *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
	return nil, errTest
}

var listBigQueriesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listBigQueriesVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	BigQuery 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Format: %h %l %u %t "%r" %>s %b
		User: service-account@domain.com
		Project ID: my-project
		Dataset: raw-logs
		Table: logs
		Template suffix: %Y%m%d
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Response condition: Prevent default logging
		Placement: none
		Format version: 0
	BigQuery 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Format: %h %l %u %t "%r" %>s %b
		User: service-account@domain.com
		Project ID: my-project
		Dataset: analytics
		Table: logs
		Template suffix: %Y%m%d
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Response condition: Prevent default logging
		Placement: none
		Format version: 0
`) + "\n\n"

func getBigQueryOK(i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		ProjectID:         "my-project",
		Dataset:           "raw-logs",
		Table:             "logs",
		User:              "service-account@domain.com",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Template:          "%Y%m%d",
		Placement:         "none",
		ResponseCondition: "Prevent default logging",
	}, nil
}

func getBigQueryError(i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

var describeBigQueryOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Format: %h %l %u %t "%r" %>s %b
User: service-account@domain.com
Project ID: my-project
Dataset: raw-logs
Table: logs
Template suffix: %Y%m%d
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Response condition: Prevent default logging
Placement: none
Format version: 0
`) + "\n"

func updateBigQueryOK(i *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ProjectID:         "my-project",
		Dataset:           "raw-logs",
		Table:             "logs",
		User:              "service-account@domain.com",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Template:          "%Y%m%d",
		Placement:         "none",
		ResponseCondition: "Prevent default logging",
	}, nil
}

func updateBigQueryError(i *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

func deleteBigQueryOK(i *fastly.DeleteBigQueryInput) error {
	return nil
}

func deleteBigQueryError(i *fastly.DeleteBigQueryInput) error {
	return errTest
}
