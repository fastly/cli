package azureblob_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/azureblob"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
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
				Name:           "logs",
				AccountName:    "account",
				Container:      "container",
				SASToken:       "token",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateBlobStorageInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "logs",
				Container:         "container",
				AccountName:       "account",
				SASToken:          "token",
				Path:              "/log",
				Period:            3600,
				GzipLevel:         0,
				Format:            `%h %l %u %t "%r" %>s %b`,
				MessageType:       "classic",
				FormatVersion:     2,
				ResponseCondition: "Prevent default logging",
				TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
				Placement:         "none",
				PublicKey:         pgpPublicKey(),
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

func TestUpdateBlobStorageInput(t *testing.T) {
	for _, testcase := range []struct {
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
				NewName:           fastly.String("new1"),
				Container:         fastly.String("new2"),
				AccountName:       fastly.String("new3"),
				SASToken:          fastly.String("new4"),
				Path:              fastly.String("new5"),
				Period:            fastly.Uint(3601),
				GzipLevel:         fastly.Uint(0),
				Format:            fastly.String("new6"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new7"),
				MessageType:       fastly.String("new8"),
				TimestampFormat:   fastly.String("new9"),
				Placement:         fastly.String("new10"),
				PublicKey:         fastly.String("new11"),
				CompressionCodec:  fastly.String("new12"),
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
	} {
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

func createCommandRequired() *azureblob.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	// TODO: make consistent (in all other logging files) with syslog_test which
	// uses a testcase.api field to assign the mock API to the global client.
	globals.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &azureblob.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
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
		Container:   "container",
		AccountName: "account",
		SASToken:    "token",
	}
}

func createCommandAll() *azureblob.CreateCommand {
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

	return &azureblob.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
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
		Container:         "container",
		AccountName:       "account",
		SASToken:          "token",
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "/log"},
		Period:            cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3600},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		MessageType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "classic"},
		PublicKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: pgpPublicKey()},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "zstd"},
	}
}

func createCommandMissingServiceID() *azureblob.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *azureblob.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &azureblob.UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
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

func updateCommandAll() *azureblob.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &azureblob.UpdateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
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
		Container:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		AccountName:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		SASToken:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		Path:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Period:            cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 0},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		MessageType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		TimestampFormat:   cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		PublicKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
		CompressionCodec:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new12"},
	}
}

func updateCommandMissingServiceID() *azureblob.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
