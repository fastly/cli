package azureblob_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"

	root "github.com/fastly/cli/pkg/commands/service"
	parent "github.com/fastly/cli/pkg/commands/service/logging"
	sub "github.com/fastly/cli/pkg/commands/service/logging/azureblob"
)

func TestBlobStorageCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --account-name account --container log --sas-token abc --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateBlobStorageFn: createBlobStorageOK,
			},
			WantOutput: "Created Azure Blob Storage logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --account-name account --container log --sas-token abc --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateBlobStorageFn: createBlobStorageError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name log --account-name account --container log --sas-token abc --compression-codec zstd --gzip-level 9 --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				CreateBlobStorageFn: createBlobStorageError,
			},
			WantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestBlobStorageList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			WantOutput: listBlobStoragesShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			WantOutput: listBlobStoragesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListBlobStoragesFn: listBlobStoragesOK,
			},
			WantOutput: listBlobStoragesVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				ListBlobStoragesFn: listBlobStoragesError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestBlobStorageDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				GetBlobStorageFn: getBlobStorageError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				GetBlobStorageFn: getBlobStorageOK,
			},
			WantOutput: describeBlobStorageOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestBlobStorageUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateBlobStorageFn: updateBlobStorageError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				UpdateBlobStorageFn: updateBlobStorageOK,
			},
			WantOutput: "Updated Azure Blob Storage logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestBlobStorageDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				DeleteBlobStorageFn: deleteBlobStorageError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				CloneVersionFn:      testutil.CloneVersionResult(4),
				DeleteBlobStorageFn: deleteBlobStorageOK,
			},
			WantOutput: "Deleted Azure Blob Storage logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createBlobStorageOK(_ context.Context, i *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error) {
	s := fastly.BlobStorage{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Path:              fastly.ToPointer("/logs"),
		AccountName:       fastly.ToPointer("account"),
		Container:         fastly.ToPointer("container"),
		SASToken:          fastly.ToPointer("token"),
		Period:            fastly.ToPointer(3600),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		PublicKey:         fastly.ToPointer(pgpPublicKey()),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		MessageType:       fastly.ToPointer("classic"),
		Placement:         fastly.ToPointer("none"),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		CompressionCodec:  fastly.ToPointer("zstd"),
	}

	return &s, nil
}

func createBlobStorageError(_ context.Context, _ *fastly.CreateBlobStorageInput) (*fastly.BlobStorage, error) {
	return nil, errTest
}

func listBlobStoragesOK(_ context.Context, i *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error) {
	return []*fastly.BlobStorage{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			Path:              fastly.ToPointer("/logs"),
			AccountName:       fastly.ToPointer("account"),
			Container:         fastly.ToPointer("container"),
			SASToken:          fastly.ToPointer("token"),
			Period:            fastly.ToPointer(3600),
			TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			PublicKey:         fastly.ToPointer(pgpPublicKey()),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			Placement:         fastly.ToPointer("none"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			CompressionCodec:  fastly.ToPointer("zstd"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			AccountName:       fastly.ToPointer("account"),
			Container:         fastly.ToPointer("analytics"),
			SASToken:          fastly.ToPointer("token"),
			Path:              fastly.ToPointer("/logs"),
			Period:            fastly.ToPointer(86400),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Placement:         fastly.ToPointer("none"),
			PublicKey:         fastly.ToPointer(pgpPublicKey()),
			CompressionCodec:  fastly.ToPointer("zstd"),
			ProcessingRegion:  fastly.ToPointer("us"),
		},
	}, nil
}

func listBlobStoragesError(_ context.Context, _ *fastly.ListBlobStoragesInput) ([]*fastly.BlobStorage, error) {
	return nil, errTest
}

var listBlobStoragesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listBlobStoragesVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

Version: 1
	BlobStorage 1/2
		Service ID: 123
		Version: 1
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
		Processing region: us
	BlobStorage 2/2
		Service ID: 123
		Version: 1
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
		Processing region: us
`) + "\n\n"

func getBlobStorageOK(_ context.Context, i *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error) {
	return &fastly.BlobStorage{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		Container:         fastly.ToPointer("container"),
		AccountName:       fastly.ToPointer("account"),
		SASToken:          fastly.ToPointer("token"),
		Path:              fastly.ToPointer("/logs"),
		Period:            fastly.ToPointer(3600),
		GzipLevel:         fastly.ToPointer(0),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Placement:         fastly.ToPointer("none"),
		ProcessingRegion:  fastly.ToPointer("us"),
		PublicKey:         fastly.ToPointer(pgpPublicKey()),
		CompressionCodec:  fastly.ToPointer("zstd"),
	}, nil
}

func getBlobStorageError(_ context.Context, _ *fastly.GetBlobStorageInput) (*fastly.BlobStorage, error) {
	return nil, errTest
}

var describeBlobStorageOutput = "\n" + strings.TrimSpace(`
Account name: account
Compression codec: zstd
Container: container
File max bytes: 0
Format: %h %l %u %t "%r" %>s %b
Format version: 2
GZip level: 0
Message type: classic
Name: logs
Path: /logs
Period: 3600
Placement: none
Processing region: us
Public key: `+pgpPublicKey()+`
Response condition: Prevent default logging
SAS token: token
Service ID: 123
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Version: 1
`) + "\n"

func updateBlobStorageOK(_ context.Context, i *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error) {
	return &fastly.BlobStorage{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		Container:         fastly.ToPointer("container"),
		AccountName:       fastly.ToPointer("account"),
		SASToken:          fastly.ToPointer("token"),
		Path:              fastly.ToPointer("/logs"),
		Period:            fastly.ToPointer(3600),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Placement:         fastly.ToPointer("none"),
		PublicKey:         fastly.ToPointer(pgpPublicKey()),
		CompressionCodec:  fastly.ToPointer("zstd"),
	}, nil
}

func updateBlobStorageError(_ context.Context, _ *fastly.UpdateBlobStorageInput) (*fastly.BlobStorage, error) {
	return nil, errTest
}

func deleteBlobStorageOK(_ context.Context, _ *fastly.DeleteBlobStorageInput) error {
	return nil
}

func deleteBlobStorageError(_ context.Context, _ *fastly.DeleteBlobStorageInput) error {
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
