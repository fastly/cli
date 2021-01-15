package digitalocean

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

func TestCreateDigitalOceanInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateDigitalOceanInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateDigitalOceanInput{
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
			want: &fastly.CreateDigitalOceanInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				BucketName:        "bucket",
				Domain:            "nyc3.digitaloceanspaces.com",
				AccessKey:         "access",
				SecretKey:         "secret",
				Path:              "/log",
				Period:            3600,
				GzipLevel:         0,
				Format:            `%h %l %u %t "%r" %>s %b`,
				MessageType:       "classic",
				FormatVersion:     2,
				ResponseCondition: "Prevent default logging",
				TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
				Placement:         "none",
				PublicKey:         pgpPublicKey(),
				CompressionCodec:  "zstd",
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

func TestUpdateDigitalOceanInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateDigitalOceanInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetDigitalOceanFn: getDigitalOceanOK},
			want: &fastly.UpdateDigitalOceanInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("new1"),
				BucketName:        fastly.String("new2"),
				Domain:            fastly.String("new3"),
				AccessKey:         fastly.String("new4"),
				SecretKey:         fastly.String("new5"),
				Path:              fastly.String("new6"),
				Period:            fastly.Uint(3601),
				GzipLevel:         fastly.Uint(0),
				Format:            fastly.String("new7"),
				FormatVersion:     fastly.Uint(3),
				ResponseCondition: fastly.String("new8"),
				MessageType:       fastly.String("new9"),
				TimestampFormat:   fastly.String("new10"),
				Placement:         fastly.String("new11"),
				PublicKey:         fastly.String("new12"),
				CompressionCodec:  fastly.String("new13"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetDigitalOceanFn: getDigitalOceanOK},
			want: &fastly.UpdateDigitalOceanInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
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
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		BucketName:        "bucket",
		AccessKey:         "access",
		SecretKey:         "secret",
		Domain:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "nyc3.digitaloceanspaces.com"},
		Path:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "/log"},
		Period:            common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3600},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "none"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "classic"},
		PublicKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: pgpPublicKey()},
		CompressionCodec:  common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "zstd"},
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
		BucketName:        common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		Domain:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		AccessKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		Path:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		Period:            common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3601},
		GzipLevel:         common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 0},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
		PublicKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new12"},
		CompressionCodec:  common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new13"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getDigitalOceanOK(i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return &fastly.DigitalOcean{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		BucketName:        "bucket",
		Domain:            "nyc3.digitaloceanspaces.com",
		AccessKey:         "access",
		SecretKey:         "secret",
		Path:              "/log",
		Period:            3600,
		GzipLevel:         0,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
		PublicKey:         pgpPublicKey(),
		CompressionCodec:  "zstd",
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
