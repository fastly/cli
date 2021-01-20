package s3

import (
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
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
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				BucketName:     "bucket",
				AccessKey:      "access",
				SecretKey:      "secret",
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateS3Input{
				ServiceID:                    "123",
				ServiceVersion:               2,
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
				PublicKey:                    pgpPublicKey(),
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
				ServiceID:                    "123",
				ServiceVersion:               2,
				Name:                         "logs",
				NewName:                      fastly.String("logs"),
				BucketName:                   fastly.String("bucket"),
				AccessKey:                    fastly.String("access"),
				SecretKey:                    fastly.String("secret"),
				Domain:                       fastly.String("domain"),
				Path:                         fastly.String("path"),
				Period:                       fastly.Uint(3600),
				GzipLevel:                    fastly.Uint(2),
				Format:                       fastly.String(`%h %l %u %t "%r" %>s %b`),
				FormatVersion:                fastly.Uint(2),
				MessageType:                  fastly.String("classic"),
				ResponseCondition:            fastly.String("Prevent default logging"),
				TimestampFormat:              fastly.String("%Y-%m-%dT%H:%M:%S.000"),
				Placement:                    fastly.String("none"),
				PublicKey:                    fastly.String(pgpPublicKey()),
				Redundancy:                   fastly.S3RedundancyStandard,
				ServerSideEncryption:         fastly.S3ServerSideEncryptionAES,
				ServerSideEncryptionKMSKeyID: fastly.String("kmskey"),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetS3Fn: getS3OK},
			want: &fastly.UpdateS3Input{
				ServiceID:                    "123",
				ServiceVersion:               2,
				Name:                         "logs",
				NewName:                      fastly.String("new1"),
				BucketName:                   fastly.String("new2"),
				AccessKey:                    fastly.String("new3"),
				SecretKey:                    fastly.String("new4"),
				Domain:                       fastly.String("new5"),
				Path:                         fastly.String("new6"),
				Period:                       fastly.Uint(3601),
				GzipLevel:                    fastly.Uint(3),
				Format:                       fastly.String("new7"),
				FormatVersion:                fastly.Uint(3),
				MessageType:                  fastly.String("new8"),
				ResponseCondition:            fastly.String("new9"),
				TimestampFormat:              fastly.String("new10"),
				Placement:                    fastly.String("new11"),
				Redundancy:                   fastly.S3RedundancyReduced,
				ServerSideEncryption:         fastly.S3ServerSideEncryptionKMS,
				ServerSideEncryptionKMSKeyID: fastly.String("new12"),
				PublicKey:                    fastly.String("new13"),
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
		Domain:                       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "domain"},
		Path:                         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "path"},
		Period:                       common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3600},
		GzipLevel:                    common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		Format:                       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:                common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		MessageType:                  common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:                    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
		PublicKey:                    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: pgpPublicKey()},
		Redundancy:                   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: string(fastly.S3RedundancyStandard)},
		ServerSideEncryption:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: string(fastly.S3ServerSideEncryptionAES)},
		ServerSideEncryptionKMSKeyID: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "kmskey"},
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
		NewName:                      common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new1"},
		BucketName:                   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		AccessKey:                    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		SecretKey:                    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		Domain:                       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Path:                         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		Period:                       common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3601},
		GzipLevel:                    common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		Format:                       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:                common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		MessageType:                  common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		ResponseCondition:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		TimestampFormat:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		Placement:                    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
		Redundancy:                   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: string(fastly.S3RedundancyReduced)},
		ServerSideEncryption:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: string(fastly.S3ServerSideEncryptionKMS)},
		ServerSideEncryptionKMSKeyID: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new12"},
		PublicKey:                    common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new13"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getS3OK(i *fastly.GetS3Input) (*fastly.S3, error) {
	return &fastly.S3{
		ServiceID:                    i.ServiceID,
		ServiceVersion:               i.ServiceVersion,
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
		PublicKey:                    pgpPublicKey(),
		Redundancy:                   fastly.S3RedundancyStandard,
		ServerSideEncryptionKMSKeyID: "kmskey",
		ServerSideEncryption:         fastly.S3ServerSideEncryptionAES,
	}, nil
}

