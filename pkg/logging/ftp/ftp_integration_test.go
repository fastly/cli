package ftp_test

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

func TestFTPCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "ftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--user", "anonymous", "--password", "foo@example.com"},
			wantError: "error parsing arguments: required flag --address not provided",
		},
		{
			args:      []string{"logging", "ftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--password", "foo@example.com"},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args:      []string{"logging", "ftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--user", "anonymous"},
			wantError: "error parsing arguments: required flag --password not provided",
		},
		{
			args:       []string{"logging", "ftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--user", "anonymous", "--password", "foo@example.com"},
			api:        mock.API{CreateFTPFn: createFTPOK},
			wantOutput: "Created FTP logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "ftp", "create", "--service-id", "123", "--version", "1", "--name", "log", "--address", "example.com", "--user", "anonymous", "--password", "foo@example.com"},
			api:       mock.API{CreateFTPFn: createFTPError},
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

func TestFTPList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "ftp", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListFTPsFn: listFTPsOK},
			wantOutput: listFTPsShortOutput,
		},
		{
			args:       []string{"logging", "ftp", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListFTPsFn: listFTPsOK},
			wantOutput: listFTPsVerboseOutput,
		},
		{
			args:       []string{"logging", "ftp", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListFTPsFn: listFTPsOK},
			wantOutput: listFTPsVerboseOutput,
		},
		{
			args:       []string{"logging", "ftp", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListFTPsFn: listFTPsOK},
			wantOutput: listFTPsVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "ftp", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListFTPsFn: listFTPsOK},
			wantOutput: listFTPsVerboseOutput,
		},
		{
			args:      []string{"logging", "ftp", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListFTPsFn: listFTPsError},
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

func TestFTPDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "ftp", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "ftp", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetFTPFn: getFTPError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "ftp", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetFTPFn: getFTPOK},
			wantOutput: describeFTPOutput,
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

func TestFTPUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "ftp", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "ftp", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetFTPFn:    getFTPError,
				UpdateFTPFn: updateFTPOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "ftp", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetFTPFn:    getFTPOK,
				UpdateFTPFn: updateFTPError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "ftp", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetFTPFn:    getFTPOK,
				UpdateFTPFn: updateFTPOK,
			},
			wantOutput: "Updated FTP logging endpoint log (service 123 version 1)",
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

func TestFTPDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "ftp", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "ftp", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteFTPFn: deleteFTPError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "ftp", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteFTPFn: deleteFTPOK},
			wantOutput: "Deleted FTP logging endpoint logs (service 123 version 1)",
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

func createFTPOK(i *fastly.CreateFTPInput) (*fastly.FTP, error) {
	return &fastly.FTP{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
	}, nil
}

func createFTPError(i *fastly.CreateFTPInput) (*fastly.FTP, error) {
	return nil, errTest
}

func listFTPsOK(i *fastly.ListFTPsInput) ([]*fastly.FTP, error) {
	return []*fastly.FTP{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Address:           "example.com",
			Port:              123,
			Username:          "anonymous",
			Password:          "foo@example.com",
			PublicKey:         pgpPublicKey(),
			Path:              "logs/",
			Period:            3600,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			Address:           "127.0.0.1",
			Port:              456,
			Username:          "foo",
			Password:          "password",
			PublicKey:         pgpPublicKey(),
			Path:              "logs/",
			Period:            86400,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
		},
	}, nil
}

func listFTPsError(i *fastly.ListFTPsInput) ([]*fastly.FTP, error) {
	return nil, errTest
}

var listFTPsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listFTPsVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	FTP 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Address: example.com
		Port: 123
		Username: anonymous
		Password: foo@example.com
		Public key: `+pgpPublicKey()+`
		Path: logs/
		Period: 3600
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
	FTP 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Address: 127.0.0.1
		Port: 456
		Username: foo
		Password: password
		Public key: `+pgpPublicKey()+`
		Path: logs/
		Period: 86400
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
`) + "\n\n"

func getFTPOK(i *fastly.GetFTPInput) (*fastly.FTP, error) {
	return &fastly.FTP{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Address:           "example.com",
		Port:              123,
		Username:          "anonymous",
		Password:          "foo@example.com",
		PublicKey:         pgpPublicKey(),
		Path:              "logs/",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

func getFTPError(i *fastly.GetFTPInput) (*fastly.FTP, error) {
	return nil, errTest
}

var describeFTPOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Address: example.com
Port: 123
Username: anonymous
Password: foo@example.com
Public key: `+pgpPublicKey()+`
Path: logs/
Period: 3600
GZip level: 9
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
`) + "\n"

func updateFTPOK(i *fastly.UpdateFTPInput) (*fastly.FTP, error) {
	return &fastly.FTP{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Address:           "example.com",
		Port:              123,
		Username:          "anonymous",
		Password:          "foo@example.com",
		PublicKey:         pgpPublicKey(),
		Path:              "logs/",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
	}, nil
}

func updateFTPError(i *fastly.UpdateFTPInput) (*fastly.FTP, error) {
	return nil, errTest
}

func deleteFTPOK(i *fastly.DeleteFTPInput) error {
	return nil
}

func deleteFTPError(i *fastly.DeleteFTPInput) error {
	return errTest
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
