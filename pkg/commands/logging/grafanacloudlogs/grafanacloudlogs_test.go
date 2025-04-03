package grafanacloudlogs_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/grafanacloudlogs"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateGrafanaCloudLogsInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *grafanacloudlogs.CreateCommand
		want      *fastly.CreateGrafanaCloudLogsInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateGrafanaCloudLogsInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           fastly.ToPointer("log"),
				User:           fastly.ToPointer("123456"),
				Index:          fastly.ToPointer("{\"label\": \"value\"}"),
				URL:            fastly.ToPointer("https://test123.grafana.net"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateGrafanaCloudLogsInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				FormatVersion:     fastly.ToPointer(2),
				Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
				MessageType:       fastly.ToPointer("classic"),
				ResponseCondition: fastly.ToPointer("Prevent default logging"),
				Placement:         fastly.ToPointer("none"),
				Name:              fastly.ToPointer("log"),
				User:              fastly.ToPointer("123456"),
				Index:             fastly.ToPointer("{\"label\": \"value\"}"),
				URL:               fastly.ToPointer("https://test123.grafana.net"),
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

func TestUpdateGrafanaCloudLogsInput(t *testing.T) {
	scenarios := []struct {
		name      string
		cmd       *grafanacloudlogs.UpdateCommand
		api       mock.API
		want      *fastly.UpdateGrafanaCloudLogsInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				GetGrafanaCloudLogsFn: getGrafanaCloudLogsOK,
			},
			want: &fastly.UpdateGrafanaCloudLogsInput{
				ServiceID:      "123",
				ServiceVersion: 4,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn:        testutil.ListVersions,
				CloneVersionFn:        testutil.CloneVersionResult(4),
				GetGrafanaCloudLogsFn: getGrafanaCloudLogsOK,
			},
			want: &fastly.UpdateGrafanaCloudLogsInput{
				ServiceID:         "123",
				ServiceVersion:    4,
				Name:              "log",
				NewName:           fastly.ToPointer("new1"),
				User:              fastly.ToPointer("new3"),
				FormatVersion:     fastly.ToPointer(3),
				Format:            fastly.ToPointer("new6"),
				ResponseCondition: fastly.ToPointer("new7"),
				Placement:         fastly.ToPointer("new9"),
				MessageType:       fastly.ToPointer("new10"),
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

func createCommandRequired() *grafanacloudlogs.CreateCommand {
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

	return &grafanacloudlogs.CreateCommand{
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
		User:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "123456"},
		Index:        argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "{\"label\": \"value\"}"},
		URL:          argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "https://test123.grafana.net"},
	}
}

func createCommandAll() *grafanacloudlogs.CreateCommand {
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

	return &grafanacloudlogs.CreateCommand{
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
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 2},
		TimestampFormat:   argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "none"},
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "123456"},
		Index:             argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "{\"label\": \"value\"}"},
		URL:               argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "https://test123.grafana.net"},
	}
}

func createCommandMissingServiceID() *grafanacloudlogs.CreateCommand {
	res := createCommandAll()
	res.Manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *grafanacloudlogs.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &grafanacloudlogs.UpdateCommand{
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

func updateCommandAll() *grafanacloudlogs.UpdateCommand {
	var b bytes.Buffer

	g := global.Data{
		Config: config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &grafanacloudlogs.UpdateCommand{
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
		User:              argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new3"},
		Format:            argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new6"},
		FormatVersion:     argparser.OptionalInt{Optional: argparser.Optional{WasSet: true}, Value: 3},
		ResponseCondition: argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new7"},
		Placement:         argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new9"},
		MessageType:       argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "new10"},
	}
}

func updateCommandMissingServiceID() *grafanacloudlogs.UpdateCommand {
	res := updateCommandAll()
	res.Manifest = manifest.Data{}
	return res
}
