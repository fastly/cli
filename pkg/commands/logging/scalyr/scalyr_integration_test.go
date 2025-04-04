package scalyr_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/app"
	fsterrs "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestScalyrCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging scalyr create --name log --service-id  --version 1 --auth-token abc --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: fsterrs.ErrNoServiceID.Error(),
		},
		{
			args: args("logging scalyr create --service-id 123 --version 1 --name log --auth-token abc --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateScalyrFn: createScalyrOK,
			},
			wantOutput: "Created Scalyr logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging scalyr create --service-id 123 --version 1 --name log --auth-token abc --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateScalyrFn: createScalyrError,
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

func TestScalyrList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging scalyr list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			wantOutput: listScalyrsShortOutput,
		},
		{
			args: args("logging scalyr list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args: args("logging scalyr list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args: args("logging scalyr --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args: args("logging -v scalyr list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsOK,
			},
			wantOutput: listScalyrsVerboseOutput,
		},
		{
			args: args("logging scalyr list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListScalyrsFn:  listScalyrsError,
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

func TestScalyrDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging scalyr describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging scalyr describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetScalyrFn:    getScalyrError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging scalyr describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetScalyrFn:    getScalyrOK,
			},
			wantOutput: describeScalyrOutput,
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

func TestScalyrUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging scalyr update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging scalyr update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateScalyrFn: updateScalyrError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging scalyr update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateScalyrFn: updateScalyrOK,
			},
			wantOutput: "Updated Scalyr logging endpoint log (service 123 version 4)",
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

func TestScalyrDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging scalyr delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging scalyr delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteScalyrFn: deleteScalyrError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging scalyr delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteScalyrFn: deleteScalyrOK,
			},
			wantOutput: "Deleted Scalyr logging endpoint logs (service 123 version 4)",
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

func createScalyrOK(i *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
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

func createScalyrError(_ *fastly.CreateScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

func listScalyrsOK(i *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
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
		},
	}, nil
}

func listScalyrsError(_ *fastly.ListScalyrsInput) ([]*fastly.Scalyr, error) {
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
`) + "\n\n"

func getScalyrOK(i *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
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
	}, nil
}

func getScalyrError(_ *fastly.GetScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

var describeScalyrOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Project ID: example-project
Region: US
Response condition: Prevent default logging
Service ID: 123
Token: abc
Version: 1
`) + "\n"

func updateScalyrOK(i *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
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

func updateScalyrError(_ *fastly.UpdateScalyrInput) (*fastly.Scalyr, error) {
	return nil, errTest
}

func deleteScalyrOK(_ *fastly.DeleteScalyrInput) error {
	return nil
}

func deleteScalyrError(_ *fastly.DeleteScalyrInput) error {
	return errTest
}
