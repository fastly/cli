package papertrail_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/commands/logging/papertrail"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestCreatePapertrailInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *papertrail.CreateCommand
		want      *fastly.CreatePapertrailInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreatePapertrailInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
				Address:        "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreatePapertrailInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				Address:           "example.com",
				Port:              22,
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
				Client:             testcase.cmd.Base.Globals.Client,
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

func TestUpdatePapertrailInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *papertrail.UpdateCommand
		api       mock.API
		want      *fastly.UpdatePapertrailInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				GetPapertrailFn: getPapertrailOK,
			},
			want: &fastly.UpdatePapertrailInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				GetPapertrailFn: getPapertrailOK,
			},
			want: &fastly.UpdatePapertrailInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Uint(23),
				Format:            fastly.String("new3"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new4"),
				Placement:         fastly.String("new5"),
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

			var bs []byte
			out := bytes.NewBuffer(bs)
			verboseMode := true

			serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
				AutoCloneFlag:      testcase.cmd.AutoClone,
				Client:             testcase.api,
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

func createCommandRequired() *papertrail.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.Client, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &papertrail.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		Address:      "example.com",
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

func createCommandAll() *papertrail.CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.Client, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &papertrail.CreateCommand{
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
		Address:           "example.com",
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		Port:              cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 22},
	}
}

func createCommandMissingServiceID() *papertrail.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *papertrail.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &papertrail.UpdateCommand{
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

func updateCommandAll() *papertrail.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &papertrail.UpdateCommand{
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
		Port:              cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 23},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
	}
}

func updateCommandMissingServiceID() *papertrail.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
