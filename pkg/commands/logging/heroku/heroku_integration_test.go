package heroku_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestHerokuCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging heroku create --service-id 123 --version 1 --name log --auth-token abc --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateHerokuFn: createHerokuOK,
			},
			wantOutput: "Created Heroku logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging heroku create --service-id 123 --version 1 --name log --auth-token abc --url example.com --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateHerokuFn: createHerokuError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestHerokuList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging heroku list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHerokusFn:  listHerokusOK,
			},
			wantOutput: listHerokusShortOutput,
		},
		{
			args: args("logging heroku list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHerokusFn:  listHerokusOK,
			},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args: args("logging heroku list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHerokusFn:  listHerokusOK,
			},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args: args("logging heroku --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHerokusFn:  listHerokusOK,
			},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args: args("logging -v heroku list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHerokusFn:  listHerokusOK,
			},
			wantOutput: listHerokusVerboseOutput,
		},
		{
			args: args("logging heroku list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListHerokusFn:  listHerokusError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestHerokuDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging heroku describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging heroku describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHerokuFn:    getHerokuError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging heroku describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHerokuFn:    getHerokuOK,
			},
			wantOutput: describeHerokuOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestHerokuUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging heroku update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging heroku update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateHerokuFn: updateHerokuError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging heroku update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateHerokuFn: updateHerokuOK,
			},
			wantOutput: "Updated Heroku logging endpoint log (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestHerokuDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging heroku delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging heroku delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteHerokuFn: deleteHerokuError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging heroku delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteHerokuFn: deleteHerokuOK,
			},
			wantOutput: "Deleted Heroku logging endpoint logs (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createHerokuOK(i *fastly.CreateHerokuInput) (*fastly.Heroku, error) {
	s := fastly.Heroku{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if *i.Name != "" {
		s.Name = *i.Name
	}

	return &s, nil
}

func createHerokuError(i *fastly.CreateHerokuInput) (*fastly.Heroku, error) {
	return nil, errTest
}

func listHerokusOK(i *fastly.ListHerokusInput) ([]*fastly.Heroku, error) {
	return []*fastly.Heroku{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			URL:               "example.com",
			Token:             "abc",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			URL:               "bar.com",
			Token:             "abc",
			Format:            `%h %l %u %t "%r" %>s %b`,
			ResponseCondition: "Prevent default logging",
			FormatVersion:     2,
			Placement:         "none",
		},
	}, nil
}

func listHerokusError(i *fastly.ListHerokusInput) ([]*fastly.Heroku, error) {
	return nil, errTest
}

var listHerokusShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHerokusVerboseOutput = strings.TrimSpace(`
Fastly API token provided via config file (profile: user)
Fastly API endpoint: https://api.fastly.com

Service ID (via --service-id): 123

Version: 1
	Heroku 1/2
		Service ID: 123
		Version: 1
		Name: logs
		URL: example.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Heroku 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		URL: bar.com
		Token: abc
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getHerokuOK(i *fastly.GetHerokuInput) (*fastly.Heroku, error) {
	return &fastly.Heroku{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		URL:               "example.com",
		Token:             "abc",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getHerokuError(i *fastly.GetHerokuInput) (*fastly.Heroku, error) {
	return nil, errTest
}

var describeHerokuOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Response condition: Prevent default logging
Service ID: 123
Token: abc
URL: example.com
Version: 1
`) + "\n"

func updateHerokuOK(i *fastly.UpdateHerokuInput) (*fastly.Heroku, error) {
	return &fastly.Heroku{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		URL:               "example.com",
		Token:             "abc",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateHerokuError(i *fastly.UpdateHerokuInput) (*fastly.Heroku, error) {
	return nil, errTest
}

func deleteHerokuOK(i *fastly.DeleteHerokuInput) error {
	return nil
}

func deleteHerokuError(i *fastly.DeleteHerokuInput) error {
	return errTest
}
