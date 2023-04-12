package elasticsearch_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/elasticsearch"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v8/fastly"
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
				Name:           fastly.String("log"),
				Index:          fastly.String("logs"),
				URL:            fastly.String("example.com"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              fastly.String("logs"),
				ResponseCondition: fastly.String("Prevent default logging"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				Index:             fastly.String("logs"),
				URL:               fastly.String("example.com"),
				Pipeline:          fastly.String("my_pipeline_id"),
				User:              fastly.String("user"),
				Password:          fastly.String("password"),
				RequestMaxEntries: fastly.Int(2),
				RequestMaxBytes:   fastly.Int(2),
				Placement:         fastly.String("none"),
				TLSCACert:         fastly.String("-----BEGIN CERTIFICATE-----foo"),
				TLSHostname:       fastly.String("example.com"),
				TLSClientCert:     fastly.String("-----BEGIN CERTIFICATE-----bar"),
				TLSClientKey:      fastly.String("-----BEGIN PRIVATE KEY-----bar"),
				FormatVersion:     fastly.Int(2),
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
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				GetElasticsearchFn: getElasticsearchOK,
			},
			want: &fastly.UpdateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Index:             fastly.String("new2"),
				URL:               fastly.String("new3"),
				Pipeline:          fastly.String("new4"),
				User:              fastly.String("new5"),
				Password:          fastly.String("new6"),
				RequestMaxEntries: fastly.Int(3),
				RequestMaxBytes:   fastly.Int(3),
				Placement:         fastly.String("new7"),
				Format:            fastly.String("new8"),
				FormatVersion:     fastly.Int(3),
				ResponseCondition: fastly.String("new9"),
				TLSCACert:         fastly.String("new10"),
				TLSClientCert:     fastly.String("new11"),
				TLSClientKey:      fastly.String("new12"),
				TLSHostname:       fastly.String("new13"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
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

func createCommandRequired() *elasticsearch.CreateCommand {
	var b bytes.Buffer

	globals := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.APIClient, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &elasticsearch.CreateCommand{
		Base: cmd.Base{
			Globals: &globals,
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
		Index:        cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "logs"},
		URL:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
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
		ListVersionsFn: testutil.ListVersions,
		CloneVersionFn: testutil.CloneVersionResult(4),
	})("token", "endpoint")

	return &elasticsearch.CreateCommand{
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
		EndpointName:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "logs"},
		Index:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "logs"},
		URL:               cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
		Pipeline:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "my_pipeline_id"},
		RequestMaxEntries: cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		RequestMaxBytes:   cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
		User:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "user"},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "password"},
		TLSCACert:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandMissingServiceID() *elasticsearch.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
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

func updateCommandAll() *elasticsearch.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &elasticsearch.UpdateCommand{
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
		Index:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		URL:               cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		Pipeline:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		RequestMaxEntries: cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		RequestMaxBytes:   cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		User:              cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Password:          cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new7"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new8"},
		FormatVersion:     cmd.OptionalInt{Optional: cmd.Optional{WasSet: true}, Value: 3},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new9"},
		TLSCACert:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new10"},
		TLSClientCert:     cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new11"},
		TLSClientKey:      cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new12"},
		TLSHostname:       cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new13"},
	}
}

func updateCommandMissingServiceID() *elasticsearch.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
