package azureblob_test

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

func TestBlobStorageCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "azureblob", "create", "--service-id", "123", "--version", "1", "--name", "log", "--account-name", "account", "--sas-token", "abc"},
			wantError: "error parsing arguments: required flag --container not provided",
		},
		{
			args:      []string{"logging", "azureblob", "create", "--service-id", "123", "--version", "1", "--name", "log", "--container", "log", "--sas-token", "abc"},
			wantError: "error parsing arguments: required flag --account-name not provided",
		},
		{
			args: []string{"logging", "azureblob", "create", "--service-id", "123", "--version", "2", "--name", "log", "--account-name", "account", "--container", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				CreateBlobStorageFn: createBlobStorageError,
			},
			wantError: "error parsing arguments: required flag --sas-token not provided",
		},
		{
			args: []string{"logging", "azureblob", "create", "--service-id", "123", "--version", "2", "--name", "log", "--account-name", "account", "--container", "log", "--sas-token", "abc", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				CreateBlobStorageFn: createBlobStorageOK,
			},
			wantOutput: "Created Azure Blob Storage logging endpoint log (service 123 version 3)",
		},
		{
			args: []string{"logging", "azureblob", "create", "--service-id", "123", "--version", "2", "--name", "log", "--account-name", "account", "--container", "log", "--sas-token", "abc", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				CreateBlobStorageFn: createBlobStorageError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "azureblob", "create", "--service-id", "123", "--version", "1", "--name", "log", "--account-name", "account", "--container", "log", "--sas-token", "abc", "--compression-codec", "zstd", "--gzip-level", "9"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				CreateBlobStorageFn: createBlobStorageError,
			},
			wantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestBlobStorageList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"logging", "azureblob", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			wantOutput: listBlobStoragesShortOutput,
		},
		{
			args: []string{"logging", "azureblob", "list", "--service-id", "123", "--version", "2", "--verbose"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			wantOutput: listBlobStoragesVerboseOutput,
		},
		{
			args: []string{"logging", "azureblob", "list", "--service-id", "123", "--version", "2", "-v"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			wantOutput: listBlobStoragesVerboseOutput,
		},
		{
			args: []string{"logging", "azureblob", "--verbose", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			wantOutput: listBlobStoragesVerboseOutput,
		},
		{
			args: []string{"logging", "-v", "azureblob", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			wantOutput: listBlobStoragesVerboseOutput,
		},
		{
			args: []string{"logging", "azureblob", "list", "--service-id", "123", "--version", "2"},
			api: mock.API{
				ListVersionsFn:     listVersionsOK,
				GetVersionFn:       getVersionOK,
				ListBlobStoragesFn: listBlobStoragesError,
			},
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestBlobStorageDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "azureblob", "describe", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "azureblob", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				GetBlobStorageFn: getBlobStorageError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "azureblob", "describe", "--service-id", "123", "--version", "2", "--name", "logs"},
			api: mock.API{
				ListVersionsFn:   listVersionsOK,
				GetVersionFn:     getVersionOK,
				GetBlobStorageFn: getBlobStorageOK,
			},
			wantOutput: describeBlobStorageOutput,
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestBlobStorageUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "azureblob", "update", "--service-id", "123", "--version", "2", "--new-name", "log"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "azureblob", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				UpdateBlobStorageFn: updateBlobStorageError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "azureblob", "update", "--service-id", "123", "--version", "2", "--name", "logs", "--new-name", "log", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				UpdateBlobStorageFn: updateBlobStorageOK,
			},
			wantOutput: "Updated Azure Blob Storage logging endpoint log (service 123 version 3)",
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func TestBlobStorageDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"logging", "azureblob", "delete", "--service-id", "123", "--version", "2"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"logging", "azureblob", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				DeleteBlobStorageFn: deleteBlobStorageError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"logging", "azureblob", "delete", "--service-id", "123", "--version", "2", "--name", "logs", "--autoclone"},
			api: mock.API{
				ListVersionsFn:      listVersionsOK,
				GetVersionFn:        getVersionOK,
				CloneVersionFn:      cloneVersionOK,
				DeleteBlobStorageFn: deleteBlobStorageOK,
			},
			wantOutput: "Deleted Azure Blob Storage logging endpoint logs (service 123 version 3)",
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
				cliVersioner  update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, cliVersioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createBlobStorageOK(i *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error) {
	s := fastly.BlobStorage{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Path:              "/logs",
		AccountName:       "account",
		Container:         "container",
		SASToken:          "token",
		Period:            3600,
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		PublicKey:         pgpPublicKey(),
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		MessageType:       "classic",
		Placement:         "none",
		ResponseCondition: "Prevent default logging",
		CompressionCodec:  "zstd",
	}

	return &s, nil
}

func createBlobStorageError(i *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error) {
	return nil, errTest
}

func listBlobStoragesOK(i *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error) {
	return []*fastly.BlobStorage{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			Path:              "/logs",
			AccountName:       "account",
			Container:         "container",
			SASToken:          "token",
			Period:            3600,
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			PublicKey:         pgpPublicKey(),
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			Placement:         "none",
			ResponseCondition: "Prevent default logging",
			CompressionCodec:  "zstd",
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			AccountName:       "account",
			Container:         "analytics",
			SASToken:          "token",
			Path:              "/logs",
			Period:            86400,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			MessageType:       "classic",
			ResponseCondition: "Prevent default logging",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			Placement:         "none",
			PublicKey:         pgpPublicKey(),
			CompressionCodec:  "zstd",
		},
	}, nil
}

func listBlobStoragesError(i *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error) {
	return nil, errTest
}

var listBlobStoragesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      2        logs
123      2        analytics
`) + "\n"

var listBlobStoragesVerboseOutput = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 2
	BlobStorage 1/2
		Service ID: 123
		Version: 2
		Name: logs
		Container: container
		Account name: account
		SAS token: token
		Path: /logs
		Period: 3600
		GZip level: 0
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Public key: `+pgpPublicKey()+`
		File max bytes: 0
		Compression codec: zstd
	BlobStorage 2/2
		Service ID: 123
		Version: 2
		Name: analytics
		Container: analytics
		Account name: account
		SAS token: token
		Path: /logs
		Period: 86400
		GZip level: 0
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Placement: none
		Public key: `+pgpPublicKey()+`
		File max bytes: 0
		Compression codec: zstd
`) + "\n\n"

