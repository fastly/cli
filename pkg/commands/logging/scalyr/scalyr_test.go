package scalyr_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/scalyr"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateScalyrInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *scalyr.CreateCommand
		want      *fastly.CreateScalyrInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateScalyrInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.String("log"),
				Token:          fastly.String("tkn"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateScalyrInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.String("log"),
				Token:             fastly.String("tkn"),
				Region:            fastly.String("US"),
				FormatVersion:     fastly.Int(2),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
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

func TestUpdateScalyrInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *scalyr.UpdateCommand
		api       mock.API
		want      *fastly.UpdateScalyrInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetScalyrFn:    getScalyrOK,
			},
			want: &fastly.UpdateScalyrInput{
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
				GetScalyrFn:    getScalyrOK,
			},
			want: &fastly.UpdateScalyrInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Token:             fastly.String("new2"),
				FormatVersion:     fastly.Int(3),
				Format:            fastly.String("new3"),
				ResponseCondition: fastly.String("new4"),
				Placement:         fastly.String("new5"),
				Region:            fastly.String("new6"),
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

func createCommandRequired() *scalyr.CreateCommand {
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

	return &scalyr.CreateCommand{
		Base: cmd.Base{
			Globals: &g,
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
		EndpointName: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "log"},
		Token:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "tkn"},
	}
}

func createCommandAll() *scalyr.CreateCommand {
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

	return &scalyr.CreateCommand{
		Base: cmd.Base{
			Globals: &g,
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
		Token:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "tkn"},
		Region:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "US"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
	}
}

func createCommandMissingServiceID() *scalyr.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *scalyr.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &scalyr.UpdateCommand{
		Base: cmd.Base{
			Globals: &g,
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

func updateCommandAll() *scalyr.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &scalyr.UpdateCommand{
		Base: cmd.Base{
			Globals: &g,
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
		Token:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Region:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
	}
}

func updateCommandMissingServiceID() *scalyr.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
