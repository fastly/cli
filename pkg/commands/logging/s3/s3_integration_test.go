package s3_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestS3Create(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags or the --iam-role flag must be provided",
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --secret-key bar --iam-role arn:aws:iam::123456789012:role/S3Access --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --access-key foo --iam-role arn:aws:iam::123456789012:role/S3Access --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key bar --iam-role arn:aws:iam::123456789012:role/S3Access --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key bar --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateS3Fn:     createS3OK,
			},
			wantOutput: "Created S3 logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key bar --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateS3Fn:     createS3Error,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log2 --bucket log --iam-role arn:aws:iam::123456789012:role/S3Access --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateS3Fn:     createS3OK,
			},
			wantOutput: "Created S3 logging endpoint log2 (service 123 version 4)",
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log2 --bucket log --iam-role arn:aws:iam::123456789012:role/S3Access --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateS3Fn:     createS3Error,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging s3 create --service-id 123 --version 1 --name log --bucket log --iam-role arn:aws:iam::123456789012:role/S3Access --compression-codec zstd --gzip-level 9 --autoclone"),
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

func TestS3List(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging s3 list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListS3sFn:      listS3sOK,
			},
			wantOutput: listS3sShortOutput,
		},
		{
			args: args("logging s3 list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListS3sFn:      listS3sOK,
			},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args: args("logging s3 list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListS3sFn:      listS3sOK,
			},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args: args("logging s3 --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListS3sFn:      listS3sOK,
			},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args: args("logging -v s3 list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListS3sFn:      listS3sOK,
			},
			wantOutput: listS3sVerboseOutput,
		},
		{
			args: args("logging s3 list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListS3sFn:      listS3sError,
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

func TestS3Describe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging s3 describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging s3 describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetS3Fn:        getS3Error,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging s3 describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetS3Fn:        getS3OK,
			},
			wantOutput: describeS3Output,
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

func TestS3Update(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging s3 update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging s3 update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateS3Fn:     updateS3Error,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging s3 update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateS3Fn:     updateS3OK,
			},
			wantOutput: "Updated S3 logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging s3 update --service-id 123 --version 1 --name logs --access-key foo --secret-key bar --iam-role  --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateS3Fn:     updateS3OK,
			},
			wantOutput: "Updated S3 logging endpoint log (service 123 version 4)",
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

func TestS3Delete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging s3 delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging s3 delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteS3Fn:     deleteS3Error,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging s3 delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteS3Fn:     deleteS3OK,
			},
			wantOutput: "Deleted S3 logging endpoint logs (service 123 version 4)",
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

func createS3OK(i *fastly.CreateS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:        fastly.ToPointer(i.ServiceID),
		ServiceVersion:   fastly.ToPointer(i.ServiceVersion),
		Name:             i.Name,
		CompressionCodec: fastly.ToPointer("zstd"),
	}, nil
}

func createS3Error(_ *fastly.CreateS3Input) (*fastly.S3, error) {
	return nil, errTest
}

func listS3sOK(i *fastly.ListS3sInput) ([]*fastly.S3, error) {
	return []*fastly.S3{
		{
			ServiceID:                    fastly.ToPointer(i.ServiceID),
			ServiceVersion:               fastly.ToPointer(i.ServiceVersion),
			Name:                         fastly.ToPointer("logs"),
			BucketName:                   fastly.ToPointer("my-logs"),
			AccessKey:                    fastly.ToPointer("1234"),
			SecretKey:                    fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			IAMRole:                      fastly.ToPointer("xyz"),
			Domain:                       fastly.ToPointer("https://s3.us-east-1.amazonaws.com"),
			Path:                         fastly.ToPointer("logs/"),
			Period:                       fastly.ToPointer(3600),
			Format:                       fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:                fastly.ToPointer(2),
			MessageType:                  fastly.ToPointer("classic"),
			ResponseCondition:            fastly.ToPointer("Prevent default logging"),
			TimestampFormat:              fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Redundancy:                   fastly.ToPointer(fastly.S3RedundancyStandard),
			Placement:                    fastly.ToPointer("none"),
			PublicKey:                    fastly.ToPointer(pgpPublicKey()),
			ServerSideEncryption:         fastly.ToPointer(fastly.S3ServerSideEncryptionKMS),
			ServerSideEncryptionKMSKeyID: fastly.ToPointer("1234"),
			CompressionCodec:             fastly.ToPointer("zstd"),
			ProcessingRegion:             fastly.ToPointer("us"),
		},
		{
			ServiceID:                    fastly.ToPointer(i.ServiceID),
			ServiceVersion:               fastly.ToPointer(i.ServiceVersion),
			Name:                         fastly.ToPointer("analytics"),
			BucketName:                   fastly.ToPointer("analytics"),
			AccessKey:                    fastly.ToPointer("1234"),
			SecretKey:                    fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Domain:                       fastly.ToPointer("https://s3.us-east-2.amazonaws.com"),
			Path:                         fastly.ToPointer("logs/"),
			Period:                       fastly.ToPointer(86400),
			Format:                       fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:                fastly.ToPointer(2),
			MessageType:                  fastly.ToPointer("classic"),
			ResponseCondition:            fastly.ToPointer("Prevent default logging"),
			TimestampFormat:              fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Redundancy:                   fastly.ToPointer(fastly.S3RedundancyStandard),
			Placement:                    fastly.ToPointer("none"),
			PublicKey:                    fastly.ToPointer(pgpPublicKey()),
			ServerSideEncryption:         fastly.ToPointer(fastly.S3ServerSideEncryptionKMS),
			ServerSideEncryptionKMSKeyID: fastly.ToPointer("1234"),
			FileMaxBytes:                 fastly.ToPointer(12345),
			CompressionCodec:             fastly.ToPointer("zstd"),
			ProcessingRegion:             fastly.ToPointer("us"),
		},
	}, nil
}

