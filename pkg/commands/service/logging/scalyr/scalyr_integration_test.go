package scalyr_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	fsterrs "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/scalyr"
)

func TestScalyrCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--name log --version 1 --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: fsterrs.ErrNoServiceID.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name log --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateScalyrFn: createScalyrOK,
			},
			WantOutput: "Created Scalyr logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --auth-token abc --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateScalyrFn: createScalyrError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestScalyrList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			WantOutput: listScalyrsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			WantOutput: listScalyrsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			WantOutput: listScalyrsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestScalyrDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetScalyrFn:    getScalyrError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetScalyrFn:    getScalyrOK,
			},
			WantOutput: describeScalyrOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestScalyrUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateScalyrFn: updateScalyrError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateScalyrFn: updateScalyrOK,
			},
			WantOutput: "Updated Scalyr logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestScalyrDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteScalyrFn: deleteScalyrError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteScalyrFn: deleteScalyrOK,
			},
			WantOutput: "Deleted Scalyr logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createScalyrOK(_ context.Context, i *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	s := fastly.Scalyr{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
	}

	// Avoids null pointer dereference for test cases with missing required params.
	// If omitted, tests are guaranteed to panic.
	if i.Name != nil {
		s.Name = i.Name
	}

	if i.Token != nil {
		s.Token = i.Token
	}

	if i.Format != nil {
		s.Format = i.Format
	}

	if i.FormatVersion != nil {
		s.FormatVersion = i.FormatVersion
	}

	if i.ResponseCondition != nil {
		s.ResponseCondition = i.ResponseCondition
	}

	if i.Placement != nil {
		s.Placement = i.Placement
	}

	return &s, nil
}

func createScalyrError(_ context.Context, _ *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

func listScalyrsOK(_ context.Context, i *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
	return []*fastly.Scalyr{
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
			ProjectID:         fastly.ToPointer("example-project"),
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
			ProjectID:         fastly.ToPointer("example-project"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listScalyrsError(_ context.Context, _ *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
	return nil, errTest
}

var listScalyrsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listScalyrsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Scalyr 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Project ID: example-project
		Processing region: us
	Scalyr 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Token: abc
		Region: US
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Project ID: example-project
		Processing region: us
`) + "\n\n"

func getScalyrOK(_ context.Context, i *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return &fastly.Scalyr{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Token:             fastly.ToPointer("abc"),
		Region:            fastly.ToPointer("US"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		ProjectID:         fastly.ToPointer("example-project"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getScalyrError(_ context.Context, _ *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

var describeScalyrOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Processing region: us
Project ID: example-project
Region: US
Response condition: Prevent default logging
Service ID: 123
Token: abc
Version: 1
`) + "\n"

func updateScalyrOK(_ context.Context, i *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return &fastly.Scalyr{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Token:             fastly.ToPointer("abc"),
		Region:            fastly.ToPointer("EU"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func updateScalyrError(_ context.Context, _ *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

func deleteScalyrOK(_ context.Context, _ *fastly.DeleteScalyrInput) error {
	return nil
}

func deleteScalyrError(_ context.Context, _ *fastly.DeleteScalyrInput) error {
	return errTest
}
