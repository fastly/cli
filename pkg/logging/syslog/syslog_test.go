package syslog

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

func TestCreateSyslogInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateSyslogInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateSyslogInput{
				Service: "123",
				Version: 2,
				Name:    "log",
				Address: "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateSyslogInput{
				Service:           "123",
				Version:           2,
				Name:              "log",
				Address:           "example.com",
				Port:              22,
				UseTLS:            fastly.CBool(true),
				TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
				TLSHostname:       "example.com",
				TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
				TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
				Token:             "tkn",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				MessageType:       "classic",
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
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

func TestUpdateSyslogInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateSyslogInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetSyslogFn: getSyslogOK},
			want: &fastly.UpdateSyslogInput{
				Service:           "123",
				Version:           2,
				Name:              "logs",
				NewName:           "logs",
				Address:           "example.com",
				Port:              22,
				UseTLS:            fastly.CBool(true),
				TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
				TLSHostname:       "example.com",
				TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
				TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
				Token:             "tkn",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				MessageType:       "classic",
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetSyslogFn: getSyslogOK},
			want: &fastly.UpdateSyslogInput{
				Service:           "123",
				Version:           2,
				Name:              "logs",
				NewName:           "new1",
				Address:           "new2",
				Port:              23,
				UseTLS:            fastly.CBool(false),
				TLSCACert:         "new3",
				TLSHostname:       "new4",
				TLSClientCert:     "new5",
				TLSClientKey:      "new6",
				Token:             "new7",
				Format:            "new8",
				FormatVersion:     3,
				MessageType:       "new9",
				ResponseCondition: "new10",
				Placement:         "new11",
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
		Address:      "example.com",
		Version:      2,
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		Address:           "example.com",
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 22},
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: true},
		TLSCACert:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "example.com"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
		Token:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "tkn"},
		MessageType:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "classic"},
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
		Address:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 23},
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: false},
		TLSCACert:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		Token:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		MessageType:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new9"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new10"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new11"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getSyslogOK(i *fastly.GetSyslogInput) (*fastly.Syslog, error) {
	return &fastly.Syslog{
		ServiceID:         i.Service,
		Version:           i.Version,
		Name:              "logs",
		Address:           "example.com",
		Port:              22,
		UseTLS:            true,
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}
