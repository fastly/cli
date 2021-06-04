package ftp

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
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
				ServiceVersion: 4,
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
				ServiceVersion:    4,
				Name:              "log",
				Address:           "example.com",
				Port:              22,
				Username:          "user",
				Password:          "password",
				Path:              "/logs",
				Period:            3600,
				FormatVersion:     2,
				Format:            `%h %l %u %t "%r" %>s %b`,
				ResponseCondition: "Prevent default logging",
				TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
				Placement:         "none",
				CompressionCodec:  "zstd",
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
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVersionFn:   testutil.GetActiveVersion(1),
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetFTPFn:       getFTPOK,
			},
			want: &fastly.UpdateFTPInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVersionFn:   testutil.GetActiveVersion(1),
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetFTPFn:       getFTPOK,
			},
			want: &fastly.UpdateFTPInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Uint(23),
				PublicKey:         fastly.String("new10"),
				Username:          fastly.String("new3"),
				Password:          fastly.String("new4"),
				Path:              fastly.String("new5"),
				Period:            fastly.Uint(3601),
				FormatVersion:     fastly.Uint(3),
				GzipLevel:         fastly.Uint8(0),
				Format:            fastly.String("new6"),
				ResponseCondition: fastly.String("new7"),
				TimestampFormat:   fastly.String("new8"),
				Placement:         fastly.String("new9"),
				CompressionCodec:  fastly.String("new11"),
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
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.Client, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		GetVersionFn:   testutil.GetActiveVersion(1),
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		Address:      "example.com",
		Username:     "user",
		Password:     "password",
		serviceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		autoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{Value: true},
		},
	}
}

func createCommandAll() *CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.Client, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		GetVersionFn:   testutil.GetActiveVersion(1),
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		serviceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		autoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{Value: true},
		},
		Address:           "example.com",
		Username:          "user",
		Password:          "password",
		Port:              cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 22},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "/logs"},
		Period:            cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3600},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "zstd"},
	}
}

func createCommandMissingServiceID() *CreateCommand {
	res := createCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		serviceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		autoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{Value: true},
		},
	}
}

func updateCommandAll() *UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		serviceVersion: cmd.OptionalServiceVersion{
			OptionalString: cmd.OptionalString{Value: "1"},
		},
		autoClone: cmd.OptionalAutoClone{
			OptionalBool: cmd.OptionalBool{Value: true},
		},
		NewName:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new1"},
		Address:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		Port:              cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 23},
		Username:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		PublicKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Period:            cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         cmd.OptionalUint8{Optional: cmd.Optional{WasSet: true}, Value: 0},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
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
		GzipLevel:         0,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
		CompressionCodec:  "zstd",
	}, nil
}
