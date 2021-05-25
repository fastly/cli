package papertrail_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestPapertrailCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "create", "--service-id", "123", "--version", "2", "--name", "log"},
			wantError: "error parsing arguments: required flag --address not provided",
		},
		{
			args: []string{"logging", "papertrail", "create", "--service-id", "123", "--version", "2", "--name", "log", "--address", "example.com:123", "--autoclone"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				CreatePapertrailFn: createPapertrailOK,
			},
			wantOutput: "Created Papertrail logging endpoint log (service 123 version 3)",
		},
		{
			args: []string{"logging", "papertrail", "create", "--service-id", "123", "--version", "2", "--name", "log", "--address", "example.com:123"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				CreatePapertrailFn: createPapertrailError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestPapertrailList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsShortOutput,
		},
		{
			args: []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "2", "--verbose"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "2", "-v"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: []string{"logging", "papertrail", "--verbose", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "papertrail", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				ListPapertrailsFn: listPapertrailsOK,
			},
			wantOutput: listPapertrailsVerboseOutput,
		},
		{
			args: []string{"logging", "papertrail", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				ListPapertrailsFn: listPapertrailsError,
			},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestPapertrailDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "describe", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "papertrail", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:  listVersionsOK,
				GetVersionFn:    getVersionOK,
				GetPapertrailFn: getPapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "papertrail", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:  listVersionsOK,
				GetVersionFn:    getVersionOK,
				GetPapertrailFn: getPapertrailOK,
			},
			wantOutput: describePapertrailOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestPapertrailUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "2", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				UpdatePapertrailFn: updatePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "papertrail", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				UpdatePapertrailFn: updatePapertrailOK,
			},
			wantOutput: "Updated Papertrail logging endpoint log (service 123 version 3)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestPapertrailDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "papertrail", "delete", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "papertrail", "delete", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				DeletePapertrailFn: deletePapertrailError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "papertrail", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				CloneVersionFn:     cloneVersionOK,
				DeletePapertrailFn: deletePapertrailOK,
			},
			wantOutput: "Deleted Papertrail logging endpoint logs (service 123 version 3)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
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
123      2        logs
123      2        analytics
`) + "\n"

var listPapertrailsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 2
	Papertrail 1/2
		Service ID: 123
		Version: 2
		Name: logs
		Address: example.com:123
		Port: 123
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Papertrail 2/2
		Service ID: 123
		Version: 2
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

var describePapertrailOutput = strings.TrimSpace(`
Service ID: 123
Version: 2
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

func listVersionsOK(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func getVersionOK(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		ServiceID: i.ServiceID,
		Number:    2,
		Active:    true,
		UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}, nil
}

func cloneVersionOK(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion + 1}, nil
}
