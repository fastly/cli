package fos_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/fos"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func optString(v string) argparser.OptionalString {
	return argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: v}
}

func optInt(v int) argparser.OptionalInt {
	return argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: v}
}

func autoClone() argparser.OptionalAutoClone {
	return argparser.OptionalAutoClone{
		OptionalBool: argparser.OptionalBool{
			Optional: argparser.Optional{WasSet: true},
			Value:    true,
		},
	}
}

func newGlobals() *global.Data {
	var b bytes.Buffer
	return &global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
}

func TestCreateFOSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *fos.CreateCommand
		want      *fastly.CreateFOSInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateFOSInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           new("log"),
				BucketName:     new("bucket"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateFOSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              new("logs"),
				BucketName:        new("bucket"),
				Region:            new("us-east"),
				Path:              new("path"),
				Period:            new(3600),
				Format:            new(`%h %l %u %t "%r" %>s %b`),
				MessageType:       new("classic"),
				FormatVersion:     new(2),
				ResponseCondition: new("Prevent default logging"),
				TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         new("none"),
				PublicKey:         new("public-key"),
				CompressionCodec:  new("zstd"),
				FileMaxBytes:      new(1048576),
			},
		},
		{
			name:      "compression codec and gzip level are mutually exclusive",
			cmd:       createCommandCodecAndGzip(),
			wantError: "the --compression-codec flag is mutually exclusive with the --gzip-level flag",
		},
		{
			name:      "error missing serviceID",
			cmd:       createCommandMissingServiceID(),
			wantError: errors.ErrNoServiceID.Error(),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.wantError == errors.ErrNoServiceID.Error() {
				t.Setenv("FASTLY_SERVICE_ID", "")
			}
			serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.cmd.Globals.APIClient,
				Manifest:           testcase.cmd.Manifest,
				Out:                &bytes.Buffer{},
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        true,
			})
			if err != nil {
				if testcase.wantError == "" {
					t.Fatalf("unexpected error getting service details: %v", err)
				}
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}

			have, err := testcase.cmd.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
			if testcase.wantError != "" {
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func TestUpdateFOSInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *fos.UpdateCommand
		api       mock.API
		want      *fastly.UpdateFOSInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			want: &fastly.UpdateFOSInput{
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
			},
			want: allUpdateInput(fastly.NewNullable("new10")),
		},
		{
			name: "reset placement to null",
			cmd:  updateCommandPlacementReset(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			want: allUpdateInput(fastly.NullValue[string]()),
		},
		{
			name:      "error missing serviceID",
			cmd:       updateCommandMissingServiceID(),
			wantError: errors.ErrNoServiceID.Error(),
		},
	}
	for i := range scenarios {
		testcase := &scenarios[i]
		t.Run(testcase.name, func(t *testing.T) {
			if testcase.wantError == errors.ErrNoServiceID.Error() {
				t.Setenv("FASTLY_SERVICE_ID", "")
			}
			testcase.cmd.Globals.APIClient = testcase.api

			serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.api,
				Manifest:           testcase.cmd.Manifest,
				Out:                &bytes.Buffer{},
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        true,
			})
			if err != nil {
				if testcase.wantError == "" {
					t.Fatalf("unexpected error getting service details: %v", err)
				}
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}
			if testcase.wantError != "" {
				t.Fatalf("expected error, have nil (service details: %s, %v)", serviceID, serviceVersion.Number)
			}

			have, err := testcase.cmd.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
			testutil.AssertNoError(t, err)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func allUpdateInput(placement *fastly.Nullable[string]) *fastly.UpdateFOSInput {
	return &fastly.UpdateFOSInput{
		ServiceID:         "123",
		ServiceVersion:    4,
		Name:              "log",
		NewName:           new("new1"),
		BucketName:        new("new2"),
		Region:            new("new3"),
		Path:              new("new4"),
		Period:            new(3601),
		GzipLevel:         new(0),
		Format:            new("new5"),
		FormatVersion:     new(3),
		MessageType:       new("new6"),
		ResponseCondition: new("new7"),
		TimestampFormat:   new("new8"),
		Placement:         placement,
		PublicKey:         new("new9"),
		CompressionCodec:  new("new11"),
		FileMaxBytes:      new(2097152),
	}
}

func createCommandBase() *fos.CreateCommand {
	g := newGlobals()
	g.APIClient, _ = mock.APIClient(mock.API{
		GetVersionFn:   testutil.GetVersion,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &fos.CreateCommand{
		Base:     argparser.Base{Globals: g},
		Manifest: manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: autoClone(),
	}
}

func createCommandRequired() *fos.CreateCommand {
	c := createCommandBase()
	c.EndpointName = optString("log")
	c.BucketName = optString("bucket")
	return c
}

func createCommandAll() *fos.CreateCommand {
	c := createCommandBase()
	c.EndpointName = optString("logs")
	c.BucketName = optString("bucket")
	c.Region = optString("us-east")
	c.Path = optString("path")
	c.Period = optInt(3600)
	c.Format = optString(`%h %l %u %t "%r" %>s %b`)
	c.MessageType = optString("classic")
	c.FormatVersion = optInt(2)
	c.ResponseCondition = optString("Prevent default logging")
	c.TimestampFormat = optString("%Y-%m-%dT%H:%M:%S.000")
	c.Placement = optString("none")
	c.PublicKey = optString("public-key")
	c.CompressionCodec = optString("zstd")
	c.FileMaxBytes = optInt(1048576)
	return c
}

func createCommandCodecAndGzip() *fos.CreateCommand {
	c := createCommandRequired()
	c.CompressionCodec = optString("zstd")
	c.GzipLevel = optInt(9)
	return c
}

func createCommandMissingServiceID() *fos.CreateCommand {
	c := createCommandRequired()
	c.Manifest = manifest.Data{}
	c.ServiceVersion = argparser.OptionalServiceVersion{}
	return c
}

func updateCommandNoUpdates() *fos.UpdateCommand {
	return &fos.UpdateCommand{
		Base:         argparser.Base{Globals: newGlobals()},
		Manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		ServiceVersion: argparser.OptionalServiceVersion{
			OptionalString: argparser.OptionalString{Value: "1"},
		},
		AutoClone: autoClone(),
	}
}

func updateCommandAll() *fos.UpdateCommand {
	c := updateCommandNoUpdates()
	c.NewName = optString("new1")
	c.BucketName = optString("new2")
	c.Region = optString("new3")
	c.Path = optString("new4")
	c.Period = optInt(3601)
	c.GzipLevel = optInt(0)
	c.Format = optString("new5")
	c.FormatVersion = optInt(3)
	c.MessageType = optString("new6")
	c.ResponseCondition = optString("new7")
	c.TimestampFormat = optString("new8")
	c.Placement = optString("new10")
	c.PublicKey = optString("new9")
	c.CompressionCodec = optString("new11")
	c.FileMaxBytes = optInt(2097152)
	return c
}

func updateCommandPlacementReset() *fos.UpdateCommand {
	c := updateCommandAll()
	c.Placement = optString("")
	return c
}

func updateCommandMissingServiceID() *fos.UpdateCommand {
	c := updateCommandAll()
	c.Manifest = manifest.Data{}
	c.ServiceVersion = argparser.OptionalServiceVersion{}
	return c
}
