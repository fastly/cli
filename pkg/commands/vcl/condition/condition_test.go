package condition_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v8/fastly"
)

func TestConditionCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("vcl condition create --version 1"),
			WantError: "error reading service: no service ID found",
		},
		{
			Args: args("vcl condition create --service-id 123 --version 1 --name always_false --statement false --type request --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateConditionFn: createConditionOK,
			},
			WantOutput: "Created condition always_false (service 123 version 4)",
		},
		{
			Args: args("vcl condition create --service-id 123 --version 1 --name always_false --statement false --type request --priority 10 --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateConditionFn: createConditionError,
			},
			WantError: errTest.Error(),
		},
	}

	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestConditionDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("vcl condition delete --service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("vcl condition delete --service-id 123 --version 1 --name always_false --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteConditionFn: deleteConditionError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("vcl condition delete --service-id 123 --version 1 --name always_false --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteConditionFn: deleteConditionOK,
			},
			WantOutput: "Deleted condition always_false (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestConditionUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("vcl condition update --service-id 123 --version 1 --new-name false_always --comment "),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("vcl condition update --service-id 123 --version 1 --name always_false --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateConditionFn: updateConditionOK,
			},
			WantError: "error parsing arguments: must provide either --new-name, --statement, --type or --priority to update condition",
		},
		{
			Args: args("vcl condition update --service-id 123 --version 1 --name always_false --new-name false_always --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateConditionFn: updateConditionError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("vcl condition update --service-id 123 --version 1 --name always_false --new-name false_always --autoclone"),
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateConditionFn: updateConditionOK,
			},
			WantOutput: "Updated condition false_always (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestConditionDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("vcl condition describe --service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("vcl condition describe --service-id 123 --version 1 --name always_false"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetConditionFn: getConditionError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("vcl condition describe --service-id 123 --version 1 --name always_false"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetConditionFn: getConditionOK,
			},
			WantOutput: describeConditionOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestConditionList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args: args("vcl condition list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsShortOutput,
		},
		{
			Args: args("vcl condition list --service-id 123 --version 1 --verbose"),
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: args("vcl condition list --service-id 123 --version 1 -v"),
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: args("vcl condition --verbose list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: args("-v vcl condition list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsOK,
			},
			WantOutput: listConditionsVerboseOutput,
		},
		{
			Args: args("vcl condition list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListConditionsFn: listConditionsError,
			},
			WantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
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
123      1        always_false_request  false      request  10
123      1        always_false_cache    false      cache    10
`) + "\n"

var listConditionsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com

Service ID (via --service-id): 123

Version: 1
	Condition 1/2
		Name: always_false_request
		Statement: false
		Type: request
		Priority: 10
	Condition 2/2
		Name: always_false_cache
		Statement: false
		Type: cache
		Priority: 10
`) + "\n\n"

var errTest = errors.New("fixture error")

func createConditionOK(i *fastly.CreateConditionInput) (*fastly.Condition, error) {
	var priority int = 10
	if i.Priority != nil {
		priority = *i.Priority
	}

	var conditionType string = "request"
	if i.Type != nil {
		conditionType = *i.Type
	}

	return &fastly.Condition{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
		Statement:      *i.Statement,
		Type:           conditionType,
		Priority:       priority,
	}, nil
}

func createConditionError(i *fastly.CreateConditionInput) (*fastly.Condition, error) {
	return nil, errTest
}

func deleteConditionOK(i *fastly.DeleteConditionInput) error {
	return nil
}

func deleteConditionError(i *fastly.DeleteConditionInput) error {
	return errTest
}

func updateConditionOK(i *fastly.UpdateConditionInput) (*fastly.Condition, error) {
	var priority int = 10
	if i.Priority != nil {
		priority = *i.Priority
	}

	var conditionType string = "request"
	if i.Type != nil {
		conditionType = *i.Type
	}

	var statement string = "false"
	if i.Statement != nil {
		statement = *i.Type
	}

	return &fastly.Condition{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Statement:      statement,
		Type:           conditionType,
		Priority:       priority,
	}, nil
}

func updateConditionError(i *fastly.UpdateConditionInput) (*fastly.Condition, error) {
	return nil, errTest
}

func getConditionOK(i *fastly.GetConditionInput) (*fastly.Condition, error) {
	var priority int = 10
	var conditionType string = "CACHE"
	var statement string = "false"

	return &fastly.Condition{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Statement:      statement,
		Type:           conditionType,
		Priority:       priority,
	}, nil
}

func getConditionError(i *fastly.GetConditionInput) (*fastly.Condition, error) {
	return nil, errTest
}

func listConditionsOK(i *fastly.ListConditionsInput) ([]*fastly.Condition, error) {
	return []*fastly.Condition{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "always_false_request",
			Statement:      "false",
			Type:           "request",
			Priority:       10,
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "always_false_cache",
			Statement:      "false",
			Type:           "cache",
			Priority:       10,
		},
	}, nil
}

func listConditionsError(i *fastly.ListConditionsInput) ([]*fastly.Condition, error) {
	return nil, errTest
}
