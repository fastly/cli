package kinesis_test

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

func TestKinesisCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log --stream-name log --region us-east-1 --secret-key bar --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log --stream-name log --region us-east-1 --access-key foo --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log --stream-name log --region us-east-1 --access-key foo --secret-key bar --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log --stream-name log --access-key foo --secret-key bar --region us-east-1 --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisOK,
			},
			wantOutput: "Created Kinesis logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log --stream-name log --access-key foo --secret-key bar --region us-east-1 --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log2 --stream-name log --region us-east-1 --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisOK,
			},
			wantOutput: "Created Kinesis logging endpoint log2 (service 123 version 4)",
		},
		{
			args: args("logging kinesis create --service-id 123 --version 1 --name log2 --stream-name log --region us-east-1 --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisError,
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

func TestKinesisList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging kinesis list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesShortOutput,
		},
		{
			args: args("logging kinesis list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: args("logging kinesis list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: args("logging kinesis --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: args("logging -v kinesis list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesOK,
			},
			wantOutput: listKinesesVerboseOutput,
		},
		{
			args: args("logging kinesis list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKinesisFn:  listKinesesError,
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

func TestKinesisDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging kinesis describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging kinesis describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetKinesisFn:   getKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kinesis describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetKinesisFn:   getKinesisOK,
			},
			wantOutput: describeKinesisOutput,
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

func TestKinesisUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging kinesis update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging kinesis update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateKinesisFn: updateKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kinesis update --service-id 123 --version 1 --name logs --new-name log --region us-west-1 --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateKinesisFn: updateKinesisOK,
			},
			wantOutput: "Updated Kinesis logging endpoint log (service 123 version 4)",
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

func TestKinesisDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging kinesis delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging kinesis delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteKinesisFn: deleteKinesisError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kinesis delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteKinesisFn: deleteKinesisOK,
			},
			wantOutput: "Deleted Kinesis logging endpoint logs (service 123 version 4)",
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

func createKinesisOK(i *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createKinesisError(_ *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

func listKinesesOK(i *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return []*fastly.Kinesis{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			StreamName:        fastly.ToPointer("my-logs"),
			AccessKey:         fastly.ToPointer("1234"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Region:            fastly.ToPointer("us-east-1"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			StreamName:        fastly.ToPointer("analytics"),
			AccessKey:         fastly.ToPointer("1234"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Region:            fastly.ToPointer("us-east-1"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Placement:         fastly.ToPointer("none"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listKinesesError(_ *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return nil, errTest
}

var listKinesesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listKinesesVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

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
		Processing region: us
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
		Processing region: us
`) + "\n\n"

func getKinesisOK(i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		StreamName:        fastly.ToPointer("my-logs"),
		AccessKey:         fastly.ToPointer("1234"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Region:            fastly.ToPointer("us-east-1"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getKinesisError(_ *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

var describeKinesisOutput = "\n" + strings.TrimSpace(`
Access key: 1234
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Name: logs
Placement: none
Processing region: us
Region: us-east-1
Response condition: Prevent default logging
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Service ID: 123
Stream name: my-logs
Version: 1
`) + "\n"

func updateKinesisOK(i *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		StreamName:        fastly.ToPointer("my-logs"),
		AccessKey:         fastly.ToPointer("1234"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Region:            fastly.ToPointer("us-west-1"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Placement:         fastly.ToPointer("none"),
	}, nil
}

func updateKinesisError(_ *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

func deleteKinesisOK(_ *fastly.DeleteKinesisInput) error {
	return nil
}

func deleteKinesisError(_ *fastly.DeleteKinesisInput) error {
	return errTest
}
