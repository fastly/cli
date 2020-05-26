package https

import (
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/fastly"
)

func TestCreateHTTPSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateHTTPSInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateHTTPSInput{
				Service: "123",
				Version: 2,
				Name:    "log",
				URL:     "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateHTTPSInput{
				Service:           "123",
				Version:           2,
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
			have, err := testcase.cmd.createInput()
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertEqual(t, testcase.want, have)
		})
	}
}

func TestUpdateHTTPSInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateHTTPSInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetHTTPSFn: getHTTPSOK},
			want: &fastly.UpdateHTTPSInput{
				Service:           "123",
				Version:           2,
				Name:              "log",
				NewName:           "new1",
				ResponseCondition: "new2",
				Format:            "new3",
				URL:               "new4",
				RequestMaxEntries: 3,
				RequestMaxBytes:   3,
				ContentType:       "new5",
				HeaderName:        "new6",
				HeaderValue:       "new7",
				Method:            "new8",
				JSONFormat:        "new9",
				Placement:         "new10",
				TLSCACert:         "new11",
				TLSClientCert:     "new12",
				TLSClientKey:      "new13",
				TLSHostname:       "new14",
				MessageType:       "new15",
				FormatVersion:     3,
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetHTTPSFn: getHTTPSOK},
			want: &fastly.UpdateHTTPSInput{
				Service:           "123",
				Version:           2,
				Name:              "log",
				NewName:           "log",
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
	return &CreateCommand{
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		Version:      2,
		URL:          "example.com",
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "logs",
		Version:           2,
		URL:               "example.com",
		ContentType:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "application/json"},
		HeaderName:        common.OptionalString{Optional: common.Optional{Valid: true}, Value: "name"},
		HeaderValue:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "value"},
		Method:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "GET"},
		JSONFormat:        common.OptionalString{Optional: common.Optional{Valid: true}, Value: "1"},
		MessageType:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "classic"},
		RequestMaxEntries: common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
		TLSCACert:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN CERTIFICATE-----foo"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "example.com"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN CERTIFICATE-----bar"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "-----BEGIN PRIVATE KEY-----bar"},
	}
}

func createCommandMissingServiceID() *CreateCommand {
	res := createCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *UpdateCommand {
	return &UpdateCommand{
		Base:         common.Base{Globals: &config.Data{Client: nil}},
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		Version:      2,
	}
}

func updateCommandAll() *UpdateCommand {
	return &UpdateCommand{
		Base:              common.Base{Globals: &config.Data{Client: nil}},
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "logs",
		Version:           2,
		NewName:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new1"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		URL:               common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		ContentType:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		HeaderName:        common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		HeaderValue:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		Method:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
		JSONFormat:        common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new9"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new10"},
		RequestMaxEntries: common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		TLSCACert:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new11"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new12"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new13"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new14"},
		MessageType:       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new15"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getHTTPSOK(i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         i.Service,
		Version:           i.Version,
		Name:              "log",
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
	}, nil
}
