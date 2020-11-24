package gcs

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

func TestCreateGCSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateGCSInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateGCSInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Bucket:         "bucket",
				User:           "user",
				SecretKey:      "-----BEGIN PRIVATE KEY-----foo",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateGCSInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				Bucket:            "bucket",
				User:              "user",
				SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
				Path:              "/logs",
				Period:            3600,
				FormatVersion:     2,
				GzipLevel:         2,
				Format:            `%h %l %u %t "%r" %>s %b`,
				MessageType:       "classic",
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

func TestUpdateGCSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateGCSInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetGCSFn: getGCSOK},
			want: &fastly.UpdateGCSInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("logs"),
				Bucket:            fastly.String("bucket"),
				User:              fastly.String("user"),
				SecretKey:         fastly.String("-----BEGIN PRIVATE KEY-----foo"),
				Path:              fastly.String("/logs"),
				Period:            fastly.Uint(3600),
				FormatVersion:     fastly.Uint(2),
				GzipLevel:         fastly.Uint8(2),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.String("Prevent default logging"),
				TimestampFormat:   fastly.String("%Y-%m-%dT%H:%M:%S.000"),
				MessageType:       fastly.String("classic"),
				Placement:         fastly.String("none"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetGCSFn: getGCSOK},
			want: &fastly.UpdateGCSInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("new1"),
				Bucket:            fastly.String("new2"),
				User:              fastly.String("new3"),
				SecretKey:         fastly.String("new4"),
				Path:              fastly.String("new5"),
				Period:            fastly.Uint(3601),
				FormatVersion:     fastly.Uint(3),
				GzipLevel:         fastly.Uint8(3),
				Format:            fastly.String("new6"),
				ResponseCondition: fastly.String("new7"),
				TimestampFormat:   fastly.String("new8"),
				Placement:         fastly.String("new9"),
				MessageType:       fastly.String("new10"),
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
		Bucket:       "bucket",
		User:         "user",
		SecretKey:    "-----BEGIN PRIVATE KEY-----foo",
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		Bucket:            "bucket",
		User:              "user",
		SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
		Path:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "/logs"},
		Period:            common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3600},
		GzipLevel:         common.OptionalUint8{Optional: common.Optional{WasSet: true}, Value: 2},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
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
		Bucket:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		User:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		Path:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Period:            common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         common.OptionalUint8{Optional: common.Optional{WasSet: true}, Value: 3},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getGCSOK(i *fastly.GetGCSInput) (*fastly.GCS, error) {
	return &fastly.GCS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Bucket:            "bucket",
		User:              "user",
		SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
		Path:              "/logs",
		Period:            3600,
		GzipLevel:         2,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}
