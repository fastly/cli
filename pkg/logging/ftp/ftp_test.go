package ftp

import (
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v2/fastly"
)

func TestCreateFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateFTPInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateFTPInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Address:        "example.com",
				Username:       "user",
				Password:       "password",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateFTPInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				Address:           "example.com",
				Port:              22,
				Username:          "user",
				Password:          "password",
				Path:              "/logs",
				Period:            3600,
				FormatVersion:     2,
				GzipLevel:         2,
				Format:            `%h %l %u %t "%r" %>s %b`,
				ResponseCondition: "Prevent default logging",
				TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
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

func TestUpdateFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateFTPInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetFTPFn: getFTPOK},
			want: &fastly.UpdateFTPInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("logs"),
				Address:           fastly.String("example.com"),
				Port:              fastly.Uint(22),
				Username:          fastly.String("user"),
				Password:          fastly.String("password"),
				Path:              fastly.String("/logs"),
				Period:            fastly.Uint(3600),
				FormatVersion:     fastly.Uint(2),
				GzipLevel:         fastly.Uint8(2),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.String("Prevent default logging"),
				TimestampFormat:   fastly.String("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         fastly.String("none"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetFTPFn: getFTPOK},
			want: &fastly.UpdateFTPInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Uint(23),
				Username:          fastly.String("new3"),
				Password:          fastly.String("new4"),
				Path:              fastly.String("new5"),
				Period:            fastly.Uint(3601),
				FormatVersion:     fastly.Uint(3),
				GzipLevel:         fastly.Uint8(3),
				Format:            fastly.String("new6"),
				ResponseCondition: fastly.String("new7"),
				TimestampFormat:   fastly.String("new8"),
				Placement:         fastly.String("new9"),
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
		Username:     "user",
		Password:     "password",
		Version:      2,
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		Address:           "example.com",
		Username:          "user",
		Password:          "password",
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 22},
		Path:              common.OptionalString{Optional: common.Optional{Valid: true}, Value: "/logs"},
		Period:            common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3600},
		GzipLevel:         common.OptionalUint8{Optional: common.Optional{Valid: true}, Value: 2},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{Valid: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
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
		Username:          common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		Password:          common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		Path:              common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		Period:            common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3601},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		GzipLevel:         common.OptionalUint8{Optional: common.Optional{Valid: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new9"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getFTPOK(i *fastly.GetFTPInput) (*fastly.FTP, error) {
	return &fastly.FTP{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com",
		Port:              22,
		Username:          "user",
		Password:          "password",
		Path:              "/logs",
		Period:            3600,
		GzipLevel:         2,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}
