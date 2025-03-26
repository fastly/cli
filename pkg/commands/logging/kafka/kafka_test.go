package kafka_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/kafka"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateKafkaInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *kafka.CreateCommand
		want      *fastly.CreateKafkaInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateKafkaInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				Topic:          fastly.ToPointer("logs"),
				Brokers:        fastly.ToPointer("127.0.0.1,127.0.0.2"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateKafkaInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.ToPointer("logs"),
				Brokers:           fastly.ToPointer("127.0.0.1,127.0.0.2"),
				Topic:             fastly.ToPointer("logs"),
				RequiredACKs:      fastly.ToPointer("-1"),
				UseTLS:            fastly.ToPointer(fastly.Compatibool(true)),
				CompressionCodec:  fastly.ToPointer("zippy"),
				Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.ToPointer(2),
				ResponseCondition: fastly.ToPointer("Prevent default logging"),
				Placement:         fastly.ToPointer("none"),
				TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
				TLSHostname:       fastly.ToPointer("example.com"),
				TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       createCommandMissingServiceID(),
			wantError: errors.ErrNoServiceID.Error(),
		},
		{
			name: "verify SASL fields",
			cmd:  createCommandSASL("scram-sha-512", "user1", "12345"),
			want: &fastly.CreateKafkaInput{
				ServiceID:       "123",
				ServiceVersion:  4,
				Name:            fastly.ToPointer("log"),
				Topic:           fastly.ToPointer("logs"),
				Brokers:         fastly.ToPointer("127.0.0.1,127.0.0.2"),
				ParseLogKeyvals: fastly.ToPointer(fastly.Compatibool(true)),
				RequestMaxBytes: fastly.ToPointer(11111),
				AuthMethod:      fastly.ToPointer("scram-sha-512"),
				User:            fastly.ToPointer("user1"),
				Password:        fastly.ToPointer("12345"),
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.cmd.Globals.APIClient,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})

			switch {
			case err != nil && testcase.wantError == "":
				t.Fatalf("unexpected error getting service details: %v", err)
				return
			case err != nil && testcase.wantError != "":
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			case err == nil && testcase.wantError != "":
				t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
			case err == nil && testcase.wantError == "":
				have, err := testcase.cmd.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
				testutil.AssertErrorContains(t, err, testcase.wantError)
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

func TestUpdateKafkaInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *kafka.UpdateCommand
		api       mock.API
		want      *fastly.UpdateKafkaInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			want: &fastly.UpdateKafkaInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.ToPointer("new1"),
				Topic:             fastly.ToPointer("new2"),
				Brokers:           fastly.ToPointer("new3"),
				RequiredACKs:      fastly.ToPointer("new4"),
				UseTLS:            fastly.ToPointer(fastly.Compatibool(false)),
				CompressionCodec:  fastly.ToPointer("new5"),
				Placement:         fastly.ToPointer("new6"),
				Format:            fastly.ToPointer("new7"),
				FormatVersion:     fastly.ToPointer(3),
				ResponseCondition: fastly.ToPointer("new8"),
				TLSCACert:         fastly.ToPointer("new9"),
				TLSClientCert:     fastly.ToPointer("new10"),
				TLSClientKey:      fastly.ToPointer("new11"),
				TLSHostname:       fastly.ToPointer("new12"),
				ParseLogKeyvals:   fastly.ToPointer(fastly.Compatibool(false)),
				RequestMaxBytes:   fastly.ToPointer(22222),
				AuthMethod:        fastly.ToPointer("plain"),
				User:              fastly.ToPointer("new13"),
				Password:          fastly.ToPointer("new14"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			want: &fastly.UpdateKafkaInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       updateCommandMissingServiceID(),
			want:      nil,
			wantError: errors.ErrNoServiceID.Error(),
		},
		{
			name: "verify SASL fields",
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			cmd: updateCommandSASL("scram-sha-512", "user1", "12345"),
			want: &fastly.UpdateKafkaInput{
				ServiceID:       "123",
				ServiceVersion:  4,
				Name:            "log",
				Topic:           fastly.ToPointer("logs"),
				Brokers:         fastly.ToPointer("127.0.0.1,127.0.0.2"),
				ParseLogKeyvals: fastly.ToPointer(fastly.Compatibool(true)),
				RequestMaxBytes: fastly.ToPointer(11111),
				AuthMethod:      fastly.ToPointer("scram-sha-512"),
				User:            fastly.ToPointer("user1"),
				Password:        fastly.ToPointer("12345"),
			},
		},
		{
			name: "verify disabling SASL",
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaSASL,
			},
			cmd: updateCommandNoSASL(),
			want: &fastly.UpdateKafkaInput{
				ServiceID:       "123",
				ServiceVersion:  4,
				Name:            "log",
				Topic:           fastly.ToPointer("logs"),
				Brokers:         fastly.ToPointer("127.0.0.1,127.0.0.2"),
				ParseLogKeyvals: fastly.ToPointer(fastly.Compatibool(true)),
				RequestMaxBytes: fastly.ToPointer(11111),
				AuthMethod:      fastly.ToPointer(""),
				User:            fastly.ToPointer(""),
				Password:        fastly.ToPointer(""),
			},
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.name, func(t *testing.T) {
			testcase.cmd.Globals.APIClient = testcase.api

			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.api,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})

			switch {
			case err != nil && testcase.wantError == "":
				t.Fatalf("unexpected error getting service details: %v", err)
				return
			case err != nil && testcase.wantError != "":
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			case err == nil && testcase.wantError != "":
				t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
			case err == nil && testcase.wantError == "":
				have, err := testcase.cmd.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
				testutil.AssertErrorContains(t, err, testcase.wantError)
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

func createCommandRequired() *kafka.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &kafka.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Topic:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Brokers:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
	}
}

