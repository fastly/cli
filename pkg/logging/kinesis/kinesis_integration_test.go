package kinesis_test

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

func TestKinesisCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--access-key", "foo", "--region", "us-east-1", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--region", "us-east-1", "--access-key", "foo", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--region", "us-east-1", "--secret-key", "bar", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: required flag --access-key not provided",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--region", "us-east-1", "--secret-key", "bar", "--iam-role", "arn:aws:iam::123456789012:role/KinesisAccess", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--region", "us-east-1", "--access-key", "foo", "--iam-role", "arn:aws:iam::123456789012:role/KinesisAccess", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--region", "us-east-1", "--access-key", "foo", "--secret-key", "bar", "--iam-role", "arn:aws:iam::123456789012:role/KinesisAccess", "--autoclone"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--access-key", "foo", "--secret-key", "bar", "--region", "us-east-1", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisOK,
			},
			wantOutput: "Created Kinesis logging endpoint log (service 123 version 4)",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log", "--stream-name", "log", "--access-key", "foo", "--secret-key", "bar", "--region", "us-east-1", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log2", "--stream-name", "log", "--region", "us-east-1", "--iam-role", "arn:aws:iam::123456789012:role/KinesisAccess", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisOK,
			},
			wantOutput: "Created Kinesis logging endpoint log2 (service 123 version 4)",
		},
		{
			args: []string{"logging", "kinesis", "create", "--service-id", "123", "--version", "1", "--name", "log2", "--stream-name", "log", "--region", "us-east-1", "--iam-role", "arn:aws:iam::123456789012:role/KinesisAccess", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisError,
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

func TestKinesisList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "kinesis", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesShortOutput,
		},
		{
			args: []string{"logging", "kinesis", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: []string{"logging", "kinesis", "list", "--service-id", "123", "--version", "1", "-v"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: []string{"logging", "kinesis", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "kinesis", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: []string{"logging", "kinesis", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesError,
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

func TestKinesisDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kinesis", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "kinesis", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetKinesisFn:   getKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "kinesis", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetKinesisFn:   getKinesisOK,
			},
			wantOutput: describeKinesisOutput,
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

func TestKinesisUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kinesis", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "kinesis", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateKinesisFn: updateKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "kinesis", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log", "--region", "us-west-1", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateKinesisFn: updateKinesisOK,
			},
			wantOutput: "Updated Kinesis logging endpoint log (service 123 version 4)",
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

func TestKinesisDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kinesis", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "kinesis", "delete", "--service-id", "123", "--version", "1", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteKinesisFn: deleteKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "kinesis", "delete", "--service-id", "123", "--version", "1", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteKinesisFn: deleteKinesisOK,
			},
			wantOutput: "Deleted Kinesis logging endpoint logs (service 123 version 4)",
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

func createKinesisOK(i *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createKinesisError(i *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

func listKinesesOK(i *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return []*fastly.Kinesis{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			StreamName:        "my-logs",
			AccessKey:         "1234",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Region:            "us-east-1",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			StreamName:        "analytics",
			AccessKey:         "1234",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Region:            "us-east-1",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			Placement:         "none",
		},
	}, nil
}

func listKinesesError(i *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return nil, errTest
}

var listKinesesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listKinesesVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Kinesis 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Stream name: my-logs
		Region: us-east-1
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
	Kinesis 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Stream name: analytics
		Region: us-east-1
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
`) + "\n\n"

func getKinesisOK(i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		StreamName:        "my-logs",
		AccessKey:         "1234",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
		Region:            "us-east-1",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func getKinesisError(i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

var describeKinesisOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Stream name: my-logs
Region: us-east-1
Access key: 1234
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Placement: none
`) + "\n"

func updateKinesisOK(i *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		StreamName:        "my-logs",
		AccessKey:         "1234",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
		Region:            "us-west-1",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}

func updateKinesisError(i *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

func deleteKinesisOK(i *fastly.DeleteKinesisInput) error {
	return nil
}

func deleteKinesisError(i *fastly.DeleteKinesisInput) error {
	return errTest
}
