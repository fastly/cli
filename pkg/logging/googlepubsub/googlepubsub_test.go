package googlepubsub

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

func TestCreateGooglePubSubInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreatePubsubInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreatePubsubInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				User:           "user@example.com",
				SecretKey:      "secret",
				ProjectID:      "project",
				Topic:          "topic",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreatePubsubInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				Topic:             "topic",
				User:              "user@example.com",
				SecretKey:         "secret",
				ProjectID:         "project",
				FormatVersion:     2,
				Format:            `%h %l %u %t "%r" %>s %b`,
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

func TestUpdateGooglePubSubInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdatePubsubInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetPubsubFn: getGooglePubSubOK},
			want: &fastly.UpdatePubsubInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("new1"),
				User:              fastly.String("new2"),
				SecretKey:         fastly.String("new3"),
				ProjectID:         fastly.String("new4"),
				Topic:             fastly.String("new5"),
				Placement:         fastly.String("new6"),
				Format:            fastly.String("new7"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new8"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetPubsubFn: getGooglePubSubOK},
			want: &fastly.UpdatePubsubInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("log"),
				User:              fastly.String("user@example.com"),
				SecretKey:         fastly.String("secret"),
				ProjectID:         fastly.String("project"),
				Topic:             fastly.String("topic"),
				Placement:         fastly.String("none"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:     fastly.Uint(2),
				ResponseCondition: fastly.String("Prevent default logging"),
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
		User:         "user@example.com",
		SecretKey:    "secret",
		ProjectID:    "project",
		Topic:        "topic",
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "logs",
		Version:           2,
		User:              "user@example.com",
		ProjectID:         "project",
		Topic:             "topic",
		SecretKey:         "secret",
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
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
		User:              common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		SecretKey:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		ProjectID:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		Topic:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getGooglePubSubOK(i *fastly.GetPubsubInput) (*fastly.Pubsub, error) {
	return &fastly.Pubsub{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		ResponseCondition: "Prevent default logging",
		Format:            `%h %l %u %t "%r" %>s %b`,
		User:              "user@example.com",
		SecretKey:         "secret",
		ProjectID:         "project",
		Topic:             "topic",
		Placement:         "none",
		FormatVersion:     2,
	}, nil
}
