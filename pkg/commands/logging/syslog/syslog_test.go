package syslog_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/syslog"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateSyslogInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *syslog.CreateCommand
		want      *fastly.CreateSyslogInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateSyslogInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				Address:        fastly.ToPointer("example.com"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateSyslogInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.ToPointer("log"),
				Address:           fastly.ToPointer("example.com"),
				Port:              fastly.ToPointer(22),
				UseTLS:            fastly.ToPointer(fastly.Compatibool(true)),
				TLSCACert:         fastly.ToPointer("-----BEGIN CERTIFICATE-----foo"),
				TLSHostname:       fastly.ToPointer("example.com"),
				TLSClientCert:     fastly.ToPointer("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      fastly.ToPointer("-----BEGIN PRIVATE KEY-----bar"),
				Token:             fastly.ToPointer("tkn"),
				Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.ToPointer(2),
				MessageType:       fastly.ToPointer("classic"),
				ResponseCondition: fastly.ToPointer("Prevent default logging"),
				Placement:         fastly.ToPointer("none"),
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

func TestUpdateSyslogInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *syslog.UpdateCommand
		api       mock.API
		want      *fastly.UpdateSyslogInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetSyslogFn:    getSyslogOK,
			},
			want: &fastly.UpdateSyslogInput{
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
				GetSyslogFn:    getSyslogOK,
			},
			want: &fastly.UpdateSyslogInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.ToPointer("new1"),
				Address:           fastly.ToPointer("new2"),
				Port:              fastly.ToPointer(23),
				UseTLS:            fastly.ToPointer(fastly.Compatibool(false)),
				TLSCACert:         fastly.ToPointer("new3"),
				TLSHostname:       fastly.ToPointer("new4"),
				TLSClientCert:     fastly.ToPointer("new5"),
				TLSClientKey:      fastly.ToPointer("new6"),
				Token:             fastly.ToPointer("new7"),
				Format:            fastly.ToPointer("new8"),
				FormatVersion:     fastly.ToPointer(3),
				MessageType:       fastly.ToPointer("new9"),
				ResponseCondition: fastly.ToPointer("new10"),
				Placement:         fastly.ToPointer("new11"),
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

func createCommandRequired() *syslog.CreateCommand {
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

	return &syslog.CreateCommand{
		Base: argparser.Base{
			Globals: &g,
		},
		Manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "log"},
		Address:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
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

func createCommandAll() *syslog.CreateCommand {
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

	return &syslog.CreateCommand{
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
		Address:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		Port:              argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 22},
		UseTLS:            argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: true},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
		Token:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "tkn"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
	}
}

func createCommandMissingServiceID() *syslog.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *syslog.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &syslog.UpdateCommand{
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

func updateCommandAll() *syslog.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &syslog.UpdateCommand{
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
		Address:           argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		Port:              argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 23},
		UseTLS:            argparser.OptionalBool{Optional: argparser.Optional{WasSet: true}, Value: false},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		Token:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
	}
}

func updateCommandMissingServiceID() *syslog.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