func getBlobStorageOK(i *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error) {
	return &fastly.BlobStorage{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		Container:         "container",
		AccountName:       "account",
		SASToken:          "token",
		Path:              "/logs",
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

func getBlobStorageError(i *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error) {
	return nil, errTest
}

var describeBlobStorageOutput = strings.TrimSpace(`
Service ID: 123
Version: 2
Name: logs
Container: container
Account name: account
SAS token: token
Path: /logs
Period: 3600
GZip level: 0
Format: %h %l %u %t "%r" %>s %b
Format version: 2
Response condition: Prevent default logging
Message type: classic
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Placement: none
Public key: `+pgpPublicKey()+`
File max bytes: 0
Compression codec: zstd
`) + "\n"

func updateBlobStorageOK(i *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error) {
	return &fastly.BlobStorage{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		Container:         "container",
		AccountName:       "account",
		SASToken:          "token",
		Path:              "/logs",
		Period:            3600,
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

func updateBlobStorageError(i *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error) {
	return nil, errTest
}

func deleteBlobStorageOK(i *fastly.DeleteBlobStorageInput) error {
	return nil
}

func deleteBlobStorageError(i *fastly.DeleteBlobStorageInput) error {
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

func listVersionsOK(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

func getVersionOK(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		ServiceID: i.ServiceID,
		Number:    2,
		Active:    true,
		UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}, nil
}

func cloneVersionOK(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion + 1}, nil
}