func listS3sError(_ *fastly.ListS3sInput) ([]*fastly.S3, error) {
	return nil, errTest
}

var listS3sShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listS3sVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	S3 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Bucket: my-logs
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		IAM role: xyz
		Path: logs/
		Period: 3600
		GZip level: 0
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Public key: `+pgpPublicKey()+`
		Redundancy: standard
		Server-side encryption: aws:kms
		Server-side encryption KMS key ID: aws:kms
		File max bytes: 0
		Compression codec: zstd
		Processing region: us
	S3 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Bucket: analytics
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
		Path: logs/
		Period: 86400
		GZip level: 0
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Public key: `+pgpPublicKey()+`
		Redundancy: standard
		Server-side encryption: aws:kms
		Server-side encryption KMS key ID: aws:kms
		File max bytes: 12345
		Compression codec: zstd
		Processing region: us
`) + "\n\n"

func getS3OK(i *fastly.GetS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:                    fastly.ToPointer(i.ServiceID),
		ServiceVersion:               fastly.ToPointer(i.ServiceVersion),
		Name:                         fastly.ToPointer("logs"),
		BucketName:                   fastly.ToPointer("my-logs"),
		AccessKey:                    fastly.ToPointer("1234"),
		SecretKey:                    fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Domain:                       fastly.ToPointer("https://s3.us-east-1.amazonaws.com"),
		Path:                         fastly.ToPointer("logs/"),
		Period:                       fastly.ToPointer(3600),
		Format:                       fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:                fastly.ToPointer(2),
		MessageType:                  fastly.ToPointer("classic"),
		ResponseCondition:            fastly.ToPointer("Prevent default logging"),
		TimestampFormat:              fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Redundancy:                   fastly.ToPointer(fastly.S3RedundancyStandard),
		Placement:                    fastly.ToPointer("none"),
		PublicKey:                    fastly.ToPointer(pgpPublicKey()),
		ServerSideEncryption:         fastly.ToPointer(fastly.S3ServerSideEncryptionKMS),
		ServerSideEncryptionKMSKeyID: fastly.ToPointer("1234"),
		CompressionCodec:             fastly.ToPointer("zstd"),
		ProcessingRegion:             fastly.ToPointer("us"),
	}, nil
}

func getS3Error(_ *fastly.GetS3Input) (*fastly.S3, error) {
	return nil, errTest
}

var describeS3Output = "\n" + strings.TrimSpace(`
Access key: 1234
Bucket: my-logs
Compression codec: zstd
File max bytes: 0
Format: %h %l %u %t "%r" %>s %b
Format version: 2
GZip level: 0
Message type: classic
Name: logs
Path: logs/
Period: 3600
Placement: none
Processing region: us
Public key: `+pgpPublicKey()+`
Redundancy: standard
Response condition: Prevent default logging
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Server-side encryption: aws:kms
Server-side encryption KMS key ID: aws:kms
Service ID: 123
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Version: 1
`) + "\n"

func updateS3OK(i *fastly.UpdateS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:                    fastly.ToPointer(i.ServiceID),
		ServiceVersion:               fastly.ToPointer(i.ServiceVersion),
		Name:                         fastly.ToPointer("log"),
		BucketName:                   fastly.ToPointer("my-logs"),
		AccessKey:                    fastly.ToPointer("1234"),
		SecretKey:                    fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Domain:                       fastly.ToPointer("https://s3.us-east-1.amazonaws.com"),
		Path:                         fastly.ToPointer("logs/"),
		Period:                       fastly.ToPointer(3600),
		Format:                       fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:                fastly.ToPointer(2),
		MessageType:                  fastly.ToPointer("classic"),
		ResponseCondition:            fastly.ToPointer("Prevent default logging"),
		TimestampFormat:              fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Redundancy:                   fastly.ToPointer(fastly.S3RedundancyStandard),
		Placement:                    fastly.ToPointer("none"),
		PublicKey:                    fastly.ToPointer(pgpPublicKey()),
		ServerSideEncryption:         fastly.ToPointer(fastly.S3ServerSideEncryptionKMS),
		ServerSideEncryptionKMSKeyID: fastly.ToPointer("1234"),
		CompressionCodec:             fastly.ToPointer("zstd"),
	}, nil
}

func updateS3Error(_ *fastly.UpdateS3Input) (*fastly.S3, error) {
	return nil, errTest
}

func deleteS3OK(_ *fastly.DeleteS3Input) error {
	return nil
}

func deleteS3Error(_ *fastly.DeleteS3Input) error {
	return errTest
}

// pgpPublicKey returns a PEM encoded PGP public key suitable for testing.
func pgpPublicKey() string {
	return strings.TrimSpace(`-----BEGIN PGP PUBLIC KEY BLOCK-----
mQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/
ibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4
8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p
lDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn
dwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB
89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz
dCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6
vFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc
9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9
OLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX
SvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq
7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx
kATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG
M1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe
u6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L
4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF
ftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K
UEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu
YrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi
kiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb
DAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml
dYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L
3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c
FaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR
5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR
wMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N
=28dr
-----END PGP PUBLIC KEY BLOCK-----
`)
}
