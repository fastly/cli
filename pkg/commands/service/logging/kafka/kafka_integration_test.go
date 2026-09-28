package kafka_test

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
	sub "github.com/fastly/cli/pkg/commands/service/logging/kafka"
)

func TestKafkaCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --topic logs --brokers 127.0.0.1127.0.0.2 --parse-log-keyvals --max-batch-size 1024 --use-sasl --auth-method plain --username user --password password --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateKafkaFn:  createKafkaOK,
			},
			WantOutput: "Created Kafka logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --topic logs --brokers 127.0.0.1127.0.0.2 --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateKafkaFn:  createKafkaError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestKafkaList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListKafkasFn: listKafkasOK,
			},
			WantOutput: listKafkasShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListKafkasFn: listKafkasOK,
			},
			WantOutput: listKafkasVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListKafkasFn: listKafkasOK,
			},
			WantOutput: listKafkasVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListKafkasFn: listKafkasError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestKafkaDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetKafkaFn:   getKafkaError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetKafkaFn:   getKafkaOK,
			},
			WantOutput: describeKafkaOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestKafkaUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateKafkaFn:  updateKafkaError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateKafkaFn:  updateKafkaOK,
			},
			WantOutput: "Updated Kafka logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --parse-log-keyvals --max-batch-size 1024 --use-sasl --auth-method plain --username user --password password --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateKafkaFn:  updateKafkaSASL,
			},
			WantOutput: "Updated Kafka logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestKafkaDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteKafkaFn:  deleteKafkaError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteKafkaFn:  deleteKafkaOK,
			},
			WantOutput: "Deleted Kafka logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createKafkaOK(_ context.Context, i *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		ResponseCondition: new("Prevent default logging"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		Topic:             new("logs"),
		Brokers:           new("127.0.0.1,127.0.0.2"),
		RequiredACKs:      new("-1"),
		CompressionCodec:  new("zippy"),
		UseTLS:            new(true),
		Placement:         new("none"),
		TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       new("127.0.0.1,127.0.0.2"),
		TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     new(2),
		ParseLogKeyvals:   new(true),
		RequestMaxBytes:   new(1024),
		AuthMethod:        new("plain"),
		User:              new("user"),
		Password:          new("password"),
	}, nil
}

