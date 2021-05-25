package sumologic_test

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

func TestSumologicCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "create", "--service-id", "123", "--version", "2", "--name", "log"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args: []string{"logging", "sumologic", "create", "--service-id", "123", "--version", "2", "--name", "log", "--url", "example.com", "--autoclone"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				CloneVersionFn:    cloneVersionOK,
				CreateSumologicFn: createSumologicOK,
			},
			wantOutput: "Created Sumologic logging endpoint log (service 123 version 3)",
		},
		{
			args: []string{"logging", "sumologic", "create", "--service-id", "123", "--version", "2", "--name", "log", "--url", "example.com"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				CloneVersionFn:    cloneVersionOK,
				CreateSumologicFn: createSumologicError,
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

func TestSumologicList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListSumologicsFn: listSumologicsOK,
			},
			wantOutput: listSumologicsShortOutput,
		},
		{
			args: []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "2", "--verbose"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListSumologicsFn: listSumologicsOK,
			},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args: []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "2", "-v"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListSumologicsFn: listSumologicsOK,
			},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args: []string{"logging", "sumologic", "--verbose", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListSumologicsFn: listSumologicsOK,
			},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "sumologic", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListSumologicsFn: listSumologicsOK,
			},
			wantOutput: listSumologicsVerboseOutput,
		},
		{
			args: []string{"logging", "sumologic", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				ListSumologicsFn: listSumologicsError,
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

func TestSumologicDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "describe", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "sumologic", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				GetSumologicFn: getSumologicError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sumologic", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				GetSumologicFn: getSumologicOK,
			},
			wantOutput: describeSumologicOutput,
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

func TestSumologicUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "2", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				CloneVersionFn:    cloneVersionOK,
				UpdateSumologicFn: updateSumologicError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sumologic", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				CloneVersionFn:    cloneVersionOK,
				UpdateSumologicFn: updateSumologicOK,
			},
			wantOutput: "Updated Sumologic logging endpoint log (service 123 version 3)",
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

func TestSumologicDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sumologic", "delete", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "sumologic", "delete", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				CloneVersionFn:    cloneVersionOK,
				DeleteSumologicFn: deleteSumologicError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sumologic", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:    listVersionsOK,
				GetVersionFn:      getVersionOK,
				CloneVersionFn:    cloneVersionOK,
				DeleteSumologicFn: deleteSumologicOK,
			},
			wantOutput: "Deleted Sumologic logging endpoint logs (service 123 version 3)",
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

func createSumologicOK(i *fastly.CreateSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createSumologicError(i *fastly.CreateSumologicInput) (*fastly.Sumologic, error) {
	return nil, errTest
}

func listSumologicsOK(i *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error) {
	return []*fastly.Sumologic{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			URL:               "example.com",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			URL:               "bar.com",
			Format:            `%h %l %u %t "%r" %>s %b`,
			ResponseCondition: "Prevent default logging",
			MessageType:       "classic",
			FormatVersion:     2,
			Placement:         "none",
		},
	}, nil
}

func listSumologicsError(i *fastly.ListSumologicsInput) ([]*fastly.Sumologic, error) {
	return nil, errTest
}

var listSumologicsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      2        logs
123      2        analytics
`) + "\n"

var listSumologicsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 2
	Sumologic 1/2
		Service ID: 123
		Version: 2
		Name: logs
		URL: example.com
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
	Sumologic 2/2
		Service ID: 123
		Version: 2
		Name: analytics
		URL: bar.com
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
`) + "\n\n"

func getSumologicOK(i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		URL:               "example.com",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getSumologicError(i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return nil, errTest
}

var describeSumologicOutput = strings.TrimSpace(`
Service ID: 123
Version: 2
Name: logs
URL: example.com
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Placement: none
`) + "\n"

func updateSumologicOK(i *fastly.UpdateSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		URL:               "example.com",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateSumologicError(i *fastly.UpdateSumologicInput) (*fastly.Sumologic, error) {
	return nil, errTest
}

func deleteSumologicOK(i *fastly.DeleteSumologicInput) error {
	return nil
}

func deleteSumologicError(i *fastly.DeleteSumologicInput) error {
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
