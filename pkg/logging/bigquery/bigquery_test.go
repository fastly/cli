package bigquery

import (
	"testing"
	"time"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/fastly"
)

func TestCreateBigQueryInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateBigQueryInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateBigQueryInput{
				Service:   "123",
				Version:   2,
				Name:      "log",
				ProjectID: "123",
				Dataset:   "dataset",
				Table:     "table",
				User:      "user",
				SecretKey: "-----BEGIN PRIVATE KEY-----foo",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateBigQueryInput{
				Service:           "123",
				Version:           2,
				Name:              "log",
				ProjectID:         "123",
				Dataset:           "dataset",
				Table:             "table",
				Template:          "template",
				User:              "user",
				SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
				Format:            `%h %l %u %t "%r" %>s %b`,
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
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

func TestUpdateBigQueryInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateBigQueryInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetBigQueryFn: getBigQueryOK},
			want: &fastly.UpdateBigQueryInput{
				Service:           "123",
				Version:           2,
				Name:              "logs",
				NewName:           "logs",
				ProjectID:         "123",
				Dataset:           "dataset",
				Table:             "table",
				Template:          "template",
				User:              "user",
				SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
				Format:            `%h %l %u %t "%r" %>s %b`,
				ResponseCondition: "Prevent default logging",
				Placement:         "none",
				FormatVersion:     2,
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetBigQueryFn: getBigQueryOK},
			want: &fastly.UpdateBigQueryInput{
				Service:           "123",
				Version:           2,
				Name:              "logs",
				NewName:           "new1",
				ProjectID:         "new2",
				Dataset:           "new3",
				Table:             "new4",
				User:              "new5",
				SecretKey:         "new6",
				Template:          "new7",
				ResponseCondition: "new8",
				Placement:         "new9",
				Format:            "new10",
				FormatVersion:     3,
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
		ProjectID:    "123",
		Dataset:      "dataset",
		Table:        "table",
		User:         "user",
		SecretKey:    "-----BEGIN PRIVATE KEY-----foo",
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		ProjectID:         "123",
		Dataset:           "dataset",
		Table:             "table",
		User:              "user",
		SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
		Template:          common.OptionalString{Optional: common.Optional{Valid: true}, Value: "template"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
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
		ProjectID:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		Dataset:           common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		Table:             common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		User:              common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		SecretKey:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		Template:          common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
		Placement:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new9"},
		Format:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new10"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getBigQueryOK(i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:         i.Service,
		Version:           i.Version,
		Name:              "logs",
		Format:            `%h %l %u %t "%r" %>s %b`,
		User:              "user",
		ProjectID:         "123",
		Dataset:           "dataset",
		Table:             "table",
		Template:          "template",
		SecretKey:         "-----BEGIN PRIVATE KEY-----foo",
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
		FormatVersion:     2,
		CreatedAt:         &time.Time{},
		UpdatedAt:         &time.Time{},
		DeletedAt:         &time.Time{},
	}, nil
}
