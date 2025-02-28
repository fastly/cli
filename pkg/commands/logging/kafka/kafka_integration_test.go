package kafka_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestKafkaCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging kafka create --service-id 123 --version 1 --name log --topic logs --brokers 127.0.0.1127.0.0.2 --parse-log-keyvals --max-batch-size 1024 --use-sasl --auth-method plain --username user --password password --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateKafkaFn:  createKafkaOK,
			},
			wantOutput: "Created Kafka logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging kafka create --service-id 123 --version 1 --name log --topic logs --brokers 127.0.0.1127.0.0.2 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateKafkaFn:  createKafkaError,
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

func TestKafkaList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging kafka list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKafkasFn:   listKafkasOK,
			},
			wantOutput: listKafkasShortOutput,
		},
		{
			args: args("logging kafka list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKafkasFn:   listKafkasOK,
			},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args: args("logging kafka list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKafkasFn:   listKafkasOK,
			},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args: args("logging kafka --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKafkasFn:   listKafkasOK,
			},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args: args("logging -v kafka list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKafkasFn:   listKafkasOK,
			},
			wantOutput: listKafkasVerboseOutput,
		},
		{
			args: args("logging kafka list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListKafkasFn:   listKafkasError,
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

func TestKafkaDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging kafka describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging kafka describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetKafkaFn:     getKafkaError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kafka describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetKafkaFn:     getKafkaOK,
			},
			wantOutput: describeKafkaOutput,
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

func TestKafkaUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging kafka update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging kafka update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateKafkaFn:  updateKafkaError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kafka update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateKafkaFn:  updateKafkaOK,
			},
			wantOutput: "Updated Kafka logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging kafka update --service-id 123 --version 1 --name logs --new-name log --parse-log-keyvals --max-batch-size 1024 --use-sasl --auth-method plain --username user --password password --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateKafkaFn:  updateKafkaSASL,
			},
			wantOutput: "Updated Kafka logging endpoint log (service 123 version 4)",
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

func TestKafkaDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging kafka delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging kafka delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteKafkaFn:  deleteKafkaError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging kafka delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteKafkaFn:  deleteKafkaOK,
			},
			wantOutput: "Deleted Kafka logging endpoint logs (service 123 version 4)",
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

func createKafkaOK(i *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Topic:             fastly.ToPointer("logs"),
		Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
		RequiredACKs:      fastly.ToPointer("-1"),
		CompressionCodec:  fastly.ToPointer("zippy"),
		UseTLS:            fastly.ToPointer(true),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("127.0.0.1,127.0.0.2"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     fastly.ToPointer(2),
		ParseLogKeyvals:   fastly.ToPointer(true),
		RequestMaxBytes:   fastly.ToPointer(1024),
		AuthMethod:        fastly.ToPointer("plain"),
		User:              fastly.ToPointer("user"),
		Password:          fastly.ToPointer("password"),
	}, nil
}

func createKafkaError(_ *fastly.CreateKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

func listKafkasOK(i *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return []*fastly.Kafka{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			Topic:             fastly.ToPointer("logs"),
			Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
			RequiredACKs:      fastly.ToPointer("-1"),
			CompressionCodec:  fastly.ToPointer("zippy"),
			UseTLS:            fastly.ToPointer(true),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       fastly.ToPointer("127.0.0.1,127.0.0.2"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			FormatVersion:     fastly.ToPointer(2),
			ParseLogKeyvals:   fastly.ToPointer(false),
			RequestMaxBytes:   fastly.ToPointer(0),
			AuthMethod:        fastly.ToPointer("plain"),
			User:              fastly.ToPointer("user"),
			Password:          fastly.ToPointer("password"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			Topic:             fastly.ToPointer("analytics"),
			Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
			RequiredACKs:      fastly.ToPointer("-1"),
			CompressionCodec:  fastly.ToPointer("zippy"),
			UseTLS:            fastly.ToPointer(true),
			TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
			TLSHostname:       fastly.ToPointer("127.0.0.1,127.0.0.2"),
			TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
			TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ParseLogKeyvals:   fastly.ToPointer(false),
			RequestMaxBytes:   fastly.ToPointer(0),
			AuthMethod:        fastly.ToPointer("plain"),
			User:              fastly.ToPointer("user"),
			Password:          fastly.ToPointer("password"),
		},
	}, nil
}

func listKafkasError(_ *fastly.ListKafkasInput) ([]*fastly.Kafka, error) {
	return nil, errTest
}

var listKafkasShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listKafkasVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

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
		Parse log key-values: false
		Max batch size: 0
		SASL authentication method: plain
		SASL authentication username: user
		SASL authentication password: password
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
		Parse log key-values: false
		Max batch size: 0
		SASL authentication method: plain
		SASL authentication username: user
		SASL authentication password: password
  `) + "\n\n"

func getKafkaOK(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
		Topic:             fastly.ToPointer("logs"),
		RequiredACKs:      fastly.ToPointer("-1"),
		UseTLS:            fastly.ToPointer(true),
		CompressionCodec:  fastly.ToPointer("zippy"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("127.0.0.1,127.0.0.2"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
	}, nil
}

func getKafkaError(_ *fastly.GetKafkaInput) (*fastly.Kafka, error) {
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

func updateKafkaOK(i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Topic:             fastly.ToPointer("logs"),
		Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
		RequiredACKs:      fastly.ToPointer("-1"),
		CompressionCodec:  fastly.ToPointer("zippy"),
		UseTLS:            fastly.ToPointer(true),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("127.0.0.1,127.0.0.2"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     fastly.ToPointer(2),
	}, nil
}

func updateKafkaSASL(i *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		Topic:             fastly.ToPointer("logs"),
		Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
		RequiredACKs:      fastly.ToPointer("-1"),
		CompressionCodec:  fastly.ToPointer("zippy"),
		UseTLS:            fastly.ToPointer(true),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("127.0.0.1,127.0.0.2"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		FormatVersion:     fastly.ToPointer(2),
		ParseLogKeyvals:   fastly.ToPointer(true),
		RequestMaxBytes:   fastly.ToPointer(1024),
		AuthMethod:        fastly.ToPointer("plain"),
		User:              fastly.ToPointer("user"),
		Password:          fastly.ToPointer("password"),
	}, nil
}

func updateKafkaError(_ *fastly.UpdateKafkaInput) (*fastly.Kafka, error) {
	return nil, errTest
}

func deleteKafkaOK(_ *fastly.DeleteKafkaInput) error {
	return nil
}

func deleteKafkaError(_ *fastly.DeleteKafkaInput) error {
	return errTest
}
