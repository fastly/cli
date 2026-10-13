package gcs_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/gcs"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateGCSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *gcs.CreateCommand
		want      *fastly.CreateGCSInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateGCSInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           new("log"),
				Bucket:         new("bucket"),
				User:           new("user"),
				SecretKey:      new("-----BEGIN PRIVATE KEY-----foo"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateGCSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              new("log"),
				Bucket:            new("bucket"),
				User:              new("user"),
				SecretKey:         new("-----BEGIN PRIVATE KEY-----foo"),
				Path:              new("/logs"),
				Period:            new(3600),
				FormatVersion:     new(2),
				Format:            new(`%h %l %u %t "%r" %>s %b`),
				MessageType:       new("classic"),
				ResponseCondition: new("Prevent default logging"),
				TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         new("none"),
				CompressionCodec:  new("zstd"),
				ProcessingRegion:  new("eu"),
				PublicKey:         new(pgpPublicKey()),
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
			if testcase.wantError == errors.ErrNoServiceID.Error() {
				t.Setenv("FASTLY_SERVICE_ID", "")
			}
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

func TestUpdateGCSInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *gcs.UpdateCommand
		api       mock.API
		want      *fastly.UpdateGCSInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetGCSFn:       getGCSOK,
			},
			want: &fastly.UpdateGCSInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetGCSFn:       getGCSOK,
			},
			want: &fastly.UpdateGCSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				Bucket:            new("new2"),
				User:              new("new3"),
				SecretKey:         new("new4"),
				Path:              new("new5"),
				Period:            new(3601),
				FormatVersion:     new(3),
				GzipLevel:         new(0),
				Format:            new("new6"),
				ResponseCondition: new("new7"),
				TimestampFormat:   new("new8"),
				Placement:         fastly.NewNullable("new9"),
				MessageType:       new("new10"),
				CompressionCodec:  new("new11"),
				ProcessingRegion:  new("eu"),
				PublicKey:         new("new12"),
			},
		},
		{
			name: "reset placement to null",
			cmd:  updateCommandPlacementReset(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetGCSFn:       getGCSOK,
			},
			want: &fastly.UpdateGCSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				Bucket:            new("new2"),
				User:              new("new3"),
				SecretKey:         new("new4"),
				Path:              new("new5"),
				Period:            new(3601),
				FormatVersion:     new(3),
				GzipLevel:         new(0),
				Format:            new("new6"),
				ResponseCondition: new("new7"),
				TimestampFormat:   new("new8"),
				Placement:         fastly.NullValue[string](),
				MessageType:       new("new10"),
				CompressionCodec:  new("new11"),
				ProcessingRegion:  new("eu"),
				PublicKey:         new("new12"),
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
			if testcase.wantError == errors.ErrNoServiceID.Error() {
				t.Setenv("FASTLY_SERVICE_ID", "")
			}
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

func createCommandRequired() *gcs.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		GetVersionFn:   testutil.GetVersion,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &gcs.CreateCommand{
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
		Bucket:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		User:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		SecretKey:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----foo"},
	}
}

func createCommandAll() *gcs.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		GetVersionFn:   testutil.GetVersion,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &gcs.CreateCommand{
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
		EndpointName:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Bucket:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		SecretKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----foo"},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "/logs"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3600},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zstd"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: pgpPublicKey()},
	}
}

func createCommandMissingServiceID() *gcs.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}

func updateCommandNoUpdates() *gcs.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &gcs.UpdateCommand{
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

func updateCommandAll() *gcs.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &gcs.UpdateCommand{
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
		Bucket:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		SecretKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 0},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
	}
}

func updateCommandPlacementReset() *gcs.UpdateCommand {
	c := updateCommandAll()
	c.Placement = argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: ""}
	return c
}

func updateCommandMissingServiceID() *gcs.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}
