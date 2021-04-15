package kinesis

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

func TestCreateKinesisInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateKinesisInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateKinesisInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				StreamName:     "stream",
				AccessKey:      "access",
				SecretKey:      "secret",
			},
		},
		{
			name: "required values set flag serviceID using IAM role",
			cmd:  createCommandRequiredIAMRole(),
			want: &fastly.CreateKinesisInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				StreamName:     "stream",
				IAMRole:        "arn:aws:iam::123456789012:role/KinesisAccess",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateKinesisInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "logs",
				StreamName:        "stream",
				Region:            "us-east-1",
				AccessKey:         "access",
				SecretKey:         "secret",
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

func TestUpdateKinesisInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateKinesisInput
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetKinesisFn: getKinesisOK},
			want: &fastly.UpdateKinesisInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetKinesisFn: getKinesisOK},
			want: &fastly.UpdateKinesisInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("new1"),
				StreamName:        fastly.String("new2"),
				AccessKey:         fastly.String("new3"),
				SecretKey:         fastly.String("new4"),
				IAMRole:           fastly.String(""),
				Region:            fastly.String("new5"),
				Format:            fastly.String("new7"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new9"),
				Placement:         fastly.String("new11"),
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
		StreamName:   "stream",
		AccessKey:    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "access"},
		SecretKey:    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "secret"},
	}
}

func createCommandRequiredIAMRole() *CreateCommand {
	return &CreateCommand{
		manifest:     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName: "log",
		Version:      2,
		StreamName:   "stream",
		IAMRole:      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "arn:aws:iam::123456789012:role/KinesisAccess"},
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "logs",
		Version:           2,
		StreamName:        "stream",
		AccessKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "access"},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "secret"},
		Region:            "us-east-1",
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
		StreamName:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		AccessKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		IAMRole:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: ""},
		Region:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getKinesisOK(i *fastly.GetKinesisInput) (*fastly.Kinesis, error) {
	return &fastly.Kinesis{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		StreamName:        "stream",
		Region:            "us-east-1",
		AccessKey:         "access",
		SecretKey:         "secret",
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		Placement:         "none",
	}, nil
}
