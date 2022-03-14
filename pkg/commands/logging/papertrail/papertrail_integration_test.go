// NOTE: We always pass the --token flag as this allows us to side-step the
// browser based authentication flow. This is because if a token is explicitly
// provided, then we respect the user knows what they're doing.
package papertrail_test

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

func TestPapertrailCreate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging papertrail create --service-id 123 --version 1 --name log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --address not provided",
		},
		{
			args: args("logging papertrail create --service-id 123 --version 1 --name log --address example.com:123 --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreatePapertrailFn: createPapertrailOK,
			},
			wantOutput: "Created Papertrail logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging papertrail create --service-id 123 --version 1 --name log --address example.com:123 --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreatePapertrailFn: createPapertrailError,
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

func TestPapertrailList(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging papertrail list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsShortOutput,
		},
		{
			args: args("logging papertrail list --service-id 123 --version 1 --verbose --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging papertrail list --service-id 123 --version 1 -v --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging papertrail --verbose list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging -v papertrail list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: args("logging papertrail list --service-id 123 --version 1 --token 123"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				ListPapertrailsFn: listPapertrailsError,
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

func TestPapertrailDescribe(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging papertrail describe --service-id 123 --version 1 --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging papertrail describe --service-id 123 --version 1 --name logs --token 123"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetPapertrailFn: getPapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging papertrail describe --service-id 123 --version 1 --name logs --token 123"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetPapertrailFn: getPapertrailOK,
			},
			wantOutput: describePapertrailOutput,
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

func TestPapertrailUpdate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging papertrail update --service-id 123 --version 1 --new-name log --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging papertrail update --service-id 123 --version 1 --name logs --new-name log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdatePapertrailFn: updatePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging papertrail update --service-id 123 --version 1 --name logs --new-name log --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdatePapertrailFn: updatePapertrailOK,
			},
			wantOutput: "Updated Papertrail logging endpoint log (service 123 version 4)",
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

func TestPapertrailDelete(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging papertrail delete --service-id 123 --version 1 --token 123"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging papertrail delete --service-id 123 --version 1 --name logs --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeletePapertrailFn: deletePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging papertrail delete --service-id 123 --version 1 --name logs --autoclone --token 123"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeletePapertrailFn: deletePapertrailOK,
			},
			wantOutput: "Deleted Papertrail logging endpoint logs (service 123 version 4)",
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

func createPapertrailOK(i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createPapertrailError(i *fastly.CreatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func listPapertrailsOK(i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return []*fastly.Papertrail{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Address:           "example.com:123",
			Port:              123,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Address:           "127.0.0.1:456",
			Port:              456,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listPapertrailsError(i *fastly.ListPapertrailsInput) ([]*fastly.Papertrail, error) {
	return nil, errTest
}

var listPapertrailsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listPapertrailsVerboseOutput = strings.TrimSpace(`
Fastly API token provided via --token
Fastly API endpoint: https://api.fastly.com
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
		Placement: none
	Papertrail 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: 127.0.0.1:456
		Port: 456
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getPapertrailOK(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com:123",
		Port:              123,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getPapertrailError(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

var describePapertrailOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Address: example.com:123
Port: 123
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updatePapertrailOK(i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Address:           "example.com:123",
		Port:              123,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updatePapertrailError(i *fastly.UpdatePapertrailInput) (*fastly.Papertrail, error) {
	return nil, errTest
}

func deletePapertrailOK(i *fastly.DeletePapertrailInput) error {
	return nil
}

func deletePapertrailError(i *fastly.DeletePapertrailInput) error {
	return errTest
}
