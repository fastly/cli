package azureblob_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/azureblob"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateBlobStorageInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *azureblob.CreateCommand
		want      *fastly.CreateBlobStorageInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateBlobStorageInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("logs"),
				AccountName:    fastly.ToPointer("account"),
				Container:      fastly.ToPointer("container"),
				SASToken:       fastly.ToPointer("token"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateBlobStorageInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.ToPointer("logs"),
				Container:         fastly.ToPointer("container"),
				AccountName:       fastly.ToPointer("account"),
				SASToken:          fastly.ToPointer("token"),
				Path:              fastly.ToPointer("/log"),
				Period:            fastly.ToPointer(3600),
				Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				MessageType:       fastly.ToPointer("classic"),
				FormatVersion:     fastly.ToPointer(2),
				ResponseCondition: fastly.ToPointer("Prevent default logging"),
				TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
				Placement:         fastly.ToPointer("none"),
				PublicKey:         fastly.ToPointer(pgpPublicKey()),
				CompressionCodec:  fastly.ToPointer("zstd"),
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

func TestUpdateBlobStorageInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *azureblob.UpdateCommand
		api       mock.API
		want      *fastly.UpdateBlobStorageInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				GetBlobStorageFn: getBlobStorageOK,
			},
			want: &fastly.UpdateBlobStorageInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "logs",
				NewName:           fastly.ToPointer("new1"),
				Container:         fastly.ToPointer("new2"),
				AccountName:       fastly.ToPointer("new3"),
				SASToken:          fastly.ToPointer("new4"),
				Path:              fastly.ToPointer("new5"),
				Period:            fastly.ToPointer(3601),
				GzipLevel:         fastly.ToPointer(0),
				Format:            fastly.ToPointer("new6"),
				FormatVersion:     fastly.ToPointer(3),
				ResponseCondition: fastly.ToPointer("new7"),
				MessageType:       fastly.ToPointer("new8"),
				TimestampFormat:   fastly.ToPointer("new9"),
				Placement:         fastly.ToPointer("new10"),
				PublicKey:         fastly.ToPointer("new11"),
				CompressionCodec:  fastly.ToPointer("new12"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				CloneVersionFn:   testutil.CloneVersionResult(4),
				GetBlobStorageFn: getBlobStorageOK,
			},
			want: &fastly.UpdateBlobStorageInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "logs",
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

func createCommandRequired() *azureblob.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	// TODO: make consistent (in all other logging files) with syslog_test which
	// uses a testcase.api field to assign the mock API to the global client.
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &azureblob.CreateCommand{
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
		EndpointName: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Container:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "container"},
		AccountName:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "account"},
		SASToken:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "token"},
	}
}

func createCommandAll() *azureblob.CreateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	g.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &azureblob.CreateCommand{
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
		EndpointName:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		Container:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "container"},
		AccountName:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "account"},
		SASToken:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "token"},
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
	}
}

func createCommandMissingServiceID() *azureblob.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *azureblob.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &azureblob.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
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

func updateCommandAll() *azureblob.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &azureblob.UpdateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
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
		Container:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		AccountName:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		SASToken:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		Path:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 0},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		PublicKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
	}
}

func updateCommandMissingServiceID() *azureblob.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
