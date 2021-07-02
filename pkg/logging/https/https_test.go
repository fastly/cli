package https_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/logging/https"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
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
				Name:           "log",
				URL:            "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateHTTPSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "logs",
				ResponseCondition: "Prevent default logging",
				Format:            `%h %l %u %t "%r" %>s %b`,
				URL:               "example.com",
				RequestMaxEntries: 2,
				RequestMaxBytes:   2,
				ContentType:       "application/json",
				HeaderName:        "name",
				HeaderValue:       "value",
				Method:            "GET",
				JSONFormat:        "1",
				Placement:         "none",
				TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
				TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
				TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
				TLSHostname:       "example.com",
				MessageType:       "classic",
				FormatVersion:     2,
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

func TestUpdateHTTPSInput(t *testing.T) {
	for _, testcase := range []struct {
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
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetHTTPSFn:     getHTTPSOK,
			},
			want: &fastly.UpdateHTTPSInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.String("new1"),
				ResponseCondition: fastly.String("new2"),
				Format:            fastly.String("new3"),
				URL:               fastly.String("new4"),
				RequestMaxEntries: fastly.Uint(3),
				RequestMaxBytes:   fastly.Uint(3),
				ContentType:       fastly.String("new5"),
				HeaderName:        fastly.String("new6"),
				HeaderValue:       fastly.String("new7"),
				Method:            fastly.String("new8"),
				JSONFormat:        fastly.String("new9"),
				Placement:         fastly.String("new10"),
				TLSCACert:         fastly.String("new11"),
				TLSClientCert:     fastly.String("new12"),
				TLSClientKey:      fastly.String("new13"),
				TLSHostname:       fastly.String("new14"),
				MessageType:       fastly.String("new15"),
				FormatVersion:     fastly.Uint(3),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
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

func createCommandRequired() *https.CreateCommand {
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

	return &https.CreateCommand{
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
		URL: "example.com",
	}
}

func createCommandAll() *https.CreateCommand {
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

	return &https.CreateCommand{
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
		URL:               "example.com",
		ContentType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "application/json"},
		HeaderName:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "name"},
		HeaderValue:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "value"},
		Method:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "GET"},
		JSONFormat:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "1"},
		MessageType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "classic"},
		RequestMaxEntries: cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		RequestMaxBytes:   cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		TLSCACert:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandMissingServiceID() *https.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *https.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &https.UpdateCommand{
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

func updateCommandAll() *https.UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &https.UpdateCommand{
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
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		URL:               cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		ContentType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		HeaderName:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		HeaderValue:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		Method:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		JSONFormat:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		RequestMaxEntries: cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		RequestMaxBytes:   cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		TLSCACert:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
		TLSClientCert:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new12"},
		TLSClientKey:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new13"},
		TLSHostname:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new14"},
		MessageType:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new15"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
	}
}

func updateCommandMissingServiceID() *https.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
