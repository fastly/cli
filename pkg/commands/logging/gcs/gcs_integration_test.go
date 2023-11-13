package gcs_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestGCSCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging gcs create --service-id 123 --version 1 --name log --bucket log --user foo@example.com --secret-key foo --period 86400 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateGCSFn:    createGCSOK,
			},
			wantOutput: "Created GCS logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging gcs create --service-id 123 --version 1 --name log --bucket log --user foo@example.com --secret-key foo --period 86400 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateGCSFn:    createGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging gcs create --service-id 123 --version 1 --name log --bucket log --user foo@example.com --secret-key foo --period 86400 --compression-codec zstd --gzip-level 9 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
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

func TestGCSList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging gcs list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsShortOutput,
		},
		{
			args: args("logging gcs list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: args("logging gcs list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: args("logging gcs --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: args("logging -v gcs list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			wantOutput: listGCSsVerboseOutput,
		},
		{
			args: args("logging gcs list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsError,
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

func TestGCSDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging gcs describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging gcs describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetGCSFn:       getGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging gcs describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetGCSFn:       getGCSOK,
			},
			wantOutput: describeGCSOutput,
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

func TestGCSUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging gcs update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging gcs update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateGCSFn:    updateGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging gcs update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateGCSFn:    updateGCSOK,
			},
			wantOutput: "Updated GCS logging endpoint log (service 123 version 4)",
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

func TestGCSDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging gcs delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging gcs delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteGCSFn:    deleteGCSError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging gcs delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteGCSFn:    deleteGCSOK,
			},
			wantOutput: "Deleted GCS logging endpoint logs (service 123 version 4)",
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

func createGCSOK(i *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.Name,
	}, nil
}

func createGCSError(_ *fastly.CreateGCSInput) (*fastly.GCS, error) {
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
			AccountName:       "me@fastly.com",
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
			AccountName:       "me@fastly.com",
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

func listGCSsError(_ *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return nil, errTest
}

var listGCSsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listGCSsVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	GCS 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Bucket: my-logs
		User: foo@example.com
		Account name: me@fastly.com
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
		Version: 1
		Name: analytics
		Bucket: analytics
		User: foo@example.com
		Account name: me@fastly.com
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
		AccountName:       "me@fastly.com",
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

func getGCSError(_ *fastly.GetGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

var describeGCSOutput = "\n" + strings.TrimSpace(`
Account name: me@fastly.com
Bucket: my-logs
Compression codec: zstd
Format: %h %l %u %t "%r" %>s %b
Format version: 2
GZip level: 0
Message type: classic
Name: logs
Path: logs/
Period: 3600
Placement: none
Response condition: Prevent default logging
Secret key: -----BEGIN RSA PRIVATE KEY-----foo
Service ID: 123
Timestamp format: %Y-%m-%dT%H:%M:%S.000
User: foo@example.com
Version: 1
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

func updateGCSError(_ *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

func deleteGCSOK(_ *fastly.DeleteGCSInput) error {
	return nil
}

func deleteGCSError(_ *fastly.DeleteGCSInput) error {
	return errTest
}
