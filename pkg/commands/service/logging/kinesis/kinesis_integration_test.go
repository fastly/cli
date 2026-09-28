package kinesis_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/kinesis"
)

func TestKinesisCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --stream-name log --region us-east-1 --secret-key bar --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			Args: "--service-id 123 --version 1 --name log --stream-name log --region us-east-1 --access-key foo --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			Args: "--service-id 123 --version 1 --name log --stream-name log --region us-east-1 --access-key foo --secret-key bar --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag",
		},
		{
			Args: "--service-id 123 --version 1 --name log --stream-name log --access-key foo --secret-key bar --region us-east-1 --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisOK,
			},
			WantOutput: "Created Kinesis logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --stream-name log --access-key foo --secret-key bar --region us-east-1 --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name log2 --stream-name log --region us-east-1 --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisOK,
			},
			WantOutput: "Created Kinesis logging endpoint log2 (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log2 --stream-name log --region us-east-1 --iam-role arn:aws:iam::123456789012:role/KinesisAccess --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateKinesisFn: createKinesisError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestKinesisList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:  testutil.GetVersion,
				ListKinesisFn: listKinesesOK,
			},
			WantOutput: listKinesesShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: &mock.API{
				GetVersionFn:  testutil.GetVersion,
				ListKinesisFn: listKinesesOK,
			},
			WantOutput: listKinesesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: &mock.API{
				GetVersionFn:  testutil.GetVersion,
				ListKinesisFn: listKinesesOK,
			},
			WantOutput: listKinesesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn:  testutil.GetVersion,
				ListKinesisFn: listKinesesError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestKinesisDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetKinesisFn: getKinesisError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetKinesisFn: getKinesisOK,
			},
			WantOutput: describeKinesisOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestKinesisUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateKinesisFn: updateKinesisError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --region us-west-1 --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				UpdateKinesisFn: updateKinesisOK,
			},
			WantOutput: "Updated Kinesis logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestKinesisDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteKinesisFn: deleteKinesisError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:    testutil.GetVersion,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteKinesisFn: deleteKinesisOK,
			},
			WantOutput: "Deleted Kinesis logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createKinesisOK(_ context.Context, i *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:      new(i.ServiceID),
		ServiceVersion: new(i.ServiceVersion),
		Name:           i.Name,
	}, nil
}

func createKinesisError(_ context.Context, _ *fastly.CreateKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

func listKinesesOK(_ context.Context, i *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return []*fastly.Kinesis{
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("logs"),
			StreamName:        new("my-logs"),
			AccessKey:         new("1234"),
			SecretKey:         new("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Region:            new("us-east-1"),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			ResponseCondition: new("Prevent default logging"),
			Placement:         new("none"),
			ProcessingRegion:  new("us"),
		},
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("analytics"),
			StreamName:        new("analytics"),
			AccessKey:         new("1234"),
			SecretKey:         new("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Region:            new("us-east-1"),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			ResponseCondition: new("Prevent default logging"),
			Placement:         new("none"),
			ProcessingRegion:  new("us"),
		},
	}, nil
}

func listKinesesError(_ context.Context, _ *fastly.ListKinesisInput) ([]*fastly.Kinesis, error) {
	return nil, errTest
}

var listKinesesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listKinesesVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (auth: user)

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

func getKinesisOK(_ context.Context, i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("logs"),
		StreamName:        new("my-logs"),
		AccessKey:         new("1234"),
		SecretKey:         new("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Region:            new("us-east-1"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
		ProcessingRegion:  new("us"),
	}, nil
}

func getKinesisError(_ context.Context, _ *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
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

func updateKinesisOK(_ context.Context, i *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		StreamName:        new("my-logs"),
		AccessKey:         new("1234"),
		SecretKey:         new("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Region:            new("us-west-1"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
	}, nil
}

func updateKinesisError(_ context.Context, _ *fastly.UpdateKinesisInput) (*fastly.Kinesis, error) {
	return nil, errTest
}

func deleteKinesisOK(_ context.Context, _ *fastly.DeleteKinesisInput) error {
	return nil
}

func deleteKinesisError(_ context.Context, _ *fastly.DeleteKinesisInput) error {
	return errTest
}
