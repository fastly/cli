package openstack_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/openstack"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateOpenstackInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *openstack.CreateCommand
		want      *fastly.CreateOpenstackInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateOpenstackInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           new("log"),
				BucketName:     new("bucket"),
				AccessKey:      new("access"),
				User:           new("user"),
				URL:            new("https://example.com"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateOpenstackInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              new("log"),
				BucketName:        new("bucket"),
				AccessKey:         new("access"),
				User:              new("user"),
				URL:               new("https://example.com"),
				Path:              new("/log"),
				Period:            new(3600),
				Format:            new(`%h %l %u %t "%r" %>s %b`),
				MessageType:       new("classic"),
				FormatVersion:     new(2),
				ResponseCondition: new("Prevent default logging"),
				TimestampFormat:   new("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         new("none"),
				PublicKey:         new(pgpPublicKey()),
				CompressionCodec:  new("zstd"),
				ProcessingRegion:  new("eu"),
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

func TestUpdateOpenstackInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *openstack.UpdateCommand
		api       mock.API
		want      *fastly.UpdateOpenstackInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetOpenstackFn: getOpenstackOK,
			},
			want: &fastly.UpdateOpenstackInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				BucketName:        new("new2"),
				User:              new("new3"),
				AccessKey:         new("new4"),
				URL:               new("new5"),
				Path:              new("new6"),
				Period:            new(3601),
				GzipLevel:         new(0),
				Format:            new("new7"),
				FormatVersion:     new(3),
				ResponseCondition: new("new8"),
				MessageType:       new("new9"),
				TimestampFormat:   new("new10"),
				Placement:         fastly.NewNullable("new11"),
				PublicKey:         new("new12"),
				CompressionCodec:  new("new13"),
				ProcessingRegion:  new("eu"),
			},
		},
		{
			name: "reset placement to null",
			cmd:  updateCommandPlacementReset(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetOpenstackFn: getOpenstackOK,
			},
			want: &fastly.UpdateOpenstackInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				BucketName:        new("new2"),
				User:              new("new3"),
				AccessKey:         new("new4"),
				URL:               new("new5"),
				Path:              new("new6"),
				Period:            new(3601),
				GzipLevel:         new(0),
				Format:            new("new7"),
				FormatVersion:     new(3),
				ResponseCondition: new("new8"),
				MessageType:       new("new9"),
				TimestampFormat:   new("new10"),
				Placement:         fastly.NullValue[string](),
				PublicKey:         new("new12"),
				CompressionCodec:  new("new13"),
				ProcessingRegion:  new("eu"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetOpenstackFn: getOpenstackOK,
			},
			want: &fastly.UpdateOpenstackInput{
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

func createCommandRequired() *openstack.CreateCommand {
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

	return &openstack.CreateCommand{
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
		BucketName:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		AccessKey:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "access"},
		User:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		URL:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "https://example.com"},
	}
}

func createCommandAll() *openstack.CreateCommand {
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

	return &openstack.CreateCommand{
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
		BucketName:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bucket"},
		AccessKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "access"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "https://example.com"},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "/log"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3600},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: pgpPublicKey()},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zstd"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
	}
}

func createCommandMissingServiceID() *openstack.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}

func updateCommandNoUpdates() *openstack.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &openstack.UpdateCommand{
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

func updateCommandAll() *openstack.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &openstack.UpdateCommand{
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
		BucketName:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		AccessKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 0},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new13"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
	}
}

func updateCommandPlacementReset() *openstack.UpdateCommand {
	c := updateCommandAll()
	c.Placement = argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: ""}
	return c
}

func updateCommandMissingServiceID() *openstack.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}
