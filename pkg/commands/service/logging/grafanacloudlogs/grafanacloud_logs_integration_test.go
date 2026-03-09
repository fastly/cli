package grafanacloudlogs_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/grafanacloudlogs"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestGrafanaCloudLogsCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --user 123456 --url https://test123.grafana.net --auth-token testtoken --index `{\"label\": \"value\" }` --autoclone",
			API: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				CreateGrafanaCloudLogsFn: createGrafanaCloudLogsOK,
			},
			WantOutput: "Created Grafana Cloud Logs logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --url https://test123.grafana.net --auth-token testtoken --index `{\"label\": \"value\" }` --autoclone",
			API: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				CreateGrafanaCloudLogsFn: createGrafanaCloudLogsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestGrafanaCloudLogsList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			WantOutput: listGrafanaCloudLogsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			WantOutput: listGrafanaCloudLogsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			WantOutput: listGrafanaCloudLogsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestGrafanaCloudLogsDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				GetGrafanaCloudLogsFn: getGrafanaCloudLogsError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				GetGrafanaCloudLogsFn: getGrafanaCloudLogsOK,
			},
			WantOutput: describeGrafanaCloudLogsOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestGrafanaCloudLogsUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				UpdateGrafanaCloudLogsFn: updateGrafanaCloudLogsError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				UpdateGrafanaCloudLogsFn: updateGrafanaCloudLogsOK,
			},
			WantOutput: "Updated Grafana Cloud Logs logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestGrafanaCloudLogsDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				DeleteGrafanaCloudLogsFn: deleteGrafanaCloudLogsError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				DeleteGrafanaCloudLogsFn: deleteGrafanaCloudLogsOK,
			},
			WantOutput: "Deleted Grafana Cloud Logs logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createGrafanaCloudLogsOK(_ context.Context, i *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return &fastly.GrafanaCloudLogs{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createGrafanaCloudLogsError(_ context.Context, _ *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

func listGrafanaCloudLogsOK(_ context.Context, i *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error) {
	return []*fastly.GrafanaCloudLogs{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			User:              fastly.ToPointer("123456"),
			Token:             fastly.ToPointer("testtoken"),
			URL:               fastly.ToPointer("https://test123.grafana.net"),
			Index:             fastly.ToPointer("{\"label\": \"value\"}"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			User:              fastly.ToPointer("123456"),
			Token:             fastly.ToPointer("testtoken"),
			URL:               fastly.ToPointer("https://test123.grafana.net"),
			Index:             fastly.ToPointer("{\"label\": \"value\"}"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listGrafanaCloudLogsError(_ context.Context, _ *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

var listGrafanaCloudLogsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listGrafanaCloudLogsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	GrafanaCloudLogs 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
		User: 123456
		URL: https://test123.grafana.net
		Token: testtoken
		Index: {"label": "value"}
		Processing region: us
	GrafanaCloudLogs 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
		User: 123456
		URL: https://test123.grafana.net
		Token: testtoken
		Index: {"label": "value"}
		Processing region: us
`) + "\n\n"

func getGrafanaCloudLogsOK(_ context.Context, i *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return &fastly.GrafanaCloudLogs{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		MessageType:       fastly.ToPointer("classic"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		User:              fastly.ToPointer("123456"),
		URL:               fastly.ToPointer("https://test123.grafana.net"),
		Token:             fastly.ToPointer("testtoken"),
		Index:             fastly.ToPointer("{\"label\": \"value\"}"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getGrafanaCloudLogsError(_ context.Context, _ *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

var describeGrafanaCloudLogsOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Index: {"label": "value"}
Message type: classic
Name: logs
Placement: none
Processing region: us
Response condition: Prevent default logging
Service ID: 123
Token: testtoken
URL: https://test123.grafana.net
User: 123456
Version: 1
`) + "\n"

func updateGrafanaCloudLogsOK(_ context.Context, i *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return &fastly.GrafanaCloudLogs{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		Placement:         fastly.ToPointer("none"),
		User:              fastly.ToPointer("123456"),
		URL:               fastly.ToPointer("https://test123.grafana.net"),
		Token:             fastly.ToPointer("testtoken"),
		Index:             fastly.ToPointer("{\"label\": \"value\"}"),
	}, nil
}

func updateGrafanaCloudLogsError(_ context.Context, _ *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

func deleteGrafanaCloudLogsOK(_ context.Context, _ *fastly.DeleteGrafanaCloudLogsInput) error {
	return nil
}

func deleteGrafanaCloudLogsError(_ context.Context, _ *fastly.DeleteGrafanaCloudLogsInput) error {
	return errTest
}