func createCommandAll() *kafka.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &kafka.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Topic:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Brokers:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		UseTLS:            argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		RequiredACKs:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-1"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zippy"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandSASL(authMethod, user, password string) *kafka.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &kafka.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		EndpointName:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Topic:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Brokers:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 11111},
		UseSASL:         argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		AuthMethod:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: authMethod},
		User:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: user},
		Password:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: password},
	}
}

func createCommandMissingServiceID() *kafka.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *kafka.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
	}
}

func updateCommandAll() *kafka.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		NewName:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new1"},
		Topic:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		Brokers:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		UseTLS:            argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: false},
		RequiredACKs:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
		ParseLogKeyvals:   argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: false},
		RequestMaxBytes:   argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 22222},
		UseSASL:           argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		AuthMethod:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "plain"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new13"},
		Password:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new14"},
	}
}

func updateCommandSASL(authMethod, user, password string) *kafka.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Brokers:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 11111},
		UseSASL:         argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		AuthMethod:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: authMethod},
		User:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: user},
		Password:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: password},
	}
}

func updateCommandNoSASL() *kafka.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Brokers:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 11111},
		UseSASL:         argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: false},
		AuthMethod:      argparser.OptionalString{Optional: argparser.Optional{WasSet: false}, Value: ""},
		User:            argparser.OptionalString{Optional: argparser.Optional{WasSet: false}, Value: ""},
		Password:        argparser.OptionalString{Optional: argparser.Optional{WasSet: false}, Value: ""},
	}
}

func updateCommandMissingServiceID() *kafka.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func getKafkaSASL(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
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
		Placement:         fastly.ToPointer("none"),
		TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
		TLSHostname:       fastly.ToPointer("example.com"),
		TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
		TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
		ParseLogKeyvals:   fastly.ToPointer(false),
		RequestMaxBytes:   fastly.ToPointer(0),
		AuthMethod:        fastly.ToPointer("plain"),
		User:              fastly.ToPointer("user"),
		Password:          fastly.ToPointer("password"),
	}, nil
}
