package s3

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

func TestCreateS3Input(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateS3Input
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateS3Input{
				Service:    "123",
				Version:    2,
				Name:       "log",
				BucketName: "bucket",
				AccessKey:  "access",
				SecretKey:  "secret",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateS3Input{
				Service:                      "123",
				Version:                      2,
				Name:                         "logs",
				BucketName:                   "bucket",
				Domain:                       "domain",
				AccessKey:                    "access",
				SecretKey:                    "secret",
				Path:                         "path",
				Period:                       3600,
				GzipLevel:                    2,
				Format:                       `%h %l %u %t "%r" %>s %b`,
				MessageType:                  "classic",
				FormatVersion:                2,
				ResponseCondition:            "Prevent default logging",
				TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
				Redundancy:                   fastly.S3RedundancyStandard,
				Placement:                    "none",
				ServerSideEncryptionKMSKeyID: "kmskey",
				ServerSideEncryption:         fastly.S3ServerSideEncryptionAES,
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

func TestUpdateS3Input(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateS3Input
		wantError string
	}{
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetS3Fn: getS3OK},
			want: &fastly.UpdateS3Input{
				Service:                      "123",
				Version:                      2,
				Name:                         "logs",
				NewName:                      "logs",
				BucketName:                   "bucket",
				AccessKey:                    "access",
				SecretKey:                    "secret",
				Domain:                       "domain",
				Path:                         "path",
				Period:                       3600,
				GzipLevel:                    2,
				Format:                       `%h %l %u %t "%r" %>s %b`,
				FormatVersion:                2,
				MessageType:                  "classic",
				ResponseCondition:            "Prevent default logging",
				TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
				Placement:                    "none",
				Redundancy:                   fastly.S3RedundancyStandard,
				ServerSideEncryption:         fastly.S3ServerSideEncryptionAES,
				ServerSideEncryptionKMSKeyID: "kmskey",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetS3Fn: getS3OK},
			want: &fastly.UpdateS3Input{
				Service:                      "123",
				Version:                      2,
				Name:                         "logs",
				NewName:                      "new1",
				BucketName:                   "new2",
				AccessKey:                    "new3",
				SecretKey:                    "new4",
				Domain:                       "new5",
				Path:                         "new6",
				Period:                       3601,
				GzipLevel:                    3,
				Format:                       "new7",
				FormatVersion:                3,
				MessageType:                  "new8",
				ResponseCondition:            "new9",
				TimestampFormat:              "new10",
				Placement:                    "new11",
				Redundancy:                   fastly.S3RedundancyReduced,
				ServerSideEncryption:         fastly.S3ServerSideEncryptionKMS,
				ServerSideEncryptionKMSKeyID: "new12",
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
		BucketName:   "bucket",
		AccessKey:    "access",
		SecretKey:    "secret",
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:                     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:                 "logs",
		Version:                      2,
		BucketName:                   "bucket",
		AccessKey:                    "access",
		SecretKey:                    "secret",
		Domain:                       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "domain"},
		Path:                         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "path"},
		Period:                       common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3600},
		GzipLevel:                    common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		Format:                       common.OptionalString{Optional: common.Optional{Valid: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:                common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 2},
		MessageType:                  common.OptionalString{Optional: common.Optional{Valid: true}, Value: "classic"},
		ResponseCondition:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "Prevent default logging"},
		TimestampFormat:              common.OptionalString{Optional: common.Optional{Valid: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:                    common.OptionalString{Optional: common.Optional{Valid: true}, Value: "none"},
		Redundancy:                   common.OptionalString{Optional: common.Optional{Valid: true}, Value: string(fastly.S3RedundancyStandard)},
		ServerSideEncryption:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: string(fastly.S3ServerSideEncryptionAES)},
		ServerSideEncryptionKMSKeyID: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "kmskey"},
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
		EndpointName: "logs",
		Version:      2,
	}
}

func updateCommandAll() *UpdateCommand {
	return &UpdateCommand{
		Base:                         common.Base{Globals: &config.Data{Client: nil}},
		manifest:                     manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:                 "logs",
		Version:                      2,
		NewName:                      common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new1"},
		BucketName:                   common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new2"},
		AccessKey:                    common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new3"},
		SecretKey:                    common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new4"},
		Domain:                       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new5"},
		Path:                         common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new6"},
		Period:                       common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3601},
		GzipLevel:                    common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		Format:                       common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new7"},
		FormatVersion:                common.OptionalUint{Optional: common.Optional{Valid: true}, Value: 3},
		MessageType:                  common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new8"},
		ResponseCondition:            common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new9"},
		TimestampFormat:              common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new10"},
		Placement:                    common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new11"},
		Redundancy:                   common.OptionalString{Optional: common.Optional{Valid: true}, Value: string(fastly.S3RedundancyReduced)},
		ServerSideEncryption:         common.OptionalString{Optional: common.Optional{Valid: true}, Value: string(fastly.S3ServerSideEncryptionKMS)},
		ServerSideEncryptionKMSKeyID: common.OptionalString{Optional: common.Optional{Valid: true}, Value: "new12"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getS3OK(i *fastly.GetS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:                    i.Service,
		Version:                      i.Version,
		Name:                         "logs",
		BucketName:                   "bucket",
		Domain:                       "domain",
		AccessKey:                    "access",
		SecretKey:                    "secret",
		Path:                         "path",
		Period:                       3600,
		GzipLevel:                    2,
		Format:                       `%h %l %u %t "%r" %>s %b`,
		FormatVersion:                2,
		ResponseCondition:            "Prevent default logging",
		MessageType:                  "classic",
		TimestampFormat:              "%Y-%m-%dT%H:%M:%S.000",
		Placement:                    "none",
		Redundancy:                   fastly.S3RedundancyStandard,
		ServerSideEncryptionKMSKeyID: "kmskey",
		ServerSideEncryption:         fastly.S3ServerSideEncryptionAES,
	}, nil
}
