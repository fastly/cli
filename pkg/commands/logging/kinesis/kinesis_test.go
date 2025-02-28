package kinesis_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/kinesis"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateKinesisInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *kinesis.CreateCommand
		want      *fastly.CreateKinesisInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateKinesisInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				StreamName:     fastly.ToPointer("stream"),
				Region:         fastly.ToPointer("us-east-1"),
				AccessKey:      fastly.ToPointer("access"),
				SecretKey:      fastly.ToPointer("secret"),
			},
		},
		{
			name: "required values set flag serviceID using IAM role",
			cmd:  createCommandRequiredIAMRole(),
			want: &fastly.CreateKinesisInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				Region:         fastly.ToPointer("us-east-1"),
				StreamName:     fastly.ToPointer("stream"),
				IAMRole:        fastly.ToPointer("arn:aws:iam::123456789012:role/KinesisAccess"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateKinesisInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.ToPointer("logs"),
				StreamName:        fastly.ToPointer("stream"),
				Region:            fastly.ToPointer("us-east-1"),
				AccessKey:         fastly.ToPointer("access"),
				SecretKey:         fastly.ToPointer("secret"),
				Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.ToPointer(2),
				ResponseCondition: fastly.ToPointer("Prevent default logging"),
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

func TestUpdateKinesisInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *kinesis.UpdateCommand
		api       mock.API
		want      *fastly.UpdateKinesisInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetKinesisFn:   getKinesisOK,
			},
			want: &fastly.UpdateKinesisInput{
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
				GetKinesisFn:   getKinesisOK,
			},
			want: &fastly.UpdateKinesisInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.ToPointer("new1"),
				StreamName:        fastly.ToPointer("new2"),
				AccessKey:         fastly.ToPointer("new3"),
				SecretKey:         fastly.ToPointer("new4"),
				IAMRole:           fastly.ToPointer(""),
				Region:            fastly.ToPointer("new5"),
				Format:            fastly.ToPointer("new7"),
				FormatVersion:     fastly.ToPointer(3),
				ResponseCondition: fastly.ToPointer("new9"),
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

func createCommandRequired() *kinesis.CreateCommand {
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

	return &kinesis.CreateCommand{
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
		Region:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "us-east-1"},
		StreamName:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "stream"},
		AccessKey:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "access"},
		SecretKey:    argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "secret"},
	}
}

func createCommandRequiredIAMRole() *kinesis.CreateCommand {
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

	return &kinesis.CreateCommand{
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
		EndpointName: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Region:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "us-east-1"},
		StreamName:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "stream"},
		AutoClone: argparser.OptionalAutoClone{
			OptionalBool: argparser.OptionalBool{
				Optional: argparser.Optional{
					WasSet: true,
				},
				Value: true,
			},
		},
		IAMRole: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "arn:aws:iam::123456789012:role/KinesisAccess"},
	}
}

func createCommandAll() *kinesis.CreateCommand {
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

	return &kinesis.CreateCommand{
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
		StreamName:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "stream"},
		Region:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "us-east-1"},
		AccessKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "access"},
		SecretKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "secret"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
	}
}

func createCommandMissingServiceID() *kinesis.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *kinesis.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kinesis.UpdateCommand{
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

func updateCommandAll() *kinesis.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kinesis.UpdateCommand{
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
		StreamName:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		AccessKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		SecretKey:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		IAMRole:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: ""},
		Region:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
	}
}

func updateCommandMissingServiceID() *kinesis.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