func createKafkaError(_ context.Context, _ *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

func listKafkasOK(_ context.Context, i *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return []*fastly.Kafka{
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("logs"),
			ResponseCondition: new("Prevent default logging"),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			Topic:             new("logs"),
			Brokers:           new("127.0.0.1,127.0.0.2"),
			RequiredACKs:      new("-1"),
			CompressionCodec:  new("zippy"),
			UseTLS:            new(true),
			Placement:         new("none"),
			TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       new("127.0.0.1,127.0.0.2"),
			TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
			FormatVersion:     new(2),
			ParseLogKeyvals:   new(false),
			RequestMaxBytes:   new(0),
			AuthMethod:        new("plain"),
			User:              new("user"),
			Password:          new("password"),
			ProcessingRegion:  new("us"),
		},
		{
			ServiceID:         new(i.ServiceID),
			ServiceVersion:    new(i.ServiceVersion),
			Name:              new("analytics"),
			Topic:             new("analytics"),
			Brokers:           new("127.0.0.1,127.0.0.2"),
			RequiredACKs:      new("-1"),
			CompressionCodec:  new("zippy"),
			UseTLS:            new(true),
			Placement:         new("none"),
			TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       new("127.0.0.1,127.0.0.2"),
			TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
			ResponseCondition: new("Prevent default logging"),
			Format:            new(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     new(2),
			ParseLogKeyvals:   new(false),
			RequestMaxBytes:   new(0),
			AuthMethod:        new("plain"),
			User:              new("user"),
			Password:          new("password"),
			ProcessingRegion:  new("us"),
		},
	}, nil
}

func listKafkasError(_ context.Context, _ *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return nil, errTest
}

var listKafkasShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listKafkasVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (auth: user)

Service ID (via --service-id): 123

Version: 1
	Kafka 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Topic: logs
		Brokers: 127.0.0.1,127.0.0.2
		Required acks: -1
		Compression codec: zippy
		Use TLS: true
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		TLS hostname: 127.0.0.1,127.0.0.2
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Parse log key-values: false
		Max batch size: 0
		SASL authentication method: plain
		SASL authentication username: user
		SASL authentication password: password
		Processing region: us
	Kafka 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Topic: analytics
		Brokers: 127.0.0.1,127.0.0.2
		Required acks: -1
		Compression codec: zippy
		Use TLS: true
		TLS CA certificate: -----BEGIN CERTIFICATE-----foo
		TLS client certificate: -----BEGIN CERTIFICATE-----bar
		TLS client key: -----BEGIN PRIVATE KEY-----bar
		TLS hostname: 127.0.0.1,127.0.0.2
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Placement: none
		Parse log key-values: false
		Max batch size: 0
		SASL authentication method: plain
		SASL authentication username: user
		SASL authentication password: password
		Processing region: us
  `) + "\n\n"

func getKafkaOK(_ context.Context, i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		Brokers:           new("127.0.0.1,127.0.0.2"),
		Topic:             new("logs"),
		RequiredACKs:      new("-1"),
		UseTLS:            new(true),
		CompressionCodec:  new("zippy"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     new(2),
		ResponseCondition: new("Prevent default logging"),
		Placement:         new("none"),
		ProcessingRegion:  new("us"),
		TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       new("127.0.0.1,127.0.0.2"),
		TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
	}, nil
}

func getKafkaError(_ context.Context, _ *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

var describeKafkaOutput = `
Brokers: 127.0.0.1,127.0.0.2
Compression codec: zippy
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Max batch size: 0
Name: log
Parse log key-values: false
Placement: none
Processing region: us
Required acks: -1
Response condition: Prevent default logging
SASL authentication method: ` + `
SASL authentication password: ` + `
SASL authentication username: ` + `
Service ID: 123
TLS CA certificate: -----BEGIN CERTIFICATE-----foo
TLS client certificate: -----BEGIN CERTIFICATE-----bar
TLS client key: -----BEGIN PRIVATE KEY-----bar
TLS hostname: 127.0.0.1,127.0.0.2
Topic: logs
Use TLS: true
Version: 1
`

func updateKafkaOK(_ context.Context, i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		ResponseCondition: new("Prevent default logging"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		Topic:             new("logs"),
		Brokers:           new("127.0.0.1,127.0.0.2"),
		RequiredACKs:      new("-1"),
		CompressionCodec:  new("zippy"),
		UseTLS:            new(true),
		Placement:         new("none"),
		TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       new("127.0.0.1,127.0.0.2"),
		TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     new(2),
	}, nil
}

func updateKafkaSASL(_ context.Context, i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         new(i.ServiceID),
		ServiceVersion:    new(i.ServiceVersion),
		Name:              new("log"),
		ResponseCondition: new("Prevent default logging"),
		Format:            new(`%h %l %u %t "%r" %>s %b`),
		Topic:             new("logs"),
		Brokers:           new("127.0.0.1,127.0.0.2"),
		RequiredACKs:      new("-1"),
		CompressionCodec:  new("zippy"),
		UseTLS:            new(true),
		Placement:         new("none"),
		TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       new("127.0.0.1,127.0.0.2"),
		TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     new(2),
		ParseLogKeyvals:   new(true),
		RequestMaxBytes:   new(1024),
		AuthMethod:        new("plain"),
		User:              new("user"),
		Password:          new("password"),
	}, nil
}

func updateKafkaError(_ context.Context, _ *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

func deleteKafkaOK(_ context.Context, _ *fastly.DeleteKafkaInput) error {
	return nil
}

func deleteKafkaError(_ context.Context, _ *fastly.DeleteKafkaInput) error {
	return errTest
}
