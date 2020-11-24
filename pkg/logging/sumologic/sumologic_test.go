package sumologic

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

func TestCreateSumologicInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateSumologicInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateSumologicInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				URL:            "example.com",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandOK(),
			want: &fastly.CreateSumologicInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				URL:               "example.com",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
				MessageType:       "classic",
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

func TestUpdateSFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateSumologicInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetSumologicFn: getSumologicOK},
			want: &fastly.UpdateSumologicInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("logs"),
				URL:               fastly.String("example.com"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.Int(2),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
				MessageType:       fastly.String("classic"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetSumologicFn: getSumologicOK},
			want: &fastly.UpdateSumologicInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("new1"),
				URL:               fastly.String("new2"),
				Format:            fastly.String("new3"),
				FormatVersion:     fastly.Int(3),
				ResponseCondition: fastly.String("new4"),
				Placement:         fastly.String("new5"),
				MessageType:       fastly.String("new6"),
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

func createCommandOK() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		URL:               "example.com",
		Version:           2,
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalInt{Optional: common.Optional{WasSet: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "classic"},
	}
}

func createCommandRequired() *CreateCommand {
	return &CreateCommand{
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		URL:          "example.com",
		Version:      2,
	}
}

func createCommandMissingServiceID() *CreateCommand {
	res := createCommandOK()
	res.manifest = manifest.Data{}
	return res
}

func updateCommandNoUpdates() *UpdateCommand {
	return &UpdateCommand{
		Base:         common.Base{Globals: &config.Data{Client: nil}},
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "logs",
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
		URL:               common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		FormatVersion:     common.OptionalInt{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getSumologicOK(i *fastly.GetSumologicInput) (*fastly.Sumologic, error) {
	return &fastly.Sumologic{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		URL:               "example.com",
		Format:            `%h %l %u %t "%r" %>s %b`,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		FormatVersion:     2,
		Placement:         "none",
	}, nil
}
