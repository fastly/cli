package healthcheck_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHealthCheckCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("healthcheck create --version 1"),
			wantError: "error reading service: no service ID found",
		},
		{
			args: args("healthcheck create --service-id 123 --version 1 --name www.test.com --autoclone"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateHealthCheckFn: createHealthCheckError,
			},
			wantError: errTest.Error(),
		},
		// NOTE: Added --timeout flag to validate that a nil pointer dereference is
		// not triggered at runtime when parsing the arguments.
		{
			args: args("healthcheck create --service-id 123 --version 1 --name www.test.com --autoclone --timeout 10"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateHealthCheckFn: createHealthCheckOK,
			},
			wantOutput: "Created healthcheck www.test.com (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestHealthCheckList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("healthcheck list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			wantOutput: listHealthChecksShortOutput,
		},
		{
			args: args("healthcheck list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args: args("healthcheck list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args: args("healthcheck --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args: args("-v healthcheck list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksOK,
			},
			wantOutput: listHealthChecksVerboseOutput,
		},
		{
			args: args("healthcheck list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListHealthChecksFn: listHealthChecksError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestHealthCheckDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("healthcheck describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("healthcheck describe --service-id 123 --version 1 --name www.test.com"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				GetHealthCheckFn: getHealthCheckError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("healthcheck describe --service-id 123 --version 1 --name www.test.com"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				GetHealthCheckFn: getHealthCheckOK,
			},
			wantOutput: describeHealthCheckOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestHealthCheckUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("healthcheck update --service-id 123 --version 1 --new-name www.test.com --comment "),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("healthcheck update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateHealthCheckFn: updateHealthCheckOK,
			},
		},
		{
			args: args("healthcheck update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateHealthCheckFn: updateHealthCheckError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("healthcheck update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateHealthCheckFn: updateHealthCheckOK,
			},
			wantOutput: "Updated healthcheck www.example.com (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestHealthCheckDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("healthcheck delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("healthcheck delete --service-id 123 --version 1 --name www.test.com --autoclone"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				DeleteHealthCheckFn: deleteHealthCheckError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("healthcheck delete --service-id 123 --version 1 --name www.test.com --autoclone"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				DeleteHealthCheckFn: deleteHealthCheckOK,
			},
			wantOutput: "Deleted healthcheck www.test.com (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (app.RunOpts, error) {
				opts := testutil.NewRunOpts(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createHealthCheckOK(i *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
		Host:           "www.test.com",
		Path:           "/health",
	}, nil
}

func createHealthCheckError(_ *fastly.CreateHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

func listHealthChecksOK(i *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
	return []*fastly.HealthCheck{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "test",
			Comment:        "test",
			Method:         "HEAD",
			Host:           "www.test.com",
			Path:           "/health",
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "example",
			Comment:        "example",
			Method:         "HEAD",
			Host:           "www.example.com",
			Path:           "/health",
		},
	}, nil
}

func listHealthChecksError(_ *fastly.ListHealthChecksInput) ([]*fastly.HealthCheck, error) {
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

func getHealthCheckOK(i *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           "test",
		Method:         "HEAD",
		Host:           "www.test.com",
		Path:           "/healthcheck",
		Comment:        "test",
	}, nil
}

func getHealthCheckError(_ *fastly.GetHealthCheckInput) (*fastly.HealthCheck, error) {
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

func updateHealthCheckOK(i *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return &fastly.HealthCheck{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.NewName,
	}, nil
}

func updateHealthCheckError(_ *fastly.UpdateHealthCheckInput) (*fastly.HealthCheck, error) {
	return nil, errTest
}

func deleteHealthCheckOK(_ *fastly.DeleteHealthCheckInput) error {
	return nil
}

func deleteHealthCheckError(_ *fastly.DeleteHealthCheckInput) error {
	return errTest
}
