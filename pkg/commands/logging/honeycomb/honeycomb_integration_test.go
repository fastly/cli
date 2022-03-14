// NOTE: We always pass the --token flag as this allows us to side-step the
// browser based authentication flow. This is because if a token is explicitly
// provided, then we respect the user knows what they're doing.
package honeycomb_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestHoneycombCreate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging honeycomb create --service-id 123 --version 1 --name log --dataset log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --auth-token not provided",
		},
		{
			args: args("logging honeycomb create --service-id 123 --version 1 --name log --auth-token abc --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --dataset not provided",
		},
		{
			args: args("logging honeycomb create --service-id 123 --version 1 --name log --auth-token abc --dataset log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateHoneycombFn: createHoneycombOK,
			},
			wantOutput: "Created Honeycomb logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging honeycomb create --service-id 123 --version 1 --name log --auth-token abc --dataset log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				CreateHoneycombFn: createHoneycombError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestHoneycombList(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging honeycomb list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsShortOutput,
		},
		{
			args: args("logging honeycomb list --service-id 123 --version 1 --verbose --token 123"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging honeycomb list --service-id 123 --version 1 -v --token 123"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging honeycomb --verbose list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging -v honeycomb list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsOK,
			},
			wantOutput: listHoneycombsVerboseOutput,
		},
		{
			args: args("logging honeycomb list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListHoneycombsFn: listHoneycombsError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestHoneycombDescribe(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging honeycomb describe --service-id 123 --version 1 --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging honeycomb describe --service-id 123 --version 1 --name logs --token 123"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHoneycombFn: getHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging honeycomb describe --service-id 123 --version 1 --name logs --token 123"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetHoneycombFn: getHoneycombOK,
			},
			wantOutput: describeHoneycombOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestHoneycombUpdate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging honeycomb update --service-id 123 --version 1 --new-name log --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging honeycomb update --service-id 123 --version 1 --name logs --new-name log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateHoneycombFn: updateHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging honeycomb update --service-id 123 --version 1 --name logs --new-name log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				UpdateHoneycombFn: updateHoneycombOK,
			},
			wantOutput: "Updated Honeycomb logging endpoint log (service 123 version 4)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestHoneycombDelete(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging honeycomb delete --service-id 123 --version 1 --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging honeycomb delete --service-id 123 --version 1 --name logs --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteHoneycombFn: deleteHoneycombError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging honeycomb delete --service-id 123 --version 1 --name logs --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				CloneVersionFn:    testutil.CloneVersionResult(4),
				DeleteHoneycombFn: deleteHoneycombOK,
			},
			wantOutput: "Deleted Honeycomb logging endpoint logs (service 123 version 4)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.ClientFactory = mock.ClientFactory(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createHoneycombOK(i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	s := fastly.Honeycomb{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createHoneycombError(i *fastly.CreateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func listHoneycombsOK(i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return []*fastly.Honeycomb{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			Dataset:           "log",
			Token:             "tkn",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Dataset:           "log",
			Token:             "tkn",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listHoneycombsError(i *fastly.ListHoneycombsInput) ([]*fastly.Honeycomb, error) {
	return nil, errTest
}

var listHoneycombsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listHoneycombsVerboseOutput = strings.TrimSpace(`
Fastly API token provided via --token
Fastly API endpoint: https://api.fastly.com
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
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Dataset:           "log",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getHoneycombError(i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

var describeHoneycombOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Dataset: log
Token: tkn
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateHoneycombOK(i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Dataset:           "log",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateHoneycombError(i *fastly.UpdateHoneycombInput) (*fastly.Honeycomb, error) {
	return nil, errTest
}

func deleteHoneycombOK(i *fastly.DeleteHoneycombInput) error {
	return nil
}

func deleteHoneycombError(i *fastly.DeleteHoneycombInput) error {
	return errTest
}
