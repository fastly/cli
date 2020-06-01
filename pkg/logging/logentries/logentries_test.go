package logentries

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

func TestCreateLogentriesInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateLogentriesInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateLogentriesInput{
				Service: "123",
				Version: 2,
				Name:    "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandOK(),
			want: &fastly.CreateLogentriesInput{
				Service:           "123",
				Version:           2,
				Name:              "log",
				Port:              22,
				UseTLS:            fastly.CBool(true),
				Token:             "tkn",
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

func TestUpdateLogentriesInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateLogentriesInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetLogentriesFn: getLogentriesOK},
			want: &fastly.UpdateLogentriesInput{
				Service:           "123",
				Version:           2,
				Name:              "logs",
				NewName:           "logs",
				Port:              22,
				UseTLS:            fastly.CBool(true),
				Token:             "tkn",
				Format:            `%h %l %u %t "%r" %>s %b`,
				FormatVersion:     2,
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetLogentriesFn: getLogentriesOK},
			want: &fastly.UpdateLogentriesInput{
				Service:           "123",
				Version:           2,
				Name:              "logs",
				NewName:           "new1",
				Port:              23,
				UseTLS:            fastly.CBool(true),
				Token:             "new2",
				Format:            "new3",
				FormatVersion:     3,
				ResponseCondition: "new4",
				Placement:         "new5",
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
		Version:           2,
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 22},
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: true},
		Token:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "tkn"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
	}
}

func createCommandRequired() *CreateCommand {
	return &CreateCommand{
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
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
		Base:              common.Base{Globals: &config.Data{Client: nil}},
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "logs",
		Version:           2,
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 22},
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: true},
		Token:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "tkn"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
	}
}

func updateCommandAll() *UpdateCommand {
	return &UpdateCommand{
		Base:              common.Base{Globals: &config.Data{Client: nil}},
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		Port:              common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 23},
		UseTLS:            common.OptionalBool{Optional: common.Optional{Valid: true}, Value: true},
		NewName:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new1"},
		Token:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
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

func getLogentriesOK(i *fastly.GetLogentriesInput) (*fastly.Logentries, error) {
	return &fastly.Logentries{
		ServiceID:         i.Service,
		Version:           i.Version,
		Name:              "logs",
		Token:             "tkn",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}
