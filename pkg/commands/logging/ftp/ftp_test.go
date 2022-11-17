package ftp_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/ftp"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestCreateFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *ftp.CreateCommand
		want      *fastly.CreateFTPInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateFTPInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.String("log"),
				Address:        fastly.String("example.com"),
				Username:       fastly.String("user"),
				Password:       fastly.String("password"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateFTPInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.String("log"),
				Address:           fastly.String("example.com"),
				Port:              fastly.Int(22),
				Username:          fastly.String("user"),
				Password:          fastly.String("password"),
				Path:              fastly.String("/logs"),
				Period:            fastly.Int(3600),
				FormatVersion:     fastly.Int(2),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.String("Prevent default logging"),
				TimestampFormat:   fastly.String("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         fastly.String("none"),
				CompressionCodec:  fastly.String("zstd"),
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
			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
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
				have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
				testutil.AssertErrorContains(t, err, testcase.wantError)
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

func TestUpdateFTPInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *ftp.UpdateCommand
		api       mock.API
		want      *fastly.UpdateFTPInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
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
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetFTPFn:       getFTPOK,
			},
			want: &fastly.UpdateFTPInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Int(23),
				PublicKey:         fastly.String("new10"),
				Username:          fastly.String("new3"),
				Password:          fastly.String("new4"),
				Path:              fastly.String("new5"),
				Period:            fastly.Int(3601),
				FormatVersion:     fastly.Int(3),
				GzipLevel:         fastly.Int(0),
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
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.name, func(t *testing.T) {
			testcase.cmd.Globals.APIClient = testcase.api

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
				have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
				testutil.AssertErrorContains(t, err, testcase.wantError)
				testutil.AssertEqual(t, testcase.want, have)
			}
		})
	}
}

func createCommandRequired() *ftp.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &ftp.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		Address:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
		Username:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "user"},
		Password:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "password"},
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

func createCommandAll() *ftp.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &ftp.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
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
		EndpointName:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		Address:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
		Username:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "user"},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "password"},
		Port:              cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 22},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "/logs"},
		Period:            cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3600},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "zstd"},
	}
}

func createCommandMissingServiceID() *ftp.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *ftp.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &ftp.UpdateCommand{
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

func updateCommandAll() *ftp.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &ftp.UpdateCommand{
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
		Address:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		Port:              cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 23},
		Username:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		PublicKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Period:            cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 0},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
	}
}

func updateCommandMissingServiceID() *ftp.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
