package honeycomb

import (
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v2/fastly"
)

func TestCreateHoneycombInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateHoneycombInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateHoneycombInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Token:          "tkn",
				Dataset:        "logs",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateHoneycombInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				Token:             "tkn",
				Dataset:           "logs",
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

func TestUpdateHoneycombInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateHoneycombInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetHoneycombFn: getHoneycombOK},
			want: &fastly.UpdateHoneycombInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("logs"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.Uint(2),
				Token:             fastly.String("tkn"),
				Dataset:           fastly.String("logs"),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetHoneycombFn: getHoneycombOK},
			want: &fastly.UpdateHoneycombInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("new1"),
				Format:            fastly.String("new2"),
				FormatVersion:     fastly.Uint(3),
				Token:             fastly.String("new3"),
				Dataset:           fastly.String("new4"),
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
		Dataset:      "logs",
		Version:      2,
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Token:             "tkn",
		Dataset:           "logs",
		Version:           2,
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
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
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		Token:             common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		Dataset:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getHoneycombOK(i *fastly.GetHoneycombInput) (*fastly.Honeycomb, error) {
	return &fastly.Honeycomb{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Token:             "tkn",
		Dataset:           "logs",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}
