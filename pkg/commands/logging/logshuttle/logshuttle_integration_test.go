package logshuttle_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v11/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestLogshuttleCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging logshuttle create --service-id 123 --version 1 --name log --url example.com --auth-token abc --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateLogshuttleFn: createLogshuttleOK,
			},
			wantOutput: "Created Logshuttle logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging logshuttle create --service-id 123 --version 1 --name log --url example.com --auth-token abc --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateLogshuttleFn: createLogshuttleError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestLogshuttleList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging logshuttle list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			wantOutput: listLogshuttlesShortOutput,
		},
		{
			args: args("logging logshuttle list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args: args("logging logshuttle list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args: args("logging logshuttle --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args: args("logging -v logshuttle list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesOK,
			},
			wantOutput: listLogshuttlesVerboseOutput,
		},
		{
			args: args("logging logshuttle list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListLogshuttlesFn: listLogshuttlesError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestLogshuttleDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging logshuttle describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging logshuttle describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetLogshuttleFn: getLogshuttleError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging logshuttle describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetLogshuttleFn: getLogshuttleOK,
			},
			wantOutput: describeLogshuttleOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestLogshuttleUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging logshuttle update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging logshuttle update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateLogshuttleFn: updateLogshuttleError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging logshuttle update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateLogshuttleFn: updateLogshuttleOK,
			},
			wantOutput: "Updated Logshuttle logging endpoint log (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestLogshuttleDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging logshuttle delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging logshuttle delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteLogshuttleFn: deleteLogshuttleError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging logshuttle delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteLogshuttleFn: deleteLogshuttleOK,
			},
			wantOutput: "Deleted Logshuttle logging endpoint logs (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
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
