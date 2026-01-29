package bigquery_test

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
	sub "github.com/fastly/cli/pkg/commands/service/logging/bigquery"
)

func TestBigQueryCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --project-id project123 --dataset logs --table logs --user user@domain.com --secret-key `\"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA\"` --autoclone",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				CreateBigQueryFn: createBigQueryOK,
			},
			WantOutput: "Created BigQuery logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --project-id project123 --dataset logs --table logs --user user@domain.com --secret-key `\"-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA\"` --autoclone",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				CreateBigQueryFn: createBigQueryError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestBigQueryList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			WantOutput: listBigQueriesShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			WantOutput: listBigQueriesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesOK,
			},
			WantOutput: listBigQueriesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListBigQueriesFn: listBigQueriesError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestBigQueryDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetBigQueryFn:  getBigQueryError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetBigQueryFn:  getBigQueryOK,
			},
			WantOutput: describeBigQueryOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestBigQueryUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				UpdateBigQueryFn: updateBigQueryError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				UpdateBigQueryFn: updateBigQueryOK,
			},
			WantOutput: "Updated BigQuery logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestBigQueryDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				DeleteBigQueryFn: deleteBigQueryError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				DeleteBigQueryFn: deleteBigQueryOK,
			},
			WantOutput: "Deleted BigQuery logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createBigQueryOK(_ context.Context, i *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createBigQueryError(_ context.Context, _ *fastly.CreateBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

func listBigQueriesOK(_ context.Context, i *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
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

func listBigQueriesError(_ context.Context, _ *fastly.ListBigQueriesInput) ([]*fastly.BigQuery, error) {
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

func getBigQueryOK(_ context.Context, i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
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

func getBigQueryError(_ context.Context, _ *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
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

func updateBigQueryOK(_ context.Context, i *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
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

func updateBigQueryError(_ context.Context, _ *fastly.UpdateBigQueryInput) (*fastly.BigQuery, error) {
	return nil, errTest
}

func deleteBigQueryOK(_ context.Context, _ *fastly.DeleteBigQueryInput) error {
	return nil
}

func deleteBigQueryError(_ context.Context, _ *fastly.DeleteBigQueryInput) error {
	return errTest
}
