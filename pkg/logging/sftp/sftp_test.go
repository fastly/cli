package sftp

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

func TestCreateSFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *CreateCommand
		want      *fastly.CreateSFTPInput
		wantError string
	}{
		{
			name: "required values set flag serviceID",
			cmd:  createCommandRequired(),
			want: &fastly.CreateSFTPInput{
				ServiceID:      "123",
				ServiceVersion: 2,
				Name:           "log",
				Address:        "127.0.0.1",
				User:           "user",
				SSHKnownHosts:  knownHosts(),
			},
		},
		{
			name: "all values set flag serviceID",
			cmd:  createCommandAll(),
			want: &fastly.CreateSFTPInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				Address:           "127.0.0.1",
				Port:              80,
				User:              "user",
				Password:          "password",
				PublicKey:         pgpPublicKey(),
				SecretKey:         sshPrivateKey(),
				SSHKnownHosts:     knownHosts(),
				Path:              "/log",
				Period:            3600,
				FormatVersion:     2,
				GzipLevel:         2,
				Format:            `%h %l %u %t "%r" %>s %b`,
				ResponseCondition: "Prevent default logging",
				MessageType:       "classic",
				TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
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

func TestUpdateSFTPInput(t *testing.T) {
	for _, testcase := range []struct {
		name      string
		cmd       *UpdateCommand
		api       mock.API
		want      *fastly.UpdateSFTPInput
		wantError string
	}{
		{
			name: "all values set flag serviceID",
			cmd:  updateCommandAll(),
			api:  mock.API{GetSFTPFn: getSFTPOK},
			want: &fastly.UpdateSFTPInput{
				ServiceID:         "123",
				ServiceVersion:    2,
				Name:              "log",
				NewName:           fastly.String("new1"),
				Address:           fastly.String("new2"),
				Port:              fastly.Uint(81),
				User:              fastly.String("new3"),
				SSHKnownHosts:     fastly.String("new4"),
				Password:          fastly.String("new5"),
				PublicKey:         fastly.String("new6"),
				SecretKey:         fastly.String("new7"),
				Path:              fastly.String("new8"),
				Period:            fastly.Uint(3601),
				FormatVersion:     fastly.Uint(3),
				GzipLevel:         fastly.Uint(3),
				Format:            fastly.String("new9"),
				ResponseCondition: fastly.String("new10"),
				TimestampFormat:   fastly.String("new11"),
				Placement:         fastly.String("new12"),
				MessageType:       fastly.String("new13"),
			},
		},
		{
			name: "no updates",
			cmd:  updateCommandNoUpdates(),
			api:  mock.API{GetSFTPFn: getSFTPOK},
			want: &fastly.UpdateSFTPInput{
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
		manifest:      manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:  "log",
		Version:       2,
		Address:       "127.0.0.1",
		User:          "user",
		SSHKnownHosts: knownHosts(),
	}
}

func createCommandAll() *CreateCommand {
	return &CreateCommand{
		manifest:          manifest.Data{Flag: manifest.Flag{ServiceID: "123"}},
		EndpointName:      "log",
		Version:           2,
		Address:           "127.0.0.1",
		User:              "user",
		SSHKnownHosts:     knownHosts(),
		Port:              common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 80},
		Password:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "password"},
		PublicKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: pgpPublicKey()},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: sshPrivateKey()},
		Path:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "/log"},
		Period:            common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3600},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: `%h %l %u %t "%r" %>s %b`},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		GzipLevel:         common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 2},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "classic"},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "Prevent default logging"},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "%Y-%m-%dT%H:%M:%S.000"},
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
		Address:           common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new2"},
		User:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new3"},
		SSHKnownHosts:     common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new4"},
		Port:              common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 81},
		Password:          common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new5"},
		PublicKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new6"},
		SecretKey:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new7"},
		Path:              common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new8"},
		Period:            common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3601},
		Format:            common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new9"},
		FormatVersion:     common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		GzipLevel:         common.OptionalUint{Optional: common.Optional{WasSet: true}, Value: 3},
		ResponseCondition: common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new10"},
		TimestampFormat:   common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new11"},
		Placement:         common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new12"},
		MessageType:       common.OptionalString{Optional: common.Optional{WasSet: true}, Value: "new13"},
	}
}

func updateCommandMissingServiceID() *UpdateCommand {
	res := updateCommandAll()
	res.manifest = manifest.Data{}
	return res
}

func getSFTPOK(i *fastly.GetSFTPInput) (*fastly.SFTP, error) {
	return &fastly.SFTP{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "127.0.0.1",
		Port:              80,
		User:              "user",
		Password:          "password",
		PublicKey:         pgpPublicKey(),
		SecretKey:         sshPrivateKey(),
		SSHKnownHosts:     knownHosts(),
		Path:              "/log",
		Period:            3600,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		GzipLevel:         2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

// knownHosts returns sample known hosts suitable for testing
func knownHosts() string {
	return strings.TrimSpace(`
example.com
127.0.0.1
`)
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

// sshPrivateKey returns a private key suitable for testing.
func sshPrivateKey() string {
	return strings.TrimSpace(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDDo+/YbQ1cZVoRhZ/bbQtPxpycDS5Lty+M8e5swCKpmo0/Eym2
KrVpEVMoU8eGtwVRvGDR2LtmFKvd86QUWkn2V3lYgY66SNj9n4R/YSDT4/GRkg+4
Egi++ihpZA+SAIODF4+l1bh/FFu0XUpQLXvJ4Tm0++7bm3tEq+XQr9znrwIDAQAB
AoGAfDa374e9te47s2hNyLmBNxN5F7Nes4AJVsm8gZuz5k9UYrm+AAU5zQ3M6IvY
4PWPEQgzyMh8oyF4xaENikaRMhSMfinUmTd979cHbOM6cEKPk28oQcIybsdSzX7G
ZWRh65Ze1DUmBe6R2BUh3Zn4lq9PsqB0TeZeV7Xo/VaIpFECQQDoznQi8HOY8MNM
7ZDdRhFAkS2X5OGqXOjYdLABGNvJhajgoRsTbgDyJG83qn6yYq7wEHYlMddGZ3ln
RLnpsThjAkEA1yGXae8WURFEqjp5dMLBxU07apKvEF4zK1OxZ0VjIOJdIpoRBBuL
IthGBuMrfbF1W5tlmQlj5ik0KhVpBZoHRQJAZP7DdTDZBT1VjHb3RHcUHu2cWOvL
VkvuG5ErlZ5CIv+gDqr1gw1SzbkuoniNdDfJao3Jo0Mm//z9tuYivRXLvwJBALG3
Wzi0vI/Nnxas5YayGJaf3XSFpj70QnsJUWUJagFRXjTmZyYohsELPpYT9eqIvXUm
o0BQBImvAhu9whtRia0CQCFdDHdNnyyzKH8vC0NsEN65h3Bp2KEPkv8SOV27ZRR2
xIGqLusk3y+yzbueLZJ117osdB1Owr19fvAHR7vq6Mw=
-----END RSA PRIVATE KEY-----`)
}
