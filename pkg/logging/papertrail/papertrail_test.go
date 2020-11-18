package papertrail

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

func TestCreatePapertrailInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreatePapertrailInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreatePapertrailInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Address:        "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreatePapertrailInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				Address:           "example.com",
				Port:              22,
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
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

func TestUpdatePapertrailInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdatePapertrailInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetPapertrailFn: getPapertrailOK},
			want: &fastly.UpdatePapertrailInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("logs"),
				Address:           fastly.String("example.com"),
				Port:              fastly.Uint(22),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.Uint(2),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetPapertrailFn: getPapertrailOK},
			want: &fastly.UpdatePapertrailInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Uint(23),
				Format:            fastly.String("new3"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new4"),
				Placement:         fastly.String("new5"),
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
		Address:      "example.com",
		Version:      2,
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		Address:           "example.com",
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 22},
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
		NewName:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new1"},
		Address:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 23},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getPapertrailOK(i *fastly.GetPapertrailInput) (*fastly.Papertrail, error) {
	return &fastly.Papertrail{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com",
		Port:              22,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}
