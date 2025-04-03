package condition_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/vcl"
	sub "github.com/fastly/cli/pkg/commands/vcl/condition"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestConditionCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--version 1",
			WantError: "error reading service: no service ID found",
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --statement false --type REQUEST --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateConditionFn: createConditionOK,
			},
			WantOutput: "Created condition always_false (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --statement false --type REQUEST --priority 10 --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateConditionFn: createConditionError,
			},
			WantError: errTest.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestConditionDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteConditionFn: deleteConditionError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteConditionFn: deleteConditionOK,
			},
			WantOutput: "Deleted condition always_false (service 123 version 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestConditionUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name false_always --comment ",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateConditionFn: updateConditionOK,
			},
			WantError: "error parsing arguments: must provide either --new-name, --statement, --type or --priority to update condition",
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --new-name false_always --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateConditionFn: updateConditionError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name always_false --new-name false_always --autoclone",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateConditionFn: updateConditionOK,
			},
			WantOutput: "Updated condition false_always (service 123 version 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestConditionDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name always_false",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetConditionFn: getConditionError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name always_false",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetConditionFn: getConditionOK,
			},
			WantOutput: describeConditionOutput,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestConditionList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: "--verbose --service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: "-v --service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsError,
			},
			WantError: errTest.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

var describeConditionOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Version: 1
Name: always_false
Statement: false
Type: CACHE
Priority: 10
`) + "\n"

var listConditionsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME                  STATEMENT  TYPE     PRIORITY
123      1        always_false_request  false      REQUEST  10
123      1        always_false_cache    false      CACHE    10
`) + "\n"

var listConditionsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	Condition 1/2
		Name: always_false_request
		Statement: false
		Type: REQUEST
		Priority: 10
	Condition 2/2
		Name: always_false_cache
		Statement: false
		Type: CACHE
		Priority: 10
`) + "\n\n"

var errTest = errors.New("fixture error")

func createConditionOK(i *fastly.CreateConditionInput) (*fastly.Condition, error) {
	priority := 10
	if i.Priority != nil {
		priority = *i.Priority
	}

	conditionType := "REQUEST"
	if i.Type != nil {
		conditionType = *i.Type
	}

	return &fastly.Condition{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
		Statement:      i.Statement,
		Type:           fastly.ToPointer(conditionType),
		Priority:       fastly.ToPointer(priority),
	}, nil
}

func createConditionError(_ *fastly.CreateConditionInput) (*fastly.Condition, error) {
	return nil, errTest
}

func deleteConditionOK(_ *fastly.DeleteConditionInput) error {
	return nil
}

func deleteConditionError(_ *fastly.DeleteConditionInput) error {
	return errTest
}

func updateConditionOK(i *fastly.UpdateConditionInput) (*fastly.Condition, error) {
	priority := 10
	if i.Priority != nil {
		priority = *i.Priority
	}

	conditionType := "REQUEST"
	if i.Type != nil {
		conditionType = *i.Type
	}

	statement := "false"
	if i.Statement != nil {
		statement = *i.Type
	}

	return &fastly.Condition{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer(i.Name),
		Statement:      fastly.ToPointer(statement),
		Type:           fastly.ToPointer(conditionType),
		Priority:       fastly.ToPointer(priority),
	}, nil
}

func updateConditionError(_ *fastly.UpdateConditionInput) (*fastly.Condition, error) {
	return nil, errTest
}

func getConditionOK(i *fastly.GetConditionInput) (*fastly.Condition, error) {
	priority := 10
	conditionType := "CACHE"
	statement := "false"

	return &fastly.Condition{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer(i.Name),
		Statement:      fastly.ToPointer(statement),
		Type:           fastly.ToPointer(conditionType),
		Priority:       fastly.ToPointer(priority),
	}, nil
}

func getConditionError(_ *fastly.GetConditionInput) (*fastly.Condition, error) {
	return nil, errTest
}

func listConditionsOK(i *fastly.ListConditionsInput) ([]*fastly.Condition, error) {
	return []*fastly.Condition{
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("always_false_request"),
			Statement:      fastly.ToPointer("false"),
			Type:           fastly.ToPointer("REQUEST"),
			Priority:       fastly.ToPointer(10),
		},
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("always_false_cache"),
			Statement:      fastly.ToPointer("false"),
			Type:           fastly.ToPointer("CACHE"),
			Priority:       fastly.ToPointer(10),
		},
	}, nil
}

func listConditionsError(_ *fastly.ListConditionsInput) ([]*fastly.Condition, error) {
	return nil, errTest
}
