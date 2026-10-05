package elasticsearch_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/elasticsearch"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateElasticsearchInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *elasticsearch.CreateCommand
		want      *fastly.CreateElasticsearchInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateElasticsearchInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           new("log"),
				Index:          new("logs"),
				URL:            new("example.com"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              new("logs"),
				ResponseCondition: new("Prevent default logging"),
				Format:            new(`%h %l %u %t "%r" %>s %b`),
				Index:             new("logs"),
				URL:               new("example.com"),
				Pipeline:          new("my_pipeline_id"),
				User:              new("user"),
				Password:          new("password"),
				ProcessingRegion:  new("eu"),
				RequestMaxEntries: new(2),
				RequestMaxBytes:   new(2),
				Placement:         new("none"),
				TLSCACert:         new("-----BEGIN CERTIFICATE-----foo"),
				TLSHostname:       new("example.com"),
				TLSClientCert:     new("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      new("-----BEGIN PRIVATE KEY-----bar"),
				FormatVersion:     new(2),
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

func TestUpdateElasticsearchInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *elasticsearch.UpdateCommand
		api       mock.API
		want      *fastly.UpdateElasticsearchInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				GetElasticsearchFn: getElasticsearchOK,
			},
			want: &fastly.UpdateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				Index:             new("new2"),
				URL:               new("new3"),
				Pipeline:          new("new4"),
				User:              new("new5"),
				Password:          new("new6"),
				RequestMaxEntries: new(3),
				RequestMaxBytes:   new(3),
				Placement:         fastly.NewNullable("new7"),
				Format:            new("new8"),
				FormatVersion:     new(3),
				ProcessingRegion:  new("eu"),
				ResponseCondition: new("new9"),
				TLSCACert:         new("new10"),
				TLSClientCert:     new("new11"),
				TLSClientKey:      new("new12"),
				TLSHostname:       new("new13"),
			},
		},
		{
			name: "reset placement to null",
			cmd:  updateCommandPlacementReset(),
			api: mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				GetElasticsearchFn: getElasticsearchOK,
			},
			want: &fastly.UpdateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           new("new1"),
				Index:             new("new2"),
				URL:               new("new3"),
				Pipeline:          new("new4"),
				User:              new("new5"),
				Password:          new("new6"),
				RequestMaxEntries: new(3),
				RequestMaxBytes:   new(3),
				Placement:         fastly.NullValue[string](),
				Format:            new("new8"),
				FormatVersion:     new(3),
				ProcessingRegion:  new("eu"),
				ResponseCondition: new("new9"),
				TLSCACert:         new("new10"),
				TLSClientCert:     new("new11"),
				TLSClientKey:      new("new12"),
				TLSHostname:       new("new13"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				GetVersionFn:       testutil.GetVersion,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				GetElasticsearchFn: getElasticsearchOK,
			},
			want: &fastly.UpdateElasticsearchInput{
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

func createCommandRequired() *elasticsearch.CreateCommand {
	var b bytes.Buffer

	globals := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.APIClient(mock.API{
		GetVersionFn:   testutil.GetVersion,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint", false)

	return &elasticsearch.CreateCommand{
		Base: argparser.Base{
			Globals: &globals,
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
		Index:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		URL:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
	}
}

func createCommandAll() *elasticsearch.CreateCommand {
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

	return &elasticsearch.CreateCommand{
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
		Index:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "logs"},
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		Pipeline:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "my_pipeline_id"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
		RequestMaxEntries: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		RequestMaxBytes:   argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "user"},
		Password:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "password"},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandMissingServiceID() *elasticsearch.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}

func updateCommandNoUpdates() *elasticsearch.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &elasticsearch.UpdateCommand{
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

func updateCommandAll() *elasticsearch.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &elasticsearch.UpdateCommand{
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
		Index:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new2"},
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		Pipeline:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new4"},
		RequestMaxEntries: argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		RequestMaxBytes:   argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new5"},
		Password:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		ProcessingRegion:  argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "eu"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new8"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		TLSCACert:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
		TLSClientCert:     argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new11"},
		TLSClientKey:      argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new12"},
		TLSHostname:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new13"},
	}
}

func updateCommandPlacementReset() *elasticsearch.UpdateCommand {
	c := updateCommandAll()
	c.Placement = argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: ""}
	return c
}

func updateCommandMissingServiceID() *elasticsearch.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	res.ServiceVersion = argparser.OptionalServiceVersion{}
	return res
}
