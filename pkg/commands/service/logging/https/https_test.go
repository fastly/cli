package https_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/https"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateHTTPSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *https.CreateCommand
		want      *fastly.CreateHTTPSInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateHTTPSInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           new("log"),
				URL:            new("example.com"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateHTTPSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              new("logs"),
				ResponseCondition: new("Prevent default logging"),
				Format:            new(`%h %l %u %t "%r" %>s %b`),
				URL:               new("example.com"),
				RequestMaxEntries: new(2),
				RequestMaxBytes:   new(2),
				ContentType:       new("application/json"),
				HeaderName:        new("name"),
				HeaderValue:       new("value"),
				Method:            new(http.MethodGet),
				JSONFormat:        new("1"),
				Period:            new(5),
				Placement:         new("none"),
				TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
				TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
				TLSHostname:       new("example.com"),
				MessageType:       new("classic"),
				FormatVersion:     new(2),
				ProcessingRegion:  new("eu"),
				CompressionCodec:  new("zstd"),
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

func TestUpdateHTTPSInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *https.UpdateCommand
		api       mock.API
		want      *fastly.UpdateHTTPSInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetHTTPSFn:     getHTTPSOK,
			},
			want: &fastly.UpdateHTTPSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				ResponseCondition: new("new2"),
				Format:            new("new3"),
				URL:               new("new4"),
				RequestMaxEntries: new(3),
				RequestMaxBytes:   new(3),
				ContentType:       new("new5"),
				HeaderName:        new("new6"),
				HeaderValue:       new("new7"),
				Method:            new("new8"),
				JSONFormat:        new("new9"),
				Period:            new(5),
				Placement:         fastly.NewNullable("new10"),
				TLSCACert:         new("new11"),
				TLSClientCert:     new("new12"),
				TLSClientKey:      new("new13"),
				TLSHostname:       new("new14"),
				MessageType:       new("new15"),
				FormatVersion:     new(3),
				ProcessingRegion:  new("eu"),
			},
		},
		{
			name: "reset placement to null",
			cmd:  updateCommandPlacementReset(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetHTTPSFn:     getHTTPSOK,
			},
			want: &fastly.UpdateHTTPSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				ResponseCondition: new("new2"),
				Format:            new("new3"),
				URL:               new("new4"),
				RequestMaxEntries: new(3),
				RequestMaxBytes:   new(3),
				ContentType:       new("new5"),
				HeaderName:        new("new6"),
				HeaderValue:       new("new7"),
				Method:            new("new8"),
				JSONFormat:        new("new9"),
				Period:            new(5),
				Placement:         fastly.NullValue[string](),
				TLSCACert:         new("new11"),
				TLSClientCert:     new("new12"),
				TLSClientKey:      new("new13"),
				TLSHostname:       new("new14"),
				MessageType:       new("new15"),
				FormatVersion:     new(3),
				ProcessingRegion:  new("eu"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetHTTPSFn:     getHTTPSOK,
			},
			want: &fastly.UpdateHTTPSInput{
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

func createCommandRequired() *https.CreateCommand {
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

	return &https.CreateCommand{
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
		URL:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
	}
}

func createCommandAll() *https.CreateCommand {
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

	return &https.CreateCommand{
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
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		ContentType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "application/json"},
		HeaderName:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "name"},
		HeaderValue:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "value"},
		Method:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: http.MethodGet},
		JSONFormat:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "1"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
		RequestMaxEntries: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		RequestMaxBytes:   argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 5},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zstd"},
	}
}

func createCommandMissingServiceID() *https.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}

func updateCommandNoUpdates() *https.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &https.UpdateCommand{
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

func updateCommandAll() *https.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &https.UpdateCommand{
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
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		ContentType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		HeaderName:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		HeaderValue:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		Method:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		JSONFormat:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		Period:            argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 5},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		RequestMaxEntries: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		RequestMaxBytes:   argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new13"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new14"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new15"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
		CompressionCodec:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "zstd"},
	}
}

func updateCommandPlacementReset() *https.UpdateCommand {
	c := updateCommandAll()
	c.Placement = argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: ""}
	return c
}

func updateCommandMissingServiceID() *https.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}
