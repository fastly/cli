package https

import (
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
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
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				URL:            "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateHTTPSInput{
				ServiceID:         "123",
				ServiceVersion:    2,
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
				ServiceID:         "123",
				ServiceVersion:    2,
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
			api:  mock.API{GetHTTPSFn: getHTTPSOK},
			want: &fastly.UpdateHTTPSInput{
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
		ContentType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "application/json"},
		HeaderName:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "name"},
		HeaderValue:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "value"},
		Method:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "GET"},
		JSONFormat:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "1"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "classic"},
		RequestMaxEntries: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
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
		EndpointName:      "log",
		Version:           2,
		NewName:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new1"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		URL:               common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		ContentType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		HeaderName:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		HeaderValue:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		Method:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		JSONFormat:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		RequestMaxEntries: common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		RequestMaxBytes:   common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		TLSCACert:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
		TLSClientCert:     common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new12"},
		TLSClientKey:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new13"},
		TLSHostname:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new14"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new15"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getHTTPSOK(i *fastly.GetHTTPSInput) (*fastly.HTTPS, error) {
	return &fastly.HTTPS{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
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
