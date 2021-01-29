package digitalocean_test

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

func TestDigitalOceanCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "digitalocean", "create", "--service-id", "123", "--version", "1", "--name", "log", "--access-key", "foo", "--secret-key", "abc"},
			wantError: "error parsing arguments: required flag --bucket not provided",
		},
		{
			args:      []string{"logging", "digitalocean", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--secret-key", "abc"},
			wantError: "error parsing arguments: required flag --access-key not provided",
		},
		{
			args:      []string{"logging", "digitalocean", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo"},
			wantError: "error parsing arguments: required flag --secret-key not provided",
		},
		{
			args:       []string{"logging", "digitalocean", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--secret-key", "abc"},
			api:        mock.API{CreateDigitalOceanFn: createDigitalOceanOK},
			wantOutput: "Created DigitalOcean Spaces logging endpoint log (service 123 version 1)",
		},
		{
			args:      []string{"logging", "digitalocean", "create", "--service-id", "123", "--version", "1", "--name", "log", "--bucket", "log", "--access-key", "foo", "--secret-key", "abc"},
			api:       mock.API{CreateDigitalOceanFn: createDigitalOceanError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
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

func TestDigitalOceanList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"logging", "digitalocean", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDigitalOceansFn: listDigitalOceansOK},
			wantOutput: listDigitalOceansShortOutput,
		},
		{
			args:       []string{"logging", "digitalocean", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api:        mock.API{ListDigitalOceansFn: listDigitalOceansOK},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args:       []string{"logging", "digitalocean", "list", "--service-id", "123", "--version", "1", "-v"},
			api:        mock.API{ListDigitalOceansFn: listDigitalOceansOK},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args:       []string{"logging", "digitalocean", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDigitalOceansFn: listDigitalOceansOK},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args:       []string{"logging", "-v", "digitalocean", "list", "--service-id", "123", "--version", "1"},
			api:        mock.API{ListDigitalOceansFn: listDigitalOceansOK},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args:      []string{"logging", "digitalocean", "list", "--service-id", "123", "--version", "1"},
			api:       mock.API{ListDigitalOceansFn: listDigitalOceansError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
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

func TestDigitalOceanDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "digitalocean", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "digitalocean", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{GetDigitalOceanFn: getDigitalOceanError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "digitalocean", "describe", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{GetDigitalOceanFn: getDigitalOceanOK},
			wantOutput: describeDigitalOceanOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
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

func TestDigitalOceanUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "digitalocean", "update", "--service-id", "123", "--version", "1", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "digitalocean", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetDigitalOceanFn:    getDigitalOceanError,
				UpdateDigitalOceanFn: updateDigitalOceanOK,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "digitalocean", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetDigitalOceanFn:    getDigitalOceanOK,
				UpdateDigitalOceanFn: updateDigitalOceanError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "digitalocean", "update", "--service-id", "123", "--version", "1", "--name", "logs", "--new-name", "log"},
			api: mock.API{
				GetDigitalOceanFn:    getDigitalOceanOK,
				UpdateDigitalOceanFn: updateDigitalOceanOK,
			},
			wantOutput: "Updated DigitalOcean Spaces logging endpoint log (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
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

func TestDigitalOceanDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "digitalocean", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:      []string{"logging", "digitalocean", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:       mock.API{DeleteDigitalOceanFn: deleteDigitalOceanError},
			wantError: errTest.Error(),
		},
		{
			args:       []string{"logging", "digitalocean", "delete", "--service-id", "123", "--version", "1", "--name", "logs"},
			api:        mock.API{DeleteDigitalOceanFn: deleteDigitalOceanOK},
			wantOutput: "Deleted DigitalOcean Spaces logging endpoint logs (service 123 version 1)",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
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

func createDigitalOceanOK(i *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	s := fastly.DigitalOcean{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if i.Name != "" {
		s.Name = i.Name
	}

	return &s, nil
}

func createDigitalOceanError(i *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

func listDigitalOceansOK(i *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
	return []*fastly.DigitalOcean{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			BucketName:        "my-logs",
			Domain:            "https://digitalocean.us-east-1.amazonaws.com",
			AccessKey:         "1234",
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
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
			SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
			Domain:            "https://digitalocean.us-east-2.amazonaws.com",
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

func listDigitalOceansError(i *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
	return nil, errTest
}

var listDigitalOceansShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listDigitalOceansVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
	DigitalOcean 1/2
		Service ID: 123
		Version: 1
		Name: logs
		Bucket: my-logs
		Domain: https://digitalocean.us-east-1.amazonaws.com
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
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
	DigitalOcean 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		Bucket: analytics
		Domain: https://digitalocean.us-east-2.amazonaws.com
		Access key: 1234
		Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
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

func getDigitalOceanOK(i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return &fastly.DigitalOcean{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		BucketName:        "my-logs",
		Domain:            "https://digitalocean.us-east-1.amazonaws.com",
		AccessKey:         "1234",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
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

func getDigitalOceanError(i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

var describeDigitalOceanOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
Name: logs
Bucket: my-logs
Domain: https://digitalocean.us-east-1.amazonaws.com
Access key: 1234
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
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

func updateDigitalOceanOK(i *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return &fastly.DigitalOcean{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		BucketName:        "my-logs",
		Domain:            "https://digitalocean.us-east-1.amazonaws.com",
		AccessKey:         "1234",
		SecretKey:         "-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA",
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

func updateDigitalOceanError(i *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

func deleteDigitalOceanOK(i *fastly.DeleteDigitalOceanInput) error {
	return nil
}

func deleteDigitalOceanError(i *fastly.DeleteDigitalOceanInput) error {
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
