package kafka

import (
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/fastly"
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
				Service: "123",
				Version: 2,
				Name:    fastly.String("log"),
				Topic:   fastly.String("logs"),
				Brokers: fastly.String("127.0.0.1,127.0.0.2"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateKafkaInput{
				Service:           "123",
				Version:           2,
				Name:              fastly.String("logs"),
				Brokers:           fastly.String("127.0.0.1,127.0.0.2"),
				Topic:             fastly.String("logs"),
				RequiredACKs:      fastly.String("-1"),
				UseTLS:            fastly.CBool(true),
				CompressionCodec:  fastly.String("zippy"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.Uint(2),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
				TLSCACert:         fastly.String("-----BEGIN CERTIFICATE-----foo"),
				TLSHostname:       fastly.String("example.com"),
				TLSClientCert:     fastly.String("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      fastly.String("-----BEGIN PRIVATE KEY-----bar"),
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       createCommandMissingServiceID(),
			want:      nil,
			wantError: errors.ErrNoServiceID.Error(),
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
				Service:           "123",
				Version:           2,
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
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetKafkaFn: getKafkaOK},
			want: &fastly.UpdateKafkaInput{
				Service:           "123",
				Version:           2,
				Name:              "log",
				NewName:           fastly.String("log"),
				Brokers:           fastly.String("127.0.0.1,127.0.0.2"),
				Topic:             fastly.String("logs"),
				RequiredACKs:      fastly.String("-1"),
				UseTLS:            fastly.CBool(true),
				CompressionCodec:  fastly.String("zippy"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.Uint(2),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
				TLSCACert:         fastly.String("-----BEGIN CERTIFICATE-----foo"),
				TLSHostname:       fastly.String("example.com"),
				TLSClientCert:     fastly.String("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      fastly.String("-----BEGIN PRIVATE KEY-----bar"),
			},
		},
		{
			name:      "error missing serviceID",
			cmd:       updateCommandMissingServiceID(),
			want:      nil,
			wantError: errors.ErrNoServiceID.Error(),
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
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: true},
		RequiredACKs:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-1"},
		CompressionCodec:  common.OptionalString{Optional: common.Optional{Valid: true}, Value: "zippy"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "example.com"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
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
		NewName:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new1"},
		Topic:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		Brokers:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: false},
		RequiredACKs:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		CompressionCodec:  common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new9"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new10"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new11"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new12"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getKafkaOK(i *fastly.GetKafkaInput) (*fastly.Kafka, error) {
	return &fastly.Kafka{
		ServiceID:         i.Service,
		Version:           i.Version,
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
	}, nil
}
