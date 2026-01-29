package datadog_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/datadog"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDatadogCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateDatadogFn: createDatadogOK,
			},
			WantOutput: "Created Datadog logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateDatadogFn: createDatadogError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestDatadogList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDatadogFn:  listDatadogsOK,
			},
			WantOutput: listDatadogsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDatadogFn:  listDatadogsOK,
			},
			WantOutput: listDatadogsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDatadogFn:  listDatadogsOK,
			},
			WantOutput: listDatadogsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListDatadogFn:  listDatadogsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestDatadogDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetDatadogFn:   getDatadogError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetDatadogFn:   getDatadogOK,
			},
			WantOutput: describeDatadogOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestDatadogUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateDatadogFn: updateDatadogError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateDatadogFn: updateDatadogOK,
			},
			WantOutput: "Updated Datadog logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestDatadogDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteDatadogFn: deleteDatadogError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteDatadogFn: deleteDatadogOK,
			},
			WantOutput: "Deleted Datadog logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createDatadogOK(_ context.Context, i *fastly.CreateDatadogInput) (*fastly.Datadog, error) {
	s := fastly.Datadog{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
	}

	if i.Name != nil {
		s.Name = i.Name
	}

	return &s, nil
}

func createDatadogError(_ context.Context, _ *fastly.CreateDatadogInput) (*fastly.Datadog, error) {
	return nil, errTest
}

func listDatadogsOK(_ context.Context, i *fastly.ListDatadogInput) ([]*fastly.Datadog, error) {
	return []*fastly.Datadog{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Token:             fastly.ToPointer("abc"),
			Region:            fastly.ToPointer("US"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Token:             fastly.ToPointer("abc"),
			Region:            fastly.ToPointer("US"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listDatadogsError(_ context.Context, _ *fastly.ListDatadogInput) ([]*fastly.Datadog, error) {
	return nil, errTest
}

var listDatadogsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listDatadogsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Datadog 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
	Datadog 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
`) + "\n\n"

func getDatadogOK(_ context.Context, i *fastly.GetDatadogInput) (*fastly.Datadog, error) {
	return &fastly.Datadog{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Token:             fastly.ToPointer("abc"),
		Region:            fastly.ToPointer("US"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getDatadogError(_ context.Context, _ *fastly.GetDatadogInput) (*fastly.Datadog, error) {
	return nil, errTest
}

var describeDatadogOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Processing region: us
Region: US
Response condition: Prevent default logging
Service ID: 123
Token: abc
Version: 1
`) + "\n"

func updateDatadogOK(_ context.Context, i *fastly.UpdateDatadogInput) (*fastly.Datadog, error) {
	return &fastly.Datadog{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Token:             fastly.ToPointer("abc"),
		Region:            fastly.ToPointer("US"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
	}, nil
}

func updateDatadogError(_ context.Context, _ *fastly.UpdateDatadogInput) (*fastly.Datadog, error) {
	return nil, errTest
}

func deleteDatadogOK(_ context.Context, _ *fastly.DeleteDatadogInput) error {
	return nil
}

func deleteDatadogError(_ context.Context, _ *fastly.DeleteDatadogInput) error {
	return errTest
}
