package elasticsearch

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestCreateElasticsearchInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateElasticsearchInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateElasticsearchInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Index:          "logs",
				URL:            "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				ResponseCondition: "Prevent default logging",
				Format:            `%h %l %u %t "%r" %>s %b`,
				Index:             "logs",
				URL:               "example.com",
				Pipeline:          "my_pipeline_id",
				User:              "user",
				Password:          "password",
				RequestMaxEntries: 2,
				RequestMaxBytes:   2,
				Placement:         "none",
				TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
				TLSHostname:       "example.com",
				TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
				TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
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
			have, err := testcase.cmd.createInput()
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func TestUpdateElasticsearchInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateElasticsearchInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersionsOk,
				GetVersionFn:       testutil.GetActiveVersionOK,
				CloneVersionFn:     testutil.CloneVersionOK,
				GetElasticsearchFn: getElasticsearchOK,
			},
			want: &fastly.UpdateElasticsearchInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Index:             fastly.String("new2"),
				URL:               fastly.String("new3"),
				Pipeline:          fastly.String("new4"),
				User:              fastly.String("new5"),
				Password:          fastly.String("new6"),
				RequestMaxEntries: fastly.Uint(3),
				RequestMaxBytes:   fastly.Uint(3),
				Placement:         fastly.String("new7"),
				Format:            fastly.String("new8"),
				FormatVersion:     fastly.Uint(3),
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
				ListVersionsFn:     testutil.ListVersionsOk,
				GetVersionFn:       testutil.GetActiveVersionOK,
				CloneVersionFn:     testutil.CloneVersionOK,
				GetElasticsearchFn: getElasticsearchOK,
			},
			want: &fastly.UpdateElasticsearchInput{
				ServiceID:      "123",
				ServiceVersion: 2,
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

			have, err := testcase.cmd.createInput()
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func createCommandRequired() *CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.Client, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersionsOk,
		GetVersionFn:   testutil.GetActiveVersionOK,
		CloneVersionFn: testutil.CloneVersionOK,
	})("token", "endpoint")

	return &CreateCommand{
		Base: common.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		serviceVersion: common.OptionalServiceVersion{
			OptionalString: common.OptionalString{Value: "1"},
		},
		autoClone: common.OptionalAutoClone{
			OptionalBool: common.OptionalBool{Value: true},
		},
		Index: "logs",
		URL:   "example.com",
	}
}

func createCommandAll() *CreateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}
	globals.Client, _ = mock.APIClient(mock.API{
		ListVersionsFn: testutil.ListVersionsOk,
		GetVersionFn:   testutil.GetActiveVersionOK,
		CloneVersionFn: testutil.CloneVersionOK,
	})("token", "endpoint")

	return &CreateCommand{
		Base: common.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "logs",
		serviceVersion: common.OptionalServiceVersion{
			OptionalString: common.OptionalString{Value: "1"},
		},
		autoClone: common.OptionalAutoClone{
			OptionalBool: common.OptionalBool{Value: true},
		},
		Index:             "logs",
		URL:               "example.com",
		Pipeline:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "my_pipeline_id"},
		RequestMaxEntries: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
		User:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "user"},
		Password:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "password"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "example.com"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandMissingServiceID() *CreateCommand {
	res := createCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &UpdateCommand{
		Base: common.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		serviceVersion: common.OptionalServiceVersion{
			OptionalString: common.OptionalString{Value: "1"},
		},
		autoClone: common.OptionalAutoClone{
			OptionalBool: common.OptionalBool{Value: true},
		},
	}
}

func updateCommandAll() *UpdateCommand {
	var b bytes.Buffer

	globals := config.Data{
		File:   config.File{},
		Env:    config.Environment{},
		Output: &b,
	}

	return &UpdateCommand{
		Base: common.Base{
			Globals: &globals,
		},
		manifest: manifest.Data{
			Flag: manifest.Flag{
				ServiceID: "123",
			},
		},
		EndpointName: "log",
		serviceVersion: common.OptionalServiceVersion{
			OptionalString: common.OptionalString{Value: "1"},
		},
		autoClone: common.OptionalAutoClone{
			OptionalBool: common.OptionalBool{Value: true},
		},
		NewName:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new1"},
		Index:             common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		URL:               common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		Pipeline:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		RequestMaxEntries: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		User:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Password:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new12"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new13"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getElasticsearchOK(i *fastly.GetElasticsearchInput) (*fastly.Elasticsearch, error) {
	return &fastly.Elasticsearch{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		Index:             "logs",
		URL:               "example.com",
		Pipeline:          "my_pipeline_id",
		User:              "user",
		Password:          "password",
		RequestMaxEntries: 2,
		RequestMaxBytes:   2,
		Placement:         "none",
		TLSCACert:         "-----BEGIN CERTIFICATE-----foo",
		TLSHostname:       "example.com",
		TLSClientCert:     "-----BEGIN CERTIFICATE-----bar",
		TLSClientKey:      "-----BEGIN PRIVATE KEY-----bar",
		FormatVersion:     2,
	}, nil
}
