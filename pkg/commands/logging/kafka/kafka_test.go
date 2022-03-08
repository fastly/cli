package kafka_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/kafka"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestCreateKafkaInput(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		cmd           *kafka.CreateCommand
		want          *fastly.CreateKafkaInput
		wantError     string
		wantSASLError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateKafkaInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
				Topic:          "logs",
				Brokers:        "127.0.0.1,127.0.0.2",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateKafkaInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "logs",
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
				TLSHostname:       "example.com",
				TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
				TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       createCommandMissingServiceID(),
			want:      nil,
			wantError: errors.ErrNoServiceID.Error(),
		},
		{
			name: "verify SASL fields",
			cmd:  createCommandSASL("scram-sha-512", "user1", "12345"),
			want: &fastly.CreateKafkaInput{
				ServiceID:       "123",
				ServiceVersion:  4,
				Name:            "log",
				Topic:           "logs",
				Brokers:         "127.0.0.1,127.0.0.2",
				ParseLogKeyvals: true,
				RequestMaxBytes: 11111,
				AuthMethod:      "scram-sha-512",
				User:            "user1",
				Password:        "12345",
			},
		},
		{
			name:          "verify SASL validation: missing username",
			cmd:           createCommandSASL("scram-sha-256", "", "password"),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:          "verify SASL validation: missing password",
			cmd:           createCommandSASL("plain", "user", ""),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:          "verify SASL validation: username with no auth method or password",
			cmd:           createCommandSASL("", "user1", ""),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:          "verify SASL validation: password with no auth method",
			cmd:           createCommandSASL("", "", "password"),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},

		{
			name:          "verify SASL validation: no SASL, but auth-method given",
			cmd:           createCommandNoSASL("scram-sha-256", "", ""),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified",
		},
		{
			name:          "verify SASL validation: no SASL, but username with given",
			cmd:           createCommandNoSASL("", "user1", ""),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified",
		},
		{
			name:          "verify SASL validation: no SASL, but password given",
			cmd:           createCommandNoSASL("", "", "password"),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.cmd.Base.Globals.APIClient,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})
			if err != nil {
				if testcase.wantError == "" {
					t.Fatalf("unexpected error getting service details: %v", err)
				}
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}
			if err == nil {
				if testcase.wantError != "" {
					t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
				}
			}

			have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
			testutil.AssertErrorContains(t, err, testcase.wantSASLError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func TestUpdateKafkaInput(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		cmd           *kafka.UpdateCommand
		api           mock.API
		want          *fastly.UpdateKafkaInput
		wantError     string
		wantSASLError string
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
				NewName:           fastly.String("new1"),
				Topic:             fastly.String("new2"),
				Brokers:           fastly.String("new3"),
				RequiredACKs:      fastly.String("new4"),
				UseTLS:            fastly.CBool(false),
				CompressionCodec:  fastly.String("new5"),
				Placement:         fastly.String("new6"),
				Format:            fastly.String("new7"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new8"),
				TLSCACert:         fastly.String("new9"),
				TLSClientCert:     fastly.String("new10"),
				TLSClientKey:      fastly.String("new11"),
				TLSHostname:       fastly.String("new12"),
				ParseLogKeyvals:   fastly.CBool(false),
				RequestMaxBytes:   fastly.Uint(22222),
				AuthMethod:        fastly.String("plain"),
				User:              fastly.String("new13"),
				Password:          fastly.String("new14"),
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
				Topic:           fastly.String("logs"),
				Brokers:         fastly.String("127.0.0.1,127.0.0.2"),
				ParseLogKeyvals: fastly.CBool(true),
				RequestMaxBytes: fastly.Uint(11111),
				AuthMethod:      fastly.String("scram-sha-512"),
				User:            fastly.String("user1"),
				Password:        fastly.String("12345"),
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
				Topic:           fastly.String("logs"),
				Brokers:         fastly.String("127.0.0.1,127.0.0.2"),
				ParseLogKeyvals: fastly.CBool(true),
				RequestMaxBytes: fastly.Uint(11111),
				AuthMethod:      fastly.String(""),
				User:            fastly.String(""),
				Password:        fastly.String(""),
			},
		},
		{
			name: "verify SASL validation: missing username",
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			cmd:           updateCommandSASL("scram-sha-256", "", "password"),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name: "verify SASL validation: missing password",
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			cmd:           updateCommandSASL("plain", "user", ""),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name: "verify SASL validation: username with no auth method",
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			cmd:           updateCommandSASL("", "user1", ""),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name: "verify SASL validation: password with no auth method",
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKafkaFn:     getKafkaOK,
			},
			cmd:           updateCommandSASL("", "", "password"),
			want:          nil,
			wantSASLError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.cmd.Base.Globals.APIClient = testcase.api

			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.api,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})
			if err != nil {
				if testcase.wantError == "" {
					t.Fatalf("unexpected error getting service details: %v", err)
				}
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}
			if err == nil {
				if testcase.wantError != "" {
					t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
				}
			}

			have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
			testutil.AssertErrorContains(t, err, testcase.wantSASLError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func createCommandRequired() *kafka.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.ClientFactory(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &kafka.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:   "logs",
		Brokers: "127.0.0.1,127.0.0.2",
	}
}

