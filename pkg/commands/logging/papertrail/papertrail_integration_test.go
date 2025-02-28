package papertrail_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestPapertrailCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging papertrail create --service-id 123 --version 1 --name log --address example.com:123 --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreatePapertrailFn: createPapertrailOK,
			},
			wantOutput: "Created Papertrail logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging papertrail create --service-id 123 --version 1 --name log --address example.com:123 --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreatePapertrailFn: createPapertrailError,
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

func TestPapertrailList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging papertrail list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsShortOutput,
		},
		{
			args: args("logging papertrail list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging papertrail list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging papertrail --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging -v papertrail list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging papertrail list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsError,
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

func TestPapertrailDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging papertrail describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging papertrail describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetPapertrailFn: getPapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging papertrail describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetPapertrailFn: getPapertrailOK,
			},
			wantOutput: describePapertrailOutput,
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

func TestPapertrailUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging papertrail update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging papertrail update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdatePapertrailFn: updatePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging papertrail update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdatePapertrailFn: updatePapertrailOK,
			},
			wantOutput: "Updated Papertrail logging endpoint log (service 123 version 4)",
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

func TestPapertrailDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging papertrail delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging papertrail delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeletePapertrailFn: deletePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging papertrail delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeletePapertrailFn: deletePapertrailOK,
			},
			wantOutput: "Deleted Papertrail logging endpoint logs (service 123 version 4)",
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

func createPapertrailOK(i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createPapertrailError(_ *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func listPapertrailsOK(i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return []*fastly.Papertrail{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Address:           fastly.ToPointer("example.com:123"),
			Port:              fastly.ToPointer(123),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Address:           fastly.ToPointer("127.0.0.1:456"),
			Port:              fastly.ToPointer(456),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
		},
	}, nil
}

func listPapertrailsError(_ *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return nil, errTest
}

var listPapertrailsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listPapertrailsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

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
	Papertrail 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: 127.0.0.1:456
		Port: 456
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
`) + "\n\n"

func getPapertrailOK(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Address:           fastly.ToPointer("example.com:123"),
		Port:              fastly.ToPointer(123),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
	}, nil
}

func getPapertrailError(_ *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

var describePapertrailOutput = "\n" + strings.TrimSpace(`
Address: example.com:123
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Port: 123
Response condition: Prevent default logging
Service ID: 123
Version: 1
`) + "\n"

func updatePapertrailOK(i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Address:           fastly.ToPointer("example.com:123"),
		Port:              fastly.ToPointer(123),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
	}, nil
}

func updatePapertrailError(_ *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func deletePapertrailOK(_ *fastly.DeletePapertrailInput) error {
	return nil
}

func deletePapertrailError(_ *fastly.DeletePapertrailInput) error {
	return errTest
}
