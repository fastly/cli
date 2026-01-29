package logshuttle_test

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
	sub "github.com/fastly/cli/pkg/commands/service/logging/logshuttle"
)

func TestLogshuttleCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --url example.com --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateLogshuttleFn: createLogshuttleOK,
			},
			WantOutput: "Created Logshuttle logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --url example.com --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateLogshuttleFn: createLogshuttleError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestLogshuttleList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			WantOutput: listLogshuttlesShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			WantOutput: listLogshuttlesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			WantOutput: listLogshuttlesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestLogshuttleDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetLogshuttleFn: getLogshuttleError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetLogshuttleFn: getLogshuttleOK,
			},
			WantOutput: describeLogshuttleOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestLogshuttleUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateLogshuttleFn: updateLogshuttleError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateLogshuttleFn: updateLogshuttleOK,
			},
			WantOutput: "Updated Logshuttle logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestLogshuttleDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteLogshuttleFn: deleteLogshuttleError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteLogshuttleFn: deleteLogshuttleOK,
			},
			WantOutput: "Deleted Logshuttle logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createLogshuttleOK(_ context.Context, i *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error) {
	s := fastly.Logshuttle{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
	}

	if i.Name != nil {
		s.Name = i.Name
	}

	return &s, nil
}

func createLogshuttleError(_ context.Context, _ *fastly.CreateLogshuttleInput) (*fastly.Logshuttle, error) {
	return nil, errTest
}

func listLogshuttlesOK(_ context.Context, i *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error) {
	return []*fastly.Logshuttle{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			URL:               fastly.ToPointer("example.com"),
			Token:             fastly.ToPointer("abc"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			URL:               fastly.ToPointer("example.com"),
			Token:             fastly.ToPointer("abc"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listLogshuttlesError(_ context.Context, _ *fastly.ListLogshuttlesInput) ([]*fastly.Logshuttle, error) {
	return nil, errTest
}

var listLogshuttlesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listLogshuttlesVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Logshuttle 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
	Logshuttle 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: example.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
`) + "\n\n"

func getLogshuttleOK(_ context.Context, i *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return &fastly.Logshuttle{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		URL:               fastly.ToPointer("example.com"),
		Token:             fastly.ToPointer("abc"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getLogshuttleError(_ context.Context, _ *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return nil, errTest
}

var describeLogshuttleOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Processing region: us
Response condition: Prevent default logging
Service ID: 123
Token: abc
URL: example.com
Version: 1
`) + "\n"

func updateLogshuttleOK(_ context.Context, i *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error) {
	return &fastly.Logshuttle{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		URL:               fastly.ToPointer("example.com"),
		Token:             fastly.ToPointer("abc"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func updateLogshuttleError(_ context.Context, _ *fastly.UpdateLogshuttleInput) (*fastly.Logshuttle, error) {
	return nil, errTest
}

func deleteLogshuttleOK(_ context.Context, _ *fastly.DeleteLogshuttleInput) error {
	return nil
}

func deleteLogshuttleError(_ context.Context, _ *fastly.DeleteLogshuttleInput) error {
	return errTest
}
