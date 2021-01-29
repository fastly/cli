package gcs_test

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

func TestGCSCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "foo@example.com", "--secret-key", "foo"},
			wantError: "error parsing arguments: required flag --bucket not provided",
		},
		{
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--secret-key", "foo"},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--user", "foo@example.com"},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args:       []string{"logging", "gcs", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--user", "foo@example.com", "--secret-key", "foo", "--period", "86400"},
			api:        mock.API{CreateGCSFn: createGCSOK},
			wantOutput: "Created GCS logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--user", "foo@example.com", "--secret-key", "foo", "--period", "86400"},
			api:       mock.API{CreateGCSFn: createGCSError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestGCSList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "gcs", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListGCSsFn: listGCSsOK},
			wantOutput: listGCSsShortOutput,
		},
		{
			args:       []string{"logging", "gcs", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListGCSsFn: listGCSsOK},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args:       []string{"logging", "gcs", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListGCSsFn: listGCSsOK},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args:       []string{"logging", "gcs", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListGCSsFn: listGCSsOK},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "gcs", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListGCSsFn: listGCSsOK},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args:      []string{"logging", "gcs", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListGCSsFn: listGCSsError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestGCSDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "gcs", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetGCSFn: getGCSError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "gcs", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetGCSFn: getGCSOK},
			wantOutput: describeGCSOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestGCSUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "gcs", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetGCSFn:    getGCSError,
				UpdateGCSFn: updateGCSOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "gcs", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetGCSFn:    getGCSOK,
				UpdateGCSFn: updateGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "gcs", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetGCSFn:    getGCSOK,
				UpdateGCSFn: updateGCSOK,
			},
			wantOutput: "Updated GCS logging endpoint log (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestGCSDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "gcs", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteGCSFn: deleteGCSError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "gcs", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteGCSFn: deleteGCSOK},
			wantOutput: "Deleted GCS logging endpoint logs (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createGCSOK(i *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createGCSError(i *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

func listGCSsOK(i *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return []*fastly.GCS{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Bucket:            "my-logs",
			User:              "foo@example.com",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----foo",
			Path:              "logs/",
			Period:            3600,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Bucket:            "analytics",
			User:              "foo@example.com",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----foo",
			Path:              "logs/",
			Period:            86400,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
		},
	}, nil
}

func listGCSsError(i *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return nil, errTest
}

var listGCSsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listGCSsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	GCS 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Bucket: my-logs
		User: foo@example.com
		Secret key: -----BEGIN RSA PRIVATE KEY-----foo
		Path: logs/
		Period: 3600
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
	GCS 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Bucket: analytics
		User: foo@example.com
		Secret key: -----BEGIN RSA PRIVATE KEY-----foo
		Path: logs/
		Period: 86400
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
`) + "\n\n"

func getGCSOK(i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Bucket:            "my-logs",
		User:              "foo@example.com",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----foo",
		Path:              "logs/",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

func getGCSError(i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

var describeGCSOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Bucket: my-logs
User: foo@example.com
Secret key: -----BEGIN RSA PRIVATE KEY-----foo
Path: logs/
Period: 3600
GZip level: 9
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
`) + "\n"

func updateGCSOK(i *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Bucket:            "logs",
		User:              "foo@example.com",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----foo",
		Path:              "logs/",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

func updateGCSError(i *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

func deleteGCSOK(i *fastly.DeleteGCSInput) error {
	return nil
}

func deleteGCSError(i *fastly.DeleteGCSInput) error {
	return errTest
}
