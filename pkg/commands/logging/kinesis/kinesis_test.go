package kinesis_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/kinesis"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
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
				Name:           "log",
				StreamName:     "stream",
				AccessKey:      "access",
				SecretKey:      "secret",
			},
		},
		{
			name: "required values set flag serviceID using IAM role",
			cmd:  createCommandRequiredIAMRole(),
			want: &fastly.CreateKinesisInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
				StreamName:     "stream",
				IAMRole:        "arn:aws:iam::123456789012:role/KinesisAccess",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateKinesisInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "logs",
				StreamName:        "stream",
				Region:            "us-east-1",
				AccessKey:         "access",
				SecretKey:         "secret",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				ResponseCondition: "Prevent default logging",
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
			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				APIClient:          testcase.cmd.Base.Globals.APIClient,
				Manifest:           testcase.cmd.Manifest,
				Out:                out,
				ServiceVersionFlag: testcase.cmd.ServiceVersion,
				VerboseMode:        verboseMode,
			})
			if err != nil {
				if testcase.wantError == "" {
					t.Fatalf("unexpected error getting service details: %v", err)
				}
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}
			if err == nil {
				if testcase.wantError != "" {
					t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
				}
			}

			have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func TestUpdateKinesisInput(t *testing.T) {
	for _, testcase := range []struct {
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
				NewName:           fastly.String("new1"),
				StreamName:        fastly.String("new2"),
				AccessKey:         fastly.String("new3"),
				SecretKey:         fastly.String("new4"),
				IAMRole:           fastly.String(""),
				Region:            fastly.String("new5"),
				Format:            fastly.String("new7"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new9"),
				Placement:         fastly.String("new11"),
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
			testcase.cmd.Base.Globals.APIClient = testcase.api

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
			if err != nil {
				if testcase.wantError == "" {
					t.Fatalf("unexpected error getting service details: %v", err)
				}
				testutil.AssertErrorContains(t, err, testcase.wantError)
				return
			}
			if err == nil {
				if testcase.wantError != "" {
					t.Fatalf("expected error, have nil (service details: %s, %d)", serviceID, serviceVersion.Number)
				}
			}

			have, err := testcase.cmd.ConstructInput(serviceID, serviceVersion.Number)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func createCommandRequired() *kinesis.CreateCommand {
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

	return &kinesis.CreateCommand{
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
		StreamName: "stream",
		AccessKey:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "access"},
		SecretKey:  cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "secret"},
	}
}

func createCommandRequiredIAMRole() *kinesis.CreateCommand {
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

	return &kinesis.CreateCommand{
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
		StreamName: "stream",
		IAMRole:    cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "arn:aws:iam::123456789012:role/KinesisAccess"},
	}
}

func createCommandAll() *kinesis.CreateCommand {
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

	return &kinesis.CreateCommand{
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
		StreamName:        "stream",
		AccessKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "access"},
		SecretKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "secret"},
		Region:            "us-east-1",
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
	}
}

func createCommandMissingServiceID() *kinesis.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *kinesis.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kinesis.UpdateCommand{
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

func updateCommandAll() *kinesis.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &kinesis.UpdateCommand{
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
		StreamName:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		AccessKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		SecretKey:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		IAMRole:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: ""},
		Region:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
	}
}

func updateCommandMissingServiceID() *kinesis.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
