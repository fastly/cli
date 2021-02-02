package kafka_test

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

func TestKafkaCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kafka", "create", "--service-id", "123", "--version", "1", "--name", "log", "--brokers", "127.0.0.1,127.0.0.2"},
			wantError: "error parsing arguments: required flag --topic not provided",
		},
		{
			args:      []string{"logging", "kafka", "create", "--service-id", "123", "--version", "1", "--name", "log", "--topic", "logs"},
			wantError: "error parsing arguments: required flag --brokers not provided",
		},
		{
			args:       []string{"logging", "kafka", "create", "--service-id", "123", "--version", "1", "--name", "log", "--topic", "logs", "--brokers", "127.0.0.1,127.0.0.2", "--parse-log-keyvals", "--max-batch-size", "1024", "--use-sasl", "--auth-method", "plain", "--username", "user", "--password", "password"},
			api:        mock.API{CreateKafkaFn: createKafkaOK},
			wantOutput: "Created Kafka logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "kafka", "create", "--service-id", "123", "--version", "1", "--name", "log", "--topic", "logs", "--brokers", "127.0.0.1,127.0.0.2"},
			api:       mock.API{CreateKafkaFn: createKafkaError},
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

func TestKafkaList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "kafka", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListKafkasFn: listKafkasOK},
			wantOutput: listKafkasShortOutput,
		},
		{
			args:       []string{"logging", "kafka", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListKafkasFn: listKafkasOK},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args:       []string{"logging", "kafka", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListKafkasFn: listKafkasOK},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args:       []string{"logging", "kafka", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListKafkasFn: listKafkasOK},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "kafka", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListKafkasFn: listKafkasOK},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args:      []string{"logging", "kafka", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListKafkasFn: listKafkasError},
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

func TestKafkaDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kafka", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "kafka", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetKafkaFn: getKafkaError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "kafka", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetKafkaFn: getKafkaOK},
			wantOutput: describeKafkaOutput,
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

func TestKafkaUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kafka", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "kafka", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetKafkaFn:    getKafkaError,
				UpdateKafkaFn: updateKafkaOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "kafka", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetKafkaFn:    getKafkaOK,
				UpdateKafkaFn: updateKafkaError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "kafka", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetKafkaFn:    getKafkaOK,
				UpdateKafkaFn: updateKafkaOK,
			},
			wantOutput: "Updated Kafka logging endpoint log (service 123 version 1)",
		},
		{
			args: []string{"logging", "kafka", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log", "--parse-log-keyvals", "--max-batch-size", "1024", "--use-sasl", "--auth-method", "plain", "--username", "user", "--password", "password"},
			api: mock.API{
				GetKafkaFn:    getKafkaOK,
				UpdateKafkaFn: updateKafkaSASL,
			},
			wantOutput: "Updated Kafka logging endpoint log (service 123 version 1)",
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

func TestKafkaDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "kafka", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "kafka", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteKafkaFn: deleteKafkaError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "kafka", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteKafkaFn: deleteKafkaOK},
			wantOutput: "Deleted Kafka logging endpoint logs (service 123 version 1)",
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

func createKafkaOK(i *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Topic:             "logs",
		Brokers:           "127.0.0.1,127.0.0.2",
		RequiredACKs:      "-1",
		CompressionCodec:  "zippy",
		UseTLS:            true,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "127.0.0.1,127.0.0.2",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
		ParseLogKeyvals:   true,
		RequestMaxBytes:   1024,
		AuthMethod:        "plain",
		User:              "user",
		Password:          "password",
	}, nil
}

func createKafkaError(i *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

func listKafkasOK(i *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return []*fastly.Kafka{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			Topic:             "logs",
			Brokers:           "127.0.0.1,127.0.0.2",
			RequiredACKs:      "-1",
			CompressionCodec:  "zippy",
			UseTLS:            true,
			Placement:         "none",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "127.0.0.1,127.0.0.2",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			FormatVersion:     2,
			ParseLogKeyvals:   false,
			RequestMaxBytes:   0,
			AuthMethod:        "",
			User:              "",
			Password:          "",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Topic:             "analytics",
			Brokers:           "127.0.0.1,127.0.0.2",
			RequiredACKs:      "-1",
			CompressionCodec:  "zippy",
			UseTLS:            true,
			Placement:         "none",
			TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
			TLSHostname:       "127.0.0.1,127.0.0.2",
			TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
			TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			ResponseCondition: "Prevent default logging",
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ParseLogKeyvals:   false,
			RequestMaxBytes:   0,
			AuthMethod:        "",
			User:              "",
			Password:          "",
		},
	}, nil
}

func listKafkasError(i *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return nil, errTest
}

var listKafkasShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listKafkasVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
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
		SASL authentication method: 
		SASL authentication username: 
		SASL authentication password: 
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
		SASL authentication method: 
		SASL authentication username: 
		SASL authentication password: 
`) + " \n\n"

func getKafkaOK(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Brokers:           "127.0.0.1,127.0.0.2",
		Topic:             "logs",
		RequiredACKs:      "-1",
		UseTLS:            true,
		CompressionCodec:  "zippy",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "127.0.0.1,127.0.0.2",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
	}, nil
}

func getKafkaError(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

var describeKafkaOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: log
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
SASL authentication method: 
SASL authentication username: 
SASL authentication password: 
`) + " \n"

func updateKafkaOK(i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Topic:             "logs",
		Brokers:           "127.0.0.1,127.0.0.2",
		RequiredACKs:      "-1",
		CompressionCodec:  "zippy",
		UseTLS:            true,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "127.0.0.1,127.0.0.2",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
	}, nil
}

func updateKafkaSASL(i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Topic:             "logs",
		Brokers:           "127.0.0.1,127.0.0.2",
		RequiredACKs:      "-1",
		CompressionCodec:  "zippy",
		UseTLS:            true,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "127.0.0.1,127.0.0.2",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
		ParseLogKeyvals:   true,
		RequestMaxBytes:   1024,
		AuthMethod:        "plain",
		User:              "user",
		Password:          "password",
	}, nil
}

func updateKafkaError(i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

func deleteKafkaOK(i *fastly.DeleteKafkaInput) error {
	return nil
}

func deleteKafkaError(i *fastly.DeleteKafkaInput) error {
	return errTest
}