// pgpPublicKey returns a PEM encoded PGP public key suitable for testing.
func pgpPublicKey() string {
	return strings.TrimSpace(`-----BEGIN PGP PUBLIC KEY BLOCK-----
mQENBFyUD8sBCACyFnB39AuuTygseek+eA4fo0cgwva6/FSjnWq7riouQee8GgQ/
ibXTRyv4iVlwI12GswvMTIy7zNvs1R54i0qvsLr+IZ4GVGJqs6ZJnvQcqe3xPoR4
8AnBfw90o32r/LuHf6QCJXi+AEu35koNlNAvLJ2B+KACaNB7N0EeWmqpV/1V2k9p
lDYk+th7LcCuaFNGqKS/PrMnnMqR6VDLCjHhNx4KR79b0Twm/2qp6an3hyNRu8Gn
dwxpf1/BUu3JWf+LqkN4Y3mbOmSUL3MaJNvyQguUzTfS0P0uGuBDHrJCVkMZCzDB
89ag55jCPHyGeHBTd02gHMWzsg3WMBWvCsrzABEBAAG0JXRlcnJhZm9ybSAodGVz
dCkgPHRlc3RAdGVycmFmb3JtLmNvbT6JAU4EEwEIADgWIQSHYyc6Kj9l6HzQsau6
vFFc9jxV/wUCXJQPywIbAwULCQgHAgYVCgkICwIEFgIDAQIeAQIXgAAKCRC6vFFc
9jxV/815CAClb32OxV7wG01yF97TzlyTl8TnvjMtoG29Mw4nSyg+mjM3b8N7iXm9
OLX59fbDAWtBSldSZE22RXd3CvlFOG/EnKBXSjBtEqfyxYSnyOPkMPBYWGL/ApkX
SvPYJ4LKdvipYToKFh3y9kk2gk1DcDBDyaaHvR+3rv1u3aoy7/s2EltAfDS3ZQIq
7/cWTLJml/lleeB/Y6rPj8xqeCYhE5ahw9gsV/Mdqatl24V9Tks30iijx0Hhw+Gx
kATUikMGr2GDVqoIRga5kXI7CzYff4rkc0Twn47fMHHHe/KY9M2yVnMHUXmAZwbG
M1cMI/NH1DjevCKdGBLcRJlhuLPKF/anuQENBFyUD8sBCADIpd7r7GuPd6n/Ikxe
u6h7umV6IIPoAm88xCYpTbSZiaK30Svh6Ywra9jfE2KlU9o6Y/art8ip0VJ3m07L
4RSfSpnzqgSwdjSq5hNour2Fo/BzYhK7yaz2AzVSbe33R0+RYhb4b/6N+bKbjwGF
ftCsqVFMH+PyvYkLbvxyQrHlA9woAZaNThI1ztO5rGSnGUR8xt84eup28WIFKg0K
UEGUcTzz+8QGAwAra+0ewPXo/AkO+8BvZjDidP417u6gpBHOJ9qYIcO9FxHeqFyu
YrjlrxowEgXn5wO8xuNz6Vu1vhHGDHGDsRbZF8pv1d5O+0F1G7ttZ2GRRgVBZPwi
kiyRABEBAAGJATYEGAEIACAWIQSHYyc6Kj9l6HzQsau6vFFc9jxV/wUCXJQPywIb
DAAKCRC6vFFc9jxV/9YOCACe8qmOSnKQpQfW+PqYOqo3dt7JyweTs3FkD6NT8Zml
dYy/vkstbTjPpX6aTvUZjkb46BVi7AOneVHpD5GBqvRsZ9iVgDYHaehmLCdKiG5L
3Tp90NN+QY5WDbsGmsyk6+6ZMYejb4qYfweQeduOj27aavCJdLkCYMoRKfcFYI8c
FaNmEfKKy/r1PO20NXEG6t9t05K/frHy6ZG8bCNYdpagfFVot47r9JaQqWlTNtIR
5+zkkSq/eG9BEtRij3a6cTdQbktdBzx2KBeI0PYc1vlZR0LpuFKZqY9vlE6vTGLR
wMfrTEOvx0NxUM3rpaCgEmuWbB1G1Hu371oyr4srrr+N
=28dr
-----END PGP PUBLIC KEY BLOCK-----
`)
}
