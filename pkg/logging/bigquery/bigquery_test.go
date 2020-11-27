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
	"github.com/fastly/go-fastly/v2/fastly"
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
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				ProjectID:      "123",
				Dataset:        "dataset",
				Table:          "table",
				User:           "user",
				SecretKey:      "-----BEGIN PRIVATE KEY-----foo",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateBigQueryInput{
				ServiceID:         "123",
				ServiceVersion:    2,
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
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("logs"),
				ProjectID:         fastly.String("123"),
				Dataset:           fastly.String("dataset"),
				Table:             fastly.String("table"),
				Template:          fastly.String("template"),
				User:              fastly.String("user"),
				SecretKey:         fastly.String("-----BEGIN PRIVATE KEY-----foo"),
				Format:            fastly.String(`%h %l %u %t "%r" %>s %b`),
				ResponseCondition: fastly.String("Prevent default logging"),
				Placement:         fastly.String("none"),
				FormatVersion:     fastly.Uint(2),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetBigQueryFn: getBigQueryOK},
			want: &fastly.UpdateBigQueryInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				NewName:           fastly.String("new1"),
				ProjectID:         fastly.String("new2"),
				Dataset:           fastly.String("new3"),
				Table:             fastly.String("new4"),
				User:              fastly.String("new5"),
				SecretKey:         fastly.String("new6"),
				Template:          fastly.String("new7"),
				ResponseCondition: fastly.String("new8"),
				Placement:         fastly.String("new9"),
				Format:            fastly.String("new10"),
				FormatVersion:     fastly.Uint(3),
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
		Template:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "template"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
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
		ProjectID:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		Dataset:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		Table:             common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		User:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		Template:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getBigQueryOK(i *fastly.GetBigQueryInput) (*fastly.BigQuery, error) {
	return &fastly.BigQuery{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
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
