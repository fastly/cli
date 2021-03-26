package kafka

import (
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestCreateKafkaInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateKafkaInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateKafkaInput{
				ServiceID:      "123",
				ServiceVersion: 2,
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
				ServiceVersion:    2,
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
				ServiceVersion:  2,
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
			name:      "verify SASL validation: missing username",
			cmd:       createCommandSASL("scram-sha-256", "", "password"),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:      "verify SASL validation: missing password",
			cmd:       createCommandSASL("plain", "user", ""),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:      "verify SASL validation: username with no auth method or password",
			cmd:       createCommandSASL("", "user1", ""),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:      "verify SASL validation: password with no auth method",
			cmd:       createCommandSASL("", "", "password"),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},

		{
			name:      "verify SASL validation: no SASL, but auth-method given",
			cmd:       createCommandNoSASL("scram-sha-256", "", ""),
			want:      nil,
			wantError: "the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified",
		},
		{
			name:      "verify SASL validation: no SASL, but username with given",
			cmd:       createCommandNoSASL("", "user1", ""),
			want:      nil,
			wantError: "the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified",
		},
		{
			name:      "verify SASL validation: no SASL, but password given",
			cmd:       createCommandNoSASL("", "", "password"),
			want:      nil,
			wantError: "the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			have, err := testcase.cmd.createInput()
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func TestUpdateKafkaInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateKafkaInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetKafkaFn: getKafkaOK},
			want: &fastly.UpdateKafkaInput{
				ServiceID:         "123",
				ServiceVersion:    2,
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
			api:  mock.API{GetKafkaFn: getKafkaOK},
			want: &fastly.UpdateKafkaInput{
				ServiceID:      "123",
				ServiceVersion: 2,
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
			api:  mock.API{GetKafkaFn: getKafkaOK},
			cmd:  updateCommandSASL("scram-sha-512", "user1", "12345"),
			want: &fastly.UpdateKafkaInput{
				ServiceID:       "123",
				ServiceVersion:  2,
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
			api:  mock.API{GetKafkaFn: getKafkaSASL},
			cmd:  updateCommandNoSASL(),
			want: &fastly.UpdateKafkaInput{
				ServiceID:       "123",
				ServiceVersion:  2,
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
			name:      "verify SASL validation: missing username",
			api:       mock.API{GetKafkaFn: getKafkaOK},
			cmd:       updateCommandSASL("scram-sha-256", "", "password"),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:      "verify SASL validation: missing password",
			api:       mock.API{GetKafkaFn: getKafkaOK},
			cmd:       updateCommandSASL("plain", "user", ""),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:      "verify SASL validation: username with no auth method",
			api:       mock.API{GetKafkaFn: getKafkaOK},
			cmd:       updateCommandSASL("", "user1", ""),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
		{
			name:      "verify SASL validation: password with no auth method",
			api:       mock.API{GetKafkaFn: getKafkaOK},
			cmd:       updateCommandSASL("", "", "password"),
			want:      nil,
			wantError: "the --auth-method, --username, and --password flags must be present when using the --use-sasl flag",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.cmd.Base.Globals.Client = testcase.api

			have, err := testcase.cmd.createInput()
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func createCommandRequired() *CreateCommand {
	return &CreateCommand{
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		Version:      2,
		Topic:        "logs",
		Brokers:      "127.0.0.1,127.0.0.2",
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "logs",
		Version:           2,
		Topic:             "logs",
		Brokers:           "127.0.0.1,127.0.0.2",
		UseTLS:            common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		RequiredACKs:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-1"},
		CompressionCodec:  common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "zippy"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandSASL(authMethod, user, password string) *CreateCommand {
	return &CreateCommand{
		manifest:        manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:    "log",
		Version:         2,
		Topic:           "logs",
		Brokers:         "127.0.0.1,127.0.0.2",
		ParseLogKeyvals: common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 11111},
		UseSASL:         common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		AuthMethod:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: authMethod},
		User:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: user},
		Password:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: password},
	}
}

func createCommandNoSASL(authMethod, user, password string) *CreateCommand {
	return &CreateCommand{
		manifest:        manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:    "log",
		Version:         2,
		Topic:           "logs",
		Brokers:         "127.0.0.1,127.0.0.2",
		ParseLogKeyvals: common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 11111},
		UseSASL:         common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: false},
		AuthMethod:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: authMethod},
		User:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: user},
		Password:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: password},
	}
}

func createCommandMissingServiceID() *CreateCommand {
	res := createCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *UpdateCommand {
	return &UpdateCommand{
		Base:         common.Base{Globals: &config.Data{Client: nil}},
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		Version:      2,
	}
}

func updateCommandAll() *UpdateCommand {
	return &UpdateCommand{
		Base:              common.Base{Globals: &config.Data{Client: nil}},
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		NewName:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new1"},
		Topic:             common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		Brokers:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		UseTLS:            common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: false},
		RequiredACKs:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		CompressionCodec:  common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new12"},
		ParseLogKeyvals:   common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: false},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 22222},
		UseSASL:           common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		AuthMethod:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "plain"},
		User:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new13"},
		Password:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new14"},
	}
}

func updateCommandSASL(authMethod, user, password string) *UpdateCommand {
	return &UpdateCommand{
		Base:            common.Base{Globals: &config.Data{Client: nil}},
		manifest:        manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:    "log",
		Version:         2,
		Topic:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "logs"},
		Brokers:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 11111},
		UseSASL:         common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		AuthMethod:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: authMethod},
		User:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: user},
		Password:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: password},
	}
}

func updateCommandNoSASL() *UpdateCommand {
	return &UpdateCommand{
		Base:            common.Base{Globals: &config.Data{Client: nil}},
		manifest:        manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:    "log",
		Version:         2,
		Topic:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "logs"},
		Brokers:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "127.0.0.1,127.0.0.2"},
		ParseLogKeyvals: common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: true},
		RequestMaxBytes: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 11111},
		UseSASL:         common.OptionalBool{Optional: common.Optional{WasSet: true}, Value: false},
		AuthMethod:      common.OptionalString{Optional: common.Optional{WasSet: false}, Value: ""},
		User:            common.OptionalString{Optional: common.Optional{WasSet: false}, Value: ""},
		Password:        common.OptionalString{Optional: common.Optional{WasSet: false}, Value: ""},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

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
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		ParseLogKeyvals:   false,
		RequestMaxBytes:   0,
		AuthMethod:        "",
		User:              "",
		Password:          "",
	}, nil
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
