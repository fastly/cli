package bigquery_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestBigQueryCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging bigquery create --service-id 123 --version 1 --name log --project-id project123 --dataset logs --table logs --user user@domain.com --secret-key `\"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA\"` --autoclone"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				CreateBigQueryFn: createBigQueryOK,
			},
			wantOutput: "Created BigQuery logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging bigquery create --service-id 123 --version 1 --name log --project-id project123 --dataset logs --table logs --user user@domain.com --secret-key `\"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA\"` --autoclone"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				CreateBigQueryFn: createBigQueryError,
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

func TestBigQueryList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging bigquery list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			wantOutput: listBigQueriesShortOutput,
		},
		{
			args: args("logging bigquery list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args: args("logging bigquery list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args: args("logging bigquery --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args: args("logging -v bigquery list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			wantOutput: listBigQueriesVerboseOutput,
		},
		{
			args: args("logging bigquery list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesError,
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

func TestBigQueryDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging bigquery describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging bigquery describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetBigQueryFn:  getBigQueryError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging bigquery describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetBigQueryFn:  getBigQueryOK,
			},
			wantOutput: describeBigQueryOutput,
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

func TestBigQueryUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging bigquery update --service-id 123 --version 1 --new-name log --project-id project123 --dataset logs --table logs --user user@domain.com --secret-key `\"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA\"`"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging bigquery update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				UpdateBigQueryFn: updateBigQueryError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging bigquery update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				UpdateBigQueryFn: updateBigQueryOK,
			},
			wantOutput: "Updated BigQuery logging endpoint log (service 123 version 4)",
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

func TestBigQueryDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging bigquery delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging bigquery delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				DeleteBigQueryFn: deleteBigQueryError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging bigquery delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				DeleteBigQueryFn: deleteBigQueryOK,
			},
			wantOutput: "Deleted BigQuery logging endpoint logs (service 123 version 4)",
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

func createBigQueryOK(i *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createBigQueryError(_ *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

func listBigQueriesOK(i *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
	return []*fastly.BigQuery{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			ProjectID:         fastly.ToPointer("my-project"),
			Dataset:           fastly.ToPointer("raw-logs"),
			Table:             fastly.ToPointer("logs"),
			User:              fastly.ToPointer("service-account@domain.com"),
			AccountName:       fastly.ToPointer("none"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			Template:          fastly.ToPointer("%Y%m%d"),
			Placement:         fastly.ToPointer("none"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			ProjectID:         fastly.ToPointer("my-project"),
			Dataset:           fastly.ToPointer("analytics"),
			Table:             fastly.ToPointer("logs"),
			User:              fastly.ToPointer("service-account@domain.com"),
			AccountName:       fastly.ToPointer("none"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			Template:          fastly.ToPointer("%Y%m%d"),
			Placement:         fastly.ToPointer("none"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listBigQueriesError(_ *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
	return nil, errTest
}

var listBigQueriesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listBigQueriesVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	BigQuery 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Format: %h %l %u %t "%r" %>s %b
		User: service-account@domain.com
		Account name: none
		Project ID: my-project
		Dataset: raw-logs
		Table: logs
		Template suffix: %Y%m%d
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Response condition: Prevent default logging
		Placement: none
		Format version: 0
		Processing region: us
	BigQuery 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Format: %h %l %u %t "%r" %>s %b
		User: service-account@domain.com
		Account name: none
		Project ID: my-project
		Dataset: analytics
		Table: logs
		Template suffix: %Y%m%d
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Response condition: Prevent default logging
		Placement: none
		Format version: 0
		Processing region: us
`) + "\n\n"

func getBigQueryOK(i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		ProjectID:         fastly.ToPointer("my-project"),
		Dataset:           fastly.ToPointer("raw-logs"),
		Table:             fastly.ToPointer("logs"),
		User:              fastly.ToPointer("service-account@domain.com"),
		AccountName:       fastly.ToPointer("none"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Template:          fastly.ToPointer("%Y%m%d"),
		Placement:         fastly.ToPointer("none"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getBigQueryError(_ *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

var describeBigQueryOutput = "\n" + strings.TrimSpace(`
Account name: none
Dataset: raw-logs
Format: %h %l %u %t "%r" %>s %b
Format version: 0
Name: logs
Placement: none
Processing region: us
Project ID: my-project
Response condition: Prevent default logging
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Service ID: 123
Table: logs
Template suffix: %Y%m%d
User: service-account@domain.com
Version: 1
`) + "\n"

func updateBigQueryOK(i *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ProjectID:         fastly.ToPointer("my-project"),
		Dataset:           fastly.ToPointer("raw-logs"),
		Table:             fastly.ToPointer("logs"),
		User:              fastly.ToPointer("service-account@domain.com"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Template:          fastly.ToPointer("%Y%m%d"),
		Placement:         fastly.ToPointer("none"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
	}, nil
}

func updateBigQueryError(_ *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

func deleteBigQueryOK(_ *fastly.DeleteBigQueryInput) error {
	return nil
}

func deleteBigQueryError(_ *fastly.DeleteBigQueryInput) error {
	return errTest
}
