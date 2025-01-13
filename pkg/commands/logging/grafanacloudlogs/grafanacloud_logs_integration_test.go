package grafanacloudlogs_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v9/fastly"
)

func TestGrafanaCloudLogsCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging grafanacloudlogs create --service-id 123 --version 1 --name log --user 123456 --url https://test123.grafana.net --auth-token testtoken --index `{\"label\": \"value\" }` --autoclone"),
			api: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				CreateGrafanaCloudLogsFn: createGrafanaCloudLogsOK,
			},
			wantOutput: "Created Grafana Cloud Logs logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging grafanacloudlogs create --service-id 123 --version 1 --name log --url https://test123.grafana.net --auth-token testtoken --index `{\"label\": \"value\" }` --autoclone"),
			api: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				CreateGrafanaCloudLogsFn: createGrafanaCloudLogsError,
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

func TestGrafanaCloudLogsList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging grafanacloudlogs list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			wantOutput: listGrafanaCloudLogsShortOutput,
		},
		{
			args: args("logging grafanacloudlogs list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			wantOutput: listGrafanaCloudLogsVerboseOutput,
		},
		{
			args: args("logging grafanacloudlogs list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			wantOutput: listGrafanaCloudLogsVerboseOutput,
		},
		{
			args: args("logging grafanacloudlogs --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			wantOutput: listGrafanaCloudLogsVerboseOutput,
		},
		{
			args: args("logging -v grafanacloudlogs list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsOK,
			},
			wantOutput: listGrafanaCloudLogsVerboseOutput,
		},
		{
			args: args("logging grafanacloudlogs list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:         testutil.ListVersions,
				ListGrafanaCloudLogsFn: listGrafanaCloudLogsError,
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

func TestGrafanaCloudLogsDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging grafanacloudlogs describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging grafanacloudlogs describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				GetGrafanaCloudLogsFn: getGrafanaCloudLogsError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging grafanacloudlogs describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				GetGrafanaCloudLogsFn: getGrafanaCloudLogsOK,
			},
			wantOutput: describeGrafanaCloudLogsOutput,
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

func TestGrafanaCloudLogsUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging grafanacloudlogs update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging grafanacloudlogs update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				UpdateGrafanaCloudLogsFn: updateGrafanaCloudLogsError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging grafanacloudlogs update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				UpdateGrafanaCloudLogsFn: updateGrafanaCloudLogsOK,
			},
			wantOutput: "Updated Grafana Cloud Logs logging endpoint log (service 123 version 4)",
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

func TestGrafanaCloudLogsDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging grafanacloudlogs delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging grafanacloudlogs delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				DeleteGrafanaCloudLogsFn: deleteGrafanaCloudLogsError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging grafanacloudlogs delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:           testutil.ListVersions,
				CloneVersionFn:           testutil.CloneVersionResult(4),
				DeleteGrafanaCloudLogsFn: deleteGrafanaCloudLogsOK,
			},
			wantOutput: "Deleted Grafana Cloud Logs logging endpoint logs (service 123 version 4)",
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

func createGrafanaCloudLogsOK(i *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return &fastly.GrafanaCloudLogs{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createGrafanaCloudLogsError(_ *fastly.CreateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

func listGrafanaCloudLogsOK(i *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error) {
	return []*fastly.GrafanaCloudLogs{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			User:              fastly.ToPointer("123456"),
			Token:             fastly.ToPointer("testtoken"),
			URL:               fastly.ToPointer("https://test123.grafana.net"),
			Index:             fastly.ToPointer("{\"label\": \"value\"}"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			User:              fastly.ToPointer("123456"),
			Token:             fastly.ToPointer("testtoken"),
			URL:               fastly.ToPointer("https://test123.grafana.net"),
			Index:             fastly.ToPointer("{\"label\": \"value\"}"),
		},
	}, nil
}

func listGrafanaCloudLogsError(_ *fastly.ListGrafanaCloudLogsInput) ([]*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

var listGrafanaCloudLogsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listGrafanaCloudLogsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	GrafanaCloudLogs 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
		User: 123456
		URL: https://test123.grafana.net
		Token: testtoken
		Index: {"label": "value"}
	GrafanaCloudLogs 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Placement: none
		User: 123456
		URL: https://test123.grafana.net
		Token: testtoken
		Index: {"label": "value"}
`) + "\n\n"

func getGrafanaCloudLogsOK(i *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return &fastly.GrafanaCloudLogs{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		MessageType:       fastly.ToPointer("classic"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		User:              fastly.ToPointer("123456"),
		URL:               fastly.ToPointer("https://test123.grafana.net"),
		Token:             fastly.ToPointer("testtoken"),
		Index:             fastly.ToPointer("{\"label\": \"value\"}"),
	}, nil
}

func getGrafanaCloudLogsError(_ *fastly.GetGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

var describeGrafanaCloudLogsOutput = "\n" + strings.TrimSpace(`
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Index: {"label": "value"}
Message type: classic
Name: logs
Placement: none
Response condition: Prevent default logging
Service ID: 123
Token: testtoken
URL: https://test123.grafana.net
User: 123456
Version: 1
`) + "\n"

func updateGrafanaCloudLogsOK(i *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return &fastly.GrafanaCloudLogs{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		Placement:         fastly.ToPointer("none"),
		User:              fastly.ToPointer("123456"),
		URL:               fastly.ToPointer("https://test123.grafana.net"),
		Token:             fastly.ToPointer("testtoken"),
		Index:             fastly.ToPointer("{\"label\": \"value\"}"),
	}, nil
}

func updateGrafanaCloudLogsError(_ *fastly.UpdateGrafanaCloudLogsInput) (*fastly.GrafanaCloudLogs, error) {
	return nil, errTest
}

func deleteGrafanaCloudLogsOK(_ *fastly.DeleteGrafanaCloudLogsInput) error {
	return nil
}

func deleteGrafanaCloudLogsError(_ *fastly.DeleteGrafanaCloudLogsInput) error {
	return errTest
}
