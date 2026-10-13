package papertrail_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/papertrail"
)

func TestPapertrailCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --address example.com:123 --autoclone",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreatePapertrailFn: createPapertrailOK,
			},
			WantOutput: "Created Papertrail logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --address example.com:123 --autoclone",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreatePapertrailFn: createPapertrailError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestPapertrailList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ListPapertrailsFn: listPapertrailsOK,
			},
			WantOutput: listPapertrailsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ListPapertrailsFn: listPapertrailsOK,
			},
			WantOutput: listPapertrailsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ListPapertrailsFn: listPapertrailsOK,
			},
			WantOutput: listPapertrailsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:      testutil.GetVersion,
				ListPapertrailsFn: listPapertrailsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestPapertrailDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				GetPapertrailFn: getPapertrailError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				GetPapertrailFn: getPapertrailOK,
			},
			WantOutput: describePapertrailOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestPapertrailUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdatePapertrailFn: updatePapertrailError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdatePapertrailFn: updatePapertrailOK,
			},
			WantOutput: "Updated Papertrail logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestPapertrailDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeletePapertrailFn: deletePapertrailError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeletePapertrailFn: deletePapertrailOK,
			},
			WantOutput: "Deleted Papertrail logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createPapertrailOK(_ context.Context, i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:      new(i.ServiceID),
		ServiceVersion: new(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createPapertrailError(_ context.Context, _ *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func listPapertrailsOK(_ context.Context, i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return []*fastly.Papertrail{
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("logs"),
			Address:           new("example.com:123"),
			Port:              new(123),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			ResponseCondition: new("Prevent default logging"),
			Placement:         new("none"),
			ProcessingRegion:  new("us"),
		},
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("analytics"),
			Address:           new("127.0.0.1:456"),
			Port:              new(456),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			ResponseCondition: new("Prevent default logging"),
			Placement:         new("none"),
			ProcessingRegion:  new("us"),
		},
	}, nil
}

func listPapertrailsError(_ context.Context, _ *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return nil, errTest
}

var listPapertrailsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listPapertrailsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (auth: user)

Service ID (via --service-id): 123

Version: 1
	Papertrail 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Address: example.com:123
		Port: 123
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
	Papertrail 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: 127.0.0.1:456
		Port: 456
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Processing region: us
`) + "\n\n"

func getPapertrailOK(_ context.Context, i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("logs"),
		Address:           new("example.com:123"),
		Port:              new(123),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
		ProcessingRegion:  new("us"),
	}, nil
}

func getPapertrailError(_ context.Context, _ *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

var describePapertrailOutput = "\n" + strings.TrimSpace(`
Address: example.com:123
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Port: 123
Processing region: us
Response condition: Prevent default logging
Service ID: 123
Version: 1
`) + "\n"

func updatePapertrailOK(_ context.Context, i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		Address:           new("example.com:123"),
		Port:              new(123),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
	}, nil
}

func updatePapertrailError(_ context.Context, _ *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func deletePapertrailOK(_ context.Context, _ *fastly.DeletePapertrailInput) error {
	return nil
}

func deletePapertrailError(_ context.Context, _ *fastly.DeletePapertrailInput) error {
	return errTest
}
