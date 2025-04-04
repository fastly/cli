package honeycomb_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHoneycombCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging honeycomb create --service-id 123 --version 1 --name log --auth-token abc --dataset log --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateHoneycombFn: createHoneycombOK,
			},
			wantOutput: "Created Honeycomb logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging honeycomb create --service-id 123 --version 1 --name log --auth-token abc --dataset log --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateHoneycombFn: createHoneycombError,
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

func TestHoneycombList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging honeycomb list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsShortOutput,
		},
		{
			args: args("logging honeycomb list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging honeycomb list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging honeycomb --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging -v honeycomb list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging honeycomb list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsError,
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

func TestHoneycombDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging honeycomb describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging honeycomb describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHoneycombFn: getHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging honeycomb describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHoneycombFn: getHoneycombOK,
			},
			wantOutput: describeHoneycombOutput,
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

func TestHoneycombUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging honeycomb update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging honeycomb update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateHoneycombFn: updateHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging honeycomb update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateHoneycombFn: updateHoneycombOK,
			},
			wantOutput: "Updated Honeycomb logging endpoint log (service 123 version 4)",
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

func TestHoneycombDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging honeycomb delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging honeycomb delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteHoneycombFn: deleteHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging honeycomb delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteHoneycombFn: deleteHoneycombOK,
			},
			wantOutput: "Deleted Honeycomb logging endpoint logs (service 123 version 4)",
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

func createHoneycombOK(i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	s := fastly.Honeycomb{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
	}

	if i.Name != nil {
		s.Name = i.Name
	}

	return &s, nil
}

func createHoneycombError(_ *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func listHoneycombsOK(i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return []*fastly.Honeycomb{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			Dataset:           fastly.ToPointer("log"),
			Token:             fastly.ToPointer("tkn"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Dataset:           fastly.ToPointer("log"),
			Token:             fastly.ToPointer("tkn"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
		},
	}, nil
}

func listHoneycombsError(_ *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return nil, errTest
}

var listHoneycombsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHoneycombsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

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
`) + "\n\n"

func getHoneycombOK(i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Dataset:           fastly.ToPointer("log"),
		Token:             fastly.ToPointer("tkn"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func getHoneycombError(_ *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

var describeHoneycombOutput = "\n" + strings.TrimSpace(`
Dataset: log
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Response condition: Prevent default logging
Service ID: 123
Token: tkn
Version: 1
`) + "\n"

func updateHoneycombOK(i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Dataset:           fastly.ToPointer("log"),
		Token:             fastly.ToPointer("tkn"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func updateHoneycombError(_ *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func deleteHoneycombOK(_ *fastly.DeleteHoneycombInput) error {
	return nil
}

func deleteHoneycombError(_ *fastly.DeleteHoneycombInput) error {
	return errTest
}
