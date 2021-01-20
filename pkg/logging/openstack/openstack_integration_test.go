package openstack_test

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

func TestOpenstackCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "openstack", "create", "--service-id", "123", "--version", "1", "--name", "log", "--access-key", "foo", "--user", "user", "--url", "https://example.com"},
			wantError: "error parsing arguments: required flag --bucket not provided",
		},
		{
			args:      []string{"logging", "openstack", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--user", "user", "--url", "https://example.com"},
			wantError: "error parsing arguments: required flag --access-key not provided",
		},
		{
			args:      []string{"logging", "openstack", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--url", "https://example.com"},
			wantError: "error parsing arguments: required flag --user not provided",
		},
		{
			args:      []string{"logging", "openstack", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--user", "user"},
			wantError: "error parsing arguments: required flag --url not provided",
		},
		{
			args:       []string{"logging", "openstack", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--user", "user", "--url", "https://example.com"},
			api:        mock.API{CreateOpenstackFn: createOpenstackOK},
			wantOutput: "Created OpenStack logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "openstack", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--user", "user", "--url", "https://example.com"},
			api:       mock.API{CreateOpenstackFn: createOpenstackError},
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

func TestOpenstackList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "openstack", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListOpenstacksFn: listOpenstacksOK},
			wantOutput: listOpenstacksShortOutput,
		},
		{
			args:       []string{"logging", "openstack", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListOpenstacksFn: listOpenstacksOK},
			wantOutput: listOpenstacksVerboseOutput,
		},
		{
			args:       []string{"logging", "openstack", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListOpenstacksFn: listOpenstacksOK},
			wantOutput: listOpenstacksVerboseOutput,
		},
		{
			args:       []string{"logging", "openstack", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListOpenstacksFn: listOpenstacksOK},
			wantOutput: listOpenstacksVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "openstack", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListOpenstacksFn: listOpenstacksOK},
			wantOutput: listOpenstacksVerboseOutput,
		},
		{
			args:      []string{"logging", "openstack", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListOpenstacksFn: listOpenstacksError},
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

func TestOpenstackDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "openstack", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "openstack", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetOpenstackFn: getOpenstackError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "openstack", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetOpenstackFn: getOpenstackOK},
			wantOutput: describeOpenstackOutput,
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

func TestOpenstackUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "openstack", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "openstack", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetOpenstackFn:    getOpenstackError,
				UpdateOpenstackFn: updateOpenstackOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "openstack", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetOpenstackFn:    getOpenstackOK,
				UpdateOpenstackFn: updateOpenstackError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "openstack", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetOpenstackFn:    getOpenstackOK,
				UpdateOpenstackFn: updateOpenstackOK,
			},
			wantOutput: "Updated OpenStack logging endpoint log (service 123 version 1)",
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

func TestOpenstackDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "openstack", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "openstack", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteOpenstackFn: deleteOpenstackError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "openstack", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteOpenstackFn: deleteOpenstackOK},
			wantOutput: "Deleted OpenStack logging endpoint logs (service 123 version 1)",
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

func createOpenstackOK(i *fastly.CreateOpenstackInput) (*fastly.Openstack, error) {
	s := fastly.Openstack{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createOpenstackError(i *fastly.CreateOpenstackInput) (*fastly.Openstack, error) {
	return nil, errTest
}

func listOpenstacksOK(i *fastly.ListOpenstackInput) ([]*fastly.Openstack, error) {
	return []*fastly.Openstack{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			BucketName:        "my-logs",
			AccessKey:         "1234",
			User:              "user",
			URL:               "https://example.com",
			Path:              "logs/",
			Period:            3600,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			MessageType:       "classic",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
			PublicKey:         pgpPublicKey(),
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			BucketName:        "analytics",
			AccessKey:         "1234",
			User:              "user2",
			URL:               "https://two.example.com",
			Path:              "logs/",
			Period:            86400,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
			PublicKey:         pgpPublicKey(),
		},
	}, nil
}

func listOpenstacksError(i *fastly.ListOpenstackInput) ([]*fastly.Openstack, error) {
	return nil, errTest
}

var listOpenstacksShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listOpenstacksVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	Openstack 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Bucket: my-logs
		Access key: 1234
		User: user
		URL: https://example.com
		Path: logs/
		Period: 3600
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Public key: `+pgpPublicKey()+`
	Openstack 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Bucket: analytics
		Access key: 1234
		User: user2
		URL: https://two.example.com
		Path: logs/
		Period: 86400
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Public key: `+pgpPublicKey()+`
`) + "\n\n"

func getOpenstackOK(i *fastly.GetOpenstackInput) (*fastly.Openstack, error) {
	return &fastly.Openstack{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		BucketName:        "my-logs",
		AccessKey:         "1234",
		User:              "user",
		URL:               "https://example.com",
		Path:              "logs/",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
		PublicKey:         pgpPublicKey(),
	}, nil
}

func getOpenstackError(i *fastly.GetOpenstackInput) (*fastly.Openstack, error) {
	return nil, errTest
}

var describeOpenstackOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Bucket: my-logs
Access key: 1234
User: user
URL: https://example.com
Path: logs/
Period: 3600
GZip level: 9
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
Public key: `+pgpPublicKey()+`
`) + "\n"

func updateOpenstackOK(i *fastly.UpdateOpenstackInput) (*fastly.Openstack, error) {
	return &fastly.Openstack{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		BucketName:        "my-logs",
		AccessKey:         "1234",
		User:              "userupdate",
		URL:               "https://update.example.com",
		Path:              "logs/",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		Placement:         "none",
		PublicKey:         pgpPublicKey(),
	}, nil
}

func updateOpenstackError(i *fastly.UpdateOpenstackInput) (*fastly.Openstack, error) {
	return nil, errTest
}

func deleteOpenstackOK(i *fastly.DeleteOpenstackInput) error {
	return nil
}

func deleteOpenstackError(i *fastly.DeleteOpenstackInput) error {
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
