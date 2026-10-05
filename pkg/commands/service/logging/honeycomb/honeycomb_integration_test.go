package honeycomb_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/honeycomb"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHoneycombCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --auth-token abc --dataset log --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateHoneycombFn: createHoneycombOK,
			},
			WantOutput: "Created Honeycomb logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --auth-token abc --dataset log --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateHoneycombFn: createHoneycombError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestHoneycombList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:     testutil.GetVersion,
				ListHoneycombsFn: listHoneycombsOK,
			},
			WantOutput: listHoneycombsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: &mock.API{
				GetVersionFn:     testutil.GetVersion,
				ListHoneycombsFn: listHoneycombsOK,
			},
			WantOutput: listHoneycombsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: &mock.API{
				GetVersionFn:     testutil.GetVersion,
				ListHoneycombsFn: listHoneycombsOK,
			},
			WantOutput: listHoneycombsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:     testutil.GetVersion,
				ListHoneycombsFn: listHoneycombsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestHoneycombDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				GetHoneycombFn: getHoneycombError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				GetHoneycombFn: getHoneycombOK,
			},
			WantOutput: describeHoneycombOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestHoneycombUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateHoneycombFn: updateHoneycombError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateHoneycombFn: updateHoneycombOK,
			},
			WantOutput: "Updated Honeycomb logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestHoneycombDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteHoneycombFn: deleteHoneycombError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteHoneycombFn: deleteHoneycombOK,
			},
			WantOutput: "Deleted Honeycomb logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createHoneycombOK(_ context.Context, i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	s := fastly.Honeycomb{
		ServiceID:      new(i.ServiceID),
		ServiceVersion: new(i.ServiceVersion),
	}

	if i.Name != nil {
		s.Name = i.Name
	}

	return &s, nil
}

func createHoneycombError(_ context.Context, _ *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func listHoneycombsOK(_ context.Context, i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return []*fastly.Honeycomb{
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("logs"),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			Dataset:           new("log"),
			Token:             new("tkn"),
			ResponseCondition: new("Prevent default logging"),
			Placement:         new("none"),
			ProcessingRegion:  new("us"),
		},
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("analytics"),
			Dataset:           new("log"),
			Token:             new("tkn"),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			ResponseCondition: new("Prevent default logging"),
			Placement:         new("none"),
			ProcessingRegion:  new("us"),
		},
	}, nil
}

func listHoneycombsError(_ context.Context, _ *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return nil, errTest
}

var listHoneycombsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHoneycombsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (auth: user)

Service ID (via --service-id): 123

Version: 1
	Honeycomb 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Dataset: log
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
	Honeycomb 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Dataset: log
		Token: tkn
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
`) + "\n\n"

func getHoneycombOK(_ context.Context, i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("logs"),
		Dataset:           new("log"),
		Token:             new("tkn"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
		ProcessingRegion:  new("us"),
	}, nil
}

func getHoneycombError(_ context.Context, _ *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

var describeHoneycombOutput = "\n" + strings.TrimSpace(`
Dataset: log
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Processing region: us
Response condition: Prevent default logging
Service ID: 123
Token: tkn
Version: 1
`) + "\n"

func updateHoneycombOK(_ context.Context, i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		Dataset:           new("log"),
		Token:             new("tkn"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
	}, nil
}

func updateHoneycombError(_ context.Context, _ *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func deleteHoneycombOK(_ context.Context, _ *fastly.DeleteHoneycombInput) error {
	return nil
}

func deleteHoneycombError(_ context.Context, _ *fastly.DeleteHoneycombInput) error {
	return errTest
}
