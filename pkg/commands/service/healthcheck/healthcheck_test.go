package healthcheck_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/healthcheck"
	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHealthCheckCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--version 1",
			WantError: "error reading service: no service ID found",
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateHealthCheckFn: createHealthCheckError,
			},
			WantError: errTest.Error(),
		},
		// NOTE: Added --timeout flag to validate that a nil pointer dereference is
		// not triggered at runtime when parsing the arguments.
		{
			Args: "--service-id 123 --version 1 --name www.test.com --autoclone --timeout 10",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateHealthCheckFn: createHealthCheckOK,
			},
			WantOutput: "Created healthcheck www.test.com (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestHealthCheckList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			WantOutput: listHealthChecksShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			WantOutput: listHealthChecksVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			WantOutput: listHealthChecksVerboseOutput,
		},
		{
			Args: "--verbose --service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			WantOutput: listHealthChecksVerboseOutput,
		},
		{
			Args: "-v --service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			WantOutput: listHealthChecksVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestHealthCheckDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				GetHealthCheckFn: getHealthCheckError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				GetHealthCheckFn: getHealthCheckOK,
			},
			WantOutput: describeHealthCheckOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestHealthCheckUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name www.test.com --comment ",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateHealthCheckFn: updateHealthCheckOK,
			},
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateHealthCheckFn: updateHealthCheckError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateHealthCheckFn: updateHealthCheckOK,
			},
			WantOutput: "Updated healthcheck www.example.com (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestHealthCheckDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      ("--service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				DeleteHealthCheckFn: deleteHealthCheckError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name www.test.com --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				DeleteHealthCheckFn: deleteHealthCheckOK,
			},
			WantOutput: "Deleted healthcheck www.test.com (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "Delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createHealthCheckOK(_ context.Context, i *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
		Host:           fastly.ToPointer("www.test.com"),
		Path:           fastly.ToPointer("/health"),
	}, nil
}

func createHealthCheckError(_ context.Context, _ *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

func listHealthChecksOK(_ context.Context, i *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return []*fastly.HealthCheck{
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("test"),
			Comment:        fastly.ToPointer("test"),
			Method:         fastly.ToPointer(http.MethodHead),
			Host:           fastly.ToPointer("www.test.com"),
			Path:           fastly.ToPointer("/health"),
		},
		{
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),
			Name:           fastly.ToPointer("example"),
			Comment:        fastly.ToPointer("example"),
			Method:         fastly.ToPointer(http.MethodHead),
			Host:           fastly.ToPointer("www.example.com"),
			Path:           fastly.ToPointer("/health"),
		},
	}, nil
}

func listHealthChecksError(_ context.Context, _ *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return nil, errTest
}

var listHealthChecksShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME     METHOD  HOST             PATH
123      1        test     HEAD    www.test.com     /health
123      1        example  HEAD    www.example.com  /health
`) + "\n"

var listHealthChecksVerboseOutput = strings.Join([]string{
	"Fastly API endpoint: https://api.fastly.com",
	"Fastly API token provided via config file (profile: user)",
	"",
	"Service ID (via --service-id): 123",
	"",
	"Version: 1",
	"	Healthcheck 1/2",
	"		Name: test",
	"		Comment: test",
	"		Method: HEAD",
	"		Host: www.test.com",
	"		Path: /health",
	"		HTTP version: ",
	"		Timeout: 0",
	"		Check interval: 0",
	"		Expected response: 0",
	"		Window: 0",
	"		Threshold: 0",
	"		Initial: 0",
	"	Healthcheck 2/2",
	"		Name: example",
	"		Comment: example",
	"		Method: HEAD",
	"		Host: www.example.com",
	"		Path: /health",
	"		HTTP version: ",
	"		Timeout: 0",
	"		Check interval: 0",
	"		Expected response: 0",
	"		Window: 0",
	"		Threshold: 0",
	"		Initial: 0",
}, "\n") + "\n\n"

func getHealthCheckOK(_ context.Context, i *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           fastly.ToPointer("test"),
		Method:         fastly.ToPointer(http.MethodHead),
		Host:           fastly.ToPointer("www.test.com"),
		Path:           fastly.ToPointer("/healthcheck"),
		Comment:        fastly.ToPointer("test"),
	}, nil
}

func getHealthCheckError(_ context.Context, _ *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

var describeHealthCheckOutput = "\n" + strings.Join([]string{
	"Service ID: 123",
	"Version: 1",
	"Name: test",
	"Comment: test",
	"Method: HEAD",
	"Host: www.test.com",
	"Path: /healthcheck",
	"HTTP version: ",
	"Timeout: 0",
	"Check interval: 0",
	"Expected response: 0",
	"Window: 0",
	"Threshold: 0",
	"Initial: 0",
}, "\n") + "\n"

func updateHealthCheckOK(_ context.Context, i *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.NewName,
	}, nil
}

func updateHealthCheckError(_ context.Context, _ *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

func deleteHealthCheckOK(_ context.Context, _ *fastly.DeleteHealthCheckInput) error {
	return nil
}

func deleteHealthCheckError(_ context.Context, _ *fastly.DeleteHealthCheckInput) error {
	return errTest
}
