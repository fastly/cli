package s3_test

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
	"github.com/fastly/go-fastly/fastly"
)

func TestS3Create(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "s3", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo"},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args:       []string{"logging", "s3", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--secret-key", "bar"},
			api:        mock.API{CreateS3Fn: createS3OK},
			wantOutput: "Created S3 logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "s3", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--secret-key", "bar"},
			api:       mock.API{CreateS3Fn: createS3Error},
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

func TestS3List(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "s3", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListS3sFn: listS3sOK},
			wantOutput: listS3sShortOutput,
		},
		{
			args:       []string{"logging", "s3", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListS3sFn: listS3sOK},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args:       []string{"logging", "s3", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListS3sFn: listS3sOK},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args:       []string{"logging", "s3", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListS3sFn: listS3sOK},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "s3", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListS3sFn: listS3sOK},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args:      []string{"logging", "s3", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListS3sFn: listS3sError},
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

func TestS3Describe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "s3", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "s3", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetS3Fn: getS3Error},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "s3", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetS3Fn: getS3OK},
			wantOutput: describeS3Output,
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

func TestS3Update(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "s3", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "s3", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetS3Fn:    getS3Error,
				UpdateS3Fn: updateS3OK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "s3", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetS3Fn:    getS3OK,
				UpdateS3Fn: updateS3Error,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "s3", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetS3Fn:    getS3OK,
				UpdateS3Fn: updateS3OK,
			},
			wantOutput: "Updated S3 logging endpoint log (service 123 version 1)",
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

func TestS3Delete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "s3", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "s3", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteS3Fn: deleteS3Error},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "s3", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteS3Fn: deleteS3OK},
			wantOutput: "Deleted S3 logging endpoint logs (service 123 version 1)",
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

func createS3OK(i *fastly.CreateS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID: i.Service,
		Version:   i.Version,
		Name:      i.Name,
	}, nil
}

func createS3Error(i *fastly.CreateS3Input) (*fastly.S3, error) {
	return nil, errTest
}

func listS3sOK(i *fastly.ListS3sInput) ([]*fastly.S3, error) {
	return []*fastly.S3{
		&fastly.S3{
			ServiceID:                    i.Service,
			Version:                      i.Version,
			Name:                         "logs",
			BucketName:                   "my-logs",
			AccessKey:                    "1234",
			SecretKey:                    "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Domain:                       "https://s3.us-east-1.amazonaws.com",
			Path:                         "logs/",
			Period:                       3600,
			GzipLevel:                    9,
			Format:                       `%h %l %u %t "%r" %>s %b`,
			FormatVersion:                2,
			MessageType:                  "classic",
			ResponseCondition:            "Prevent default logging",
			TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
			Redundancy:                   "standard",
			Placement:                    "none",
			ServerSideEncryption:         "aws:kms",
			ServerSideEncryptionKMSKeyID: "1234",
		},
		&fastly.S3{
			ServiceID:                    i.Service,
			Version:                      i.Version,
			Name:                         "analytics",
			BucketName:                   "analytics",
			AccessKey:                    "1234",
			SecretKey:                    "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Domain:                       "https://s3.us-east-2.amazonaws.com",
			Path:                         "logs/",
			Period:                       86400,
			GzipLevel:                    9,
			Format:                       `%h %l %u %t "%r" %>s %b`,
			FormatVersion:                2,
			MessageType:                  "classic",
			ResponseCondition:            "Prevent default logging",
			TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
			Redundancy:                   "standard",
			Placement:                    "none",
			ServerSideEncryption:         "aws:kms",
			ServerSideEncryptionKMSKeyID: "1234",
		},
	}, nil
}

func listS3sError(i *fastly.ListS3sInput) ([]*fastly.S3, error) {
	return nil, errTest
}

var listS3sShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listS3sVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	S3 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Bucket: my-logs
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Path: logs/
		Period: 3600
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Redundancy: standard
		Server-side encryption: aws:kms
		Server-side encryption KMS key ID: aws:kms
	S3 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Bucket: analytics
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Path: logs/
		Period: 86400
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Redundancy: standard
		Server-side encryption: aws:kms
		Server-side encryption KMS key ID: aws:kms
`) + "\n\n"

func getS3OK(i *fastly.GetS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:                    i.Service,
		Version:                      i.Version,
		Name:                         "logs",
		BucketName:                   "my-logs",
		AccessKey:                    "1234",
		SecretKey:                    "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
		Domain:                       "https://s3.us-east-1.amazonaws.com",
		Path:                         "logs/",
		Period:                       3600,
		GzipLevel:                    9,
		Format:                       `%h %l %u %t "%r" %>s %b`,
		FormatVersion:                2,
		MessageType:                  "classic",
		ResponseCondition:            "Prevent default logging",
		TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
		Redundancy:                   "standard",
		Placement:                    "none",
		ServerSideEncryption:         "aws:kms",
		ServerSideEncryptionKMSKeyID: "1234",
	}, nil
}

func getS3Error(i *fastly.GetS3Input) (*fastly.S3, error) {
	return nil, errTest
}

var describeS3Output = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Bucket: my-logs
Access key: 1234
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Path: logs/
Period: 3600
GZip level: 9
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
Redundancy: standard
Server-side encryption: aws:kms
Server-side encryption KMS key ID: aws:kms
`) + "\n"

func updateS3OK(i *fastly.UpdateS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:                    i.Service,
		Version:                      i.Version,
		Name:                         "log",
		BucketName:                   "my-logs",
		AccessKey:                    "1234",
		SecretKey:                    "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
		Domain:                       "https://s3.us-east-1.amazonaws.com",
		Path:                         "logs/",
		Period:                       3600,
		GzipLevel:                    9,
		Format:                       `%h %l %u %t "%r" %>s %b`,
		FormatVersion:                2,
		MessageType:                  "classic",
		ResponseCondition:            "Prevent default logging",
		TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
		Redundancy:                   "standard",
		Placement:                    "none",
		ServerSideEncryption:         "aws:kms",
		ServerSideEncryptionKMSKeyID: "1234",
	}, nil
}

func updateS3Error(i *fastly.UpdateS3Input) (*fastly.S3, error) {
	return nil, errTest
}

func deleteS3OK(i *fastly.DeleteS3Input) error {
	return nil
}

func deleteS3Error(i *fastly.DeleteS3Input) error {
	return errTest
}
