package logshuttle

import (
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestCreateLogshuttleInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateLogshuttleInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateLogshuttleInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Token:          "tkn",
				URL:            "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateLogshuttleInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				URL:               "example.com",
				Token:             "tkn",
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
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

func TestUpdateLogshuttleInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateLogshuttleInput
		wantError string
	}{
		{
			name: "no update",
			cmd:  updateCommandNoUpdate(),
			api:  mock.API{GetLogshuttleFn: getLogshuttleOK},
			want: &fastly.UpdateLogshuttleInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetLogshuttleFn: getLogshuttleOK},
			want: &fastly.UpdateLogshuttleInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Format:            fastly.String("new2"),
				FormatVersion:     fastly.Uint(3),
				Token:             fastly.String("new3"),
				URL:               fastly.String("new4"),
				ResponseCondition: fastly.String("new5"),
				Placement:         fastly.String("new6"),
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
		Token:        "tkn",
		URL:          "example.com",
		Version:      2,
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Token:             "tkn",
		URL:               "example.com",
		Version:           2,
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 2},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "none"},
	}
}

func createCommandMissingServiceID() *CreateCommand {
	res := createCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdate() *UpdateCommand {
	return &UpdateCommand{
		Base:         cmd.Base{Globals: &config.Data{Client: nil}},
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		Version:      2,
	}
}

func updateCommandAll() *UpdateCommand {
	return &UpdateCommand{
		Base:              cmd.Base{Globals: &config.Data{Client: nil}},
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		NewName:           cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new1"},
		Format:            cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new2"},
		FormatVersion:     cmd.OptionalUint{Optional: cmd.Optional{WasSet: true}, Value: 3},
		Token:             cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new3"},
		URL:               cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new4"},
		ResponseCondition: cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new5"},
		Placement:         cmd.OptionalString{Optional: cmd.Optional{WasSet: true}, Value: "new6"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getLogshuttleOK(i *fastly.GetLogshuttleInput) (*fastly.Logshuttle, error) {
	return &fastly.Logshuttle{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		URL:               "example.com",
		Token:             "tkn",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}
