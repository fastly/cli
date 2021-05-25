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
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "2", "--name", "log", "--user", "foo@example.com", "--secret-key", "foo"},
			wantError: "error parsing arguments: required flag --bucket not provided",
		},
		{
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "2", "--name", "log", "--bucket", "log", "--secret-key", "foo"},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args:      []string{"logging", "gcs", "create", "--service-id", "123", "--version", "2", "--name", "log", "--bucket", "log", "--user", "foo@example.com"},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args: []string{"logging", "gcs", "create", "--service-id", "123", "--version", "2", "--name", "log", "--bucket", "log", "--user", "foo@example.com", "--secret-key", "foo", "--period", "86400", "--autoclone"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				CreateGCSFn:    createGCSOK,
			},
			wantOutput: "Created GCS logging endpoint log (service 123 version 3)",
		},
		{
			args: []string{"logging", "gcs", "create", "--service-id", "123", "--version", "2", "--name", "log", "--bucket", "log", "--user", "foo@example.com", "--secret-key", "foo", "--period", "86400"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				CreateGCSFn:    createGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "gcs", "create", "--service-id", "123", "--version", "2", "--name", "log", "--bucket", "log", "--user", "foo@example.com", "--secret-key", "foo", "--period", "86400", "--compression-codec", "zstd", "--gzip-level", "9"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
			},
			wantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
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

func TestGCSList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "gcs", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsShortOutput,
		},
		{
			args: []string{"logging", "gcs", "list", "--service-id", "123", "--version", "2", "--verbose"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: []string{"logging", "gcs", "list", "--service-id", "123", "--version", "2", "-v"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: []string{"logging", "gcs", "--verbose", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "gcs", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: []string{"logging", "gcs", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				ListGCSsFn:     listGCSsError,
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

func TestGCSDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "describe", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "gcs", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				GetGCSFn:       getGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "gcs", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				GetGCSFn:       getGCSOK,
			},
			wantOutput: describeGCSOutput,
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

func TestGCSUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "update", "--service-id", "123", "--version", "2", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "gcs", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				UpdateGCSFn:    updateGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "gcs", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				UpdateGCSFn:    updateGCSOK,
			},
			wantOutput: "Updated GCS logging endpoint log (service 123 version 3)",
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

func TestGCSDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "gcs", "delete", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "gcs", "delete", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				DeleteGCSFn:    deleteGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "gcs", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn: listVersionsOK,
				GetVersionFn:   getVersionOK,
				CloneVersionFn: cloneVersionOK,
				DeleteGCSFn:    deleteGCSOK,
			},
			wantOutput: "Deleted GCS logging endpoint logs (service 123 version 3)",
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
			GzipLevel:         0,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
			CompressionCodec:  "zstd",
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
			GzipLevel:         0,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
			CompressionCodec:  "zstd",
		},
	}, nil
}

func listGCSsError(i *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return nil, errTest
}

var listGCSsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      2        logs
123      2        analytics
`) + "\n"

var listGCSsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 2
	GCS 1/2
		Service ID: 123
		Version: 2
		Name: logs
		Bucket: my-logs
		User: foo@example.com
		Secret key: -----BEGIN RSA PRIVATE KEY-----foo
		Path: logs/
		Period: 3600
		GZip level: 0
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Compression codec: zstd
	GCS 2/2
		Service ID: 123
		Version: 2
		Name: analytics
		Bucket: analytics
		User: foo@example.com
		Secret key: -----BEGIN RSA PRIVATE KEY-----foo
		Path: logs/
		Period: 86400
		GZip level: 0
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Compression codec: zstd
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
		GzipLevel:         0,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
		CompressionCodec:  "zstd",
	}, nil
}

func getGCSError(i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

var describeGCSOutput = strings.TrimSpace(`
Service ID: 123
Version: 2
Name: logs
Bucket: my-logs
User: foo@example.com
Secret key: -----BEGIN RSA PRIVATE KEY-----foo
Path: logs/
Period: 3600
GZip level: 0
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
Compression codec: zstd
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
		GzipLevel:         0,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
		CompressionCodec:  "zstd",
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
