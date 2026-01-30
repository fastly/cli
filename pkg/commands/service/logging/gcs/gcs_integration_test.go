package gcs_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/gcs"
)

func TestGCSCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --user foo@example.com --secret-key foo --period 86400 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateGCSFn:    createGCSOK,
			},
			WantOutput: "Created GCS logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --account-name service-account-id --project-id gcp-prj-id --period 86400 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateGCSFn:    createGCSOK,
			},
			WantOutput: "Created GCS logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --user foo@example.com --secret-key foo --period 86400 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateGCSFn:    createGCSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --user foo@example.com --secret-key foo --period 86400 --compression-codec zstd --gzip-level 9 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestGCSList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			WantOutput: listGCSsShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			WantOutput: listGCSsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsOK,
			},
			WantOutput: listGCSsVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListGCSsFn:     listGCSsError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestGCSDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetGCSFn:       getGCSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetGCSFn:       getGCSOK,
			},
			WantOutput: describeGCSOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestGCSUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateGCSFn:    updateGCSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateGCSFn:    updateGCSOK,
			},
			WantOutput: "Updated GCS logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestGCSDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteGCSFn:    deleteGCSError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteGCSFn:    deleteGCSOK,
			},
			WantOutput: "Deleted GCS logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createGCSOK(_ context.Context, i *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createGCSError(_ context.Context, _ *fastly.CreateGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

func listGCSsOK(_ context.Context, i *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
	return []*fastly.GCS{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Bucket:            fastly.ToPointer("my-logs"),
			User:              fastly.ToPointer("foo@example.com"),
			AccountName:       fastly.ToPointer("me@fastly.com"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----foo"),
			Path:              fastly.ToPointer("logs/"),
			Period:            fastly.ToPointer(3600),
			GzipLevel:         fastly.ToPointer(0),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Placement:         fastly.ToPointer("none"),
			CompressionCodec:  fastly.ToPointer("zstd"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Bucket:            fastly.ToPointer("analytics"),
			User:              fastly.ToPointer("foo@example.com"),
			AccountName:       fastly.ToPointer("me@fastly.com"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----foo"),
			Path:              fastly.ToPointer("logs/"),
			Period:            fastly.ToPointer(86400),
			GzipLevel:         fastly.ToPointer(0),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Placement:         fastly.ToPointer("none"),
			CompressionCodec:  fastly.ToPointer("zstd"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listGCSsError(_ context.Context, _ *fastly.ListGCSsInput) ([]*fastly.GCS, error) {
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
		Processing region: us
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
		Processing region: us
`) + "\n\n"

func getGCSOK(_ context.Context, i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Bucket:            fastly.ToPointer("my-logs"),
		User:              fastly.ToPointer("foo@example.com"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----foo"),
		AccountName:       fastly.ToPointer("me@fastly.com"),
		Path:              fastly.ToPointer("logs/"),
		Period:            fastly.ToPointer(3600),
		GzipLevel:         fastly.ToPointer(0),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		MessageType:       fastly.ToPointer("classic"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Placement:         fastly.ToPointer("none"),
		CompressionCodec:  fastly.ToPointer("zstd"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getGCSError(_ context.Context, _ *fastly.GetGCSInput) (*fastly.GCS, error) {
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
Processing region: us
Project ID: 
Response condition: Prevent default logging
Secret key: -----BEGIN RSA PRIVATE KEY-----foo
Service ID: 123
Timestamp format: %Y-%m-%dT%H:%M:%S.000
User: foo@example.com
Version: 1
`) + "\n"

func updateGCSOK(_ context.Context, i *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Bucket:            fastly.ToPointer("logs"),
		User:              fastly.ToPointer("foo@example.com"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----foo"),
		Path:              fastly.ToPointer("logs/"),
		Period:            fastly.ToPointer(3600),
		GzipLevel:         fastly.ToPointer(0),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Placement:         fastly.ToPointer("none"),
		CompressionCodec:  fastly.ToPointer("zstd"),
	}, nil
}

func updateGCSError(_ context.Context, _ *fastly.UpdateGCSInput) (*fastly.GCS, error) {
	return nil, errTest
}

func deleteGCSOK(_ context.Context, _ *fastly.DeleteGCSInput) error {
	return nil
}

func deleteGCSError(_ context.Context, _ *fastly.DeleteGCSInput) error {
	return errTest
}
