package sftp_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestSFTPCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "user", "--ssh-known-hosts", knownHosts(), "--port", "80"},
			wantError: "error parsing arguments: required flag --address not provided",
		},
		{
			args:      []string{"logging", "sftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--ssh-known-hosts", knownHosts(), "--port", "80"},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args:      []string{"logging", "sftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--user", "user", "--port", "80"},
			wantError: "error parsing arguments: required flag --ssh-known-hosts not provided",
		},
		{
			args:       []string{"logging", "sftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--user", "user", "--ssh-known-hosts", knownHosts(), "--port", "80"},
			api:        mock.API{CreateSFTPFn: createSFTPOK},
			wantOutput: "Created SFTP logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "sftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--user", "user", "--ssh-known-hosts", knownHosts(), "--port", "80"},
			api:       mock.API{CreateSFTPFn: createSFTPError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestSFTPList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "sftp", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSFTPsFn: listSFTPsOK},
			wantOutput: listSFTPsShortOutput,
		},
		{
			args:       []string{"logging", "sftp", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListSFTPsFn: listSFTPsOK},
			wantOutput: listSFTPsVerboseOutput,
		},
		{
			args:       []string{"logging", "sftp", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListSFTPsFn: listSFTPsOK},
			wantOutput: listSFTPsVerboseOutput,
		},
		{
			args:       []string{"logging", "sftp", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSFTPsFn: listSFTPsOK},
			wantOutput: listSFTPsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "sftp", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListSFTPsFn: listSFTPsOK},
			wantOutput: listSFTPsVerboseOutput,
		},
		{
			args:      []string{"logging", "sftp", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListSFTPsFn: listSFTPsError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestSFTPDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sftp", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "sftp", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetSFTPFn: getSFTPError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "sftp", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetSFTPFn: getSFTPOK},
			wantOutput: describeSFTPOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestSFTPUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sftp", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "sftp", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSFTPFn:    getSFTPError,
				UpdateSFTPFn: updateSFTPOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sftp", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSFTPFn:    getSFTPOK,
				UpdateSFTPFn: updateSFTPError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "sftp", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetSFTPFn:    getSFTPOK,
				UpdateSFTPFn: updateSFTPOK,
			},
			wantOutput: "Updated SFTP logging endpoint log (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestSFTPDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "sftp", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "sftp", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteSFTPFn: deleteSFTPError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "sftp", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteSFTPFn: deleteSFTPOK},
			wantOutput: "Deleted SFTP logging endpoint logs (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createSFTPOK(i *fastly.CreateSFTPInput) (*fastly.SFTP, error) {
	s := fastly.SFTP{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createSFTPError(i *fastly.CreateSFTPInput) (*fastly.SFTP, error) {
	return nil, errTest
}

func listSFTPsOK(i *fastly.ListSFTPsInput) ([]*fastly.SFTP, error) {
	return []*fastly.SFTP{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Address:           "127.0.0.1",
			Port:              514,
			User:              "user",
			Password:          "password",
			PublicKey:         pgpPublicKey(),
			SecretKey:         sshPrivateKey(),
			SSHKnownHosts:     knownHosts(),
			Path:              "/logs",
			Period:            3600,
			GzipLevel:         2,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Address:           "example.com",
			Port:              123,
			User:              "user",
			Password:          "password",
			PublicKey:         pgpPublicKey(),
			SecretKey:         sshPrivateKey(),
			SSHKnownHosts:     knownHosts(),
			Path:              "/analytics",
			Period:            3600,
			GzipLevel:         3,
			Format:            `%h %l %u %t "%r" %>s %b`,
			MessageType:       "classic",
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
		},
	}, nil
}

func listSFTPsError(i *fastly.ListSFTPsInput) ([]*fastly.SFTP, error) {
	return nil, errTest
}

var listSFTPsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listSFTPsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	SFTP 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Address: 127.0.0.1
		Port: 514
		User: user
		Password: password
		Public key: `+pgpPublicKey()+`
		Secret key: `+sshPrivateKey()+`
		SSH known hosts: `+knownHosts()+`
		Path: /logs
		Period: 3600
		GZip level: 2
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Message type: classic
		Response condition: Prevent default logging
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
	SFTP 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: example.com
		Port: 123
		User: user
		Password: password
		Public key: `+pgpPublicKey()+`
		Secret key: `+sshPrivateKey()+`
		SSH known hosts: `+knownHosts()+`
		Path: /analytics
		Period: 3600
		GZip level: 3
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Message type: classic
		Response condition: Prevent default logging
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
`) + "\n\n"

func getSFTPOK(i *fastly.GetSFTPInput) (*fastly.SFTP, error) {
	return &fastly.SFTP{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com",
		Port:              514,
		User:              "user",
		Password:          "password",
		PublicKey:         pgpPublicKey(),
		SecretKey:         sshPrivateKey(),
		SSHKnownHosts:     knownHosts(),
		Path:              "/logs",
		Period:            3600,
		GzipLevel:         2,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

func getSFTPError(i *fastly.GetSFTPInput) (*fastly.SFTP, error) {
	return nil, errTest
}

var describeSFTPOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Address: example.com
Port: 514
User: user
Password: password
Public key: `+pgpPublicKey()+`
Secret key: `+sshPrivateKey()+`
SSH known hosts: `+knownHosts()+`
Path: /logs
Period: 3600
GZip level: 2
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Message type: classic
Response condition: Prevent default logging
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
`) + "\n"

func updateSFTPOK(i *fastly.UpdateSFTPInput) (*fastly.SFTP, error) {
	return &fastly.SFTP{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Address:           "example.com",
		Port:              514,
		User:              "user",
		Password:          "password",
		PublicKey:         pgpPublicKey(),
		SecretKey:         sshPrivateKey(),
		SSHKnownHosts:     knownHosts(),
		Path:              "/logs",
		Period:            3600,
		GzipLevel:         3,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

func updateSFTPError(i *fastly.UpdateSFTPInput) (*fastly.SFTP, error) {
	return nil, errTest
}

func deleteSFTPOK(i *fastly.DeleteSFTPInput) error {
	return nil
}

func deleteSFTPError(i *fastly.DeleteSFTPInput) error {
	return errTest
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