func createCommandAll() *kafka.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.ClientFactory(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &kafka.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:             "logs",
		Brokers:           "127.0.0.1,127.0.0.2",
		UseTLS:            cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		RequiredACKs:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-1"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "zippy"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		TLSCACert:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandSASL(authMethod, user, password string) *kafka.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.ClientFactory(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &kafka.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:           "logs",
		Brokers:         "127.0.0.1,127.0.0.2",
		ParseLogKeyvals: cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 11111},
		UseSASL:         cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		AuthMethod:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: authMethod},
		User:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: user},
		Password:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: password},
	}
}

func createCommandNoSASL(authMethod, user, password string) *kafka.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.ClientFactory(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &kafka.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:           "logs",
		Brokers:         "127.0.0.1,127.0.0.2",
		ParseLogKeyvals: cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 11111},
		UseSASL:         cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: false},
		AuthMethod:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: authMethod},
		User:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: user},
		Password:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: password},
	}
}

func createCommandMissingServiceID() *kafka.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *kafka.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
	}
}

func updateCommandAll() *kafka.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		NewName:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new1"},
		Topic:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		Brokers:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		UseTLS:            cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: false},
		RequiredACKs:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		TLSCACert:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		TLSClientCert:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		TLSClientKey:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
		TLSHostname:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new12"},
		ParseLogKeyvals:   cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: false},
		RequestMaxBytes:   cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 22222},
		UseSASL:           cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		AuthMethod:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "plain"},
		User:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new13"},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new14"},
	}
}

func updateCommandSASL(authMethod, user, password string) *kafka.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "logs"},
		Brokers:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 11111},
		UseSASL:         cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		AuthMethod:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: authMethod},
		User:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: user},
		Password:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: password},
	}
}

func updateCommandNoSASL() *kafka.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kafka.UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		ServiceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		AutoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{
				Optional: cmd.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		Topic:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "logs"},
		Brokers:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 11111},
		UseSASL:         cmd.OptionalBool{Optional: cmd.Optional{WasSet: true}, Value: false},
		AuthMethod:      cmd.OptionalString{Optional: cmd.Optional{WasSet: false}, Value: ""},
		User:            cmd.OptionalString{Optional: cmd.Optional{WasSet: false}, Value: ""},
		Password:        cmd.OptionalString{Optional: cmd.Optional{WasSet: false}, Value: ""},
	}
}

func updateCommandMissingServiceID() *kafka.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func getKafkaSASL(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
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
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		ParseLogKeyvals:   false,
		RequestMaxBytes:   0,
		AuthMethod:        "plain",
		User:              "user",
		Password:          "password",
	}, nil
}
