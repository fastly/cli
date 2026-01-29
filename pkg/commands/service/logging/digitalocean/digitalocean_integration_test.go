package digitalocean_test

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
	sub "github.com/fastly/cli/pkg/commands/service/logging/digitalocean"
)

func TestDigitalOceanCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key abc --autoclone",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				CreateDigitalOceanFn: createDigitalOceanOK,
			},
			WantOutput: "Created DigitalOcean Spaces logging endpoint log (service 123 version 4)",
		},
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key abc --autoclone",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				CreateDigitalOceanFn: createDigitalOceanError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key abc --compression-codec zstd --gzip-level 9 --autoclone",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			WantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestDigitalOceanList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			WantOutput: listDigitalOceansShortOutput,
		},
		{
			Args: "--service-id 123 --version 1 --verbose",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			WantOutput: listDigitalOceansVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1 -v",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			WantOutput: listDigitalOceansVerboseOutput,
		},
		{
			Args: "--service-id 123 --version 1",
			API: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansError,
			},
			WantError: errTest.Error(),
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestDigitalOceanDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetDigitalOceanFn: getDigitalOceanError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetDigitalOceanFn: getDigitalOceanOK,
			},
			WantOutput: describeDigitalOceanOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestDigitalOceanUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1 --new-name log",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				UpdateDigitalOceanFn: updateDigitalOceanError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --new-name log --autoclone",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				UpdateDigitalOceanFn: updateDigitalOceanOK,
			},
			WantOutput: "Updated DigitalOcean Spaces logging endpoint log (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestDigitalOceanDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args: "--service-id 123 --version 1",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				DeleteDigitalOceanFn: deleteDigitalOceanError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: "--service-id 123 --version 1 --name logs --autoclone",
			API: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				DeleteDigitalOceanFn: deleteDigitalOceanOK,
			},
			WantOutput: "Deleted DigitalOcean Spaces logging endpoint logs (service 123 version 4)",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, parent.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createDigitalOceanOK(_ context.Context, i *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	s := fastly.DigitalOcean{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
	}

	if i.Name != nil {
		s.Name = i.Name
	}

	return &s, nil
}

func createDigitalOceanError(_ context.Context, _ *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

func listDigitalOceansOK(_ context.Context, i *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
	return []*fastly.DigitalOcean{
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("logs"),
			BucketName:        fastly.ToPointer("my-logs"),
			Domain:            fastly.ToPointer("https://digitalocean.us-east-1.amazonaws.com"),
			AccessKey:         fastly.ToPointer("1234"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Path:              fastly.ToPointer("logs/"),
			Period:            fastly.ToPointer(3600),
			GzipLevel:         fastly.ToPointer(9),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			MessageType:       fastly.ToPointer("classic"),
			TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Placement:         fastly.ToPointer("none"),
			PublicKey:         fastly.ToPointer(pgpPublicKey()),
		},
		{
			ServiceID:         fastly.ToPointer(i.ServiceID),
			ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
			Name:              fastly.ToPointer("analytics"),
			BucketName:        fastly.ToPointer("analytics"),
			AccessKey:         fastly.ToPointer("1234"),
			SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
			Domain:            fastly.ToPointer("https://digitalocean.us-east-2.amazonaws.com"),
			Path:              fastly.ToPointer("logs/"),
			Period:            fastly.ToPointer(86400),
			GzipLevel:         fastly.ToPointer(9),
			Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
			FormatVersion:     fastly.ToPointer(2),
			MessageType:       fastly.ToPointer("classic"),
			ResponseCondition: fastly.ToPointer("Prevent default logging"),
			TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
			Placement:         fastly.ToPointer("none"),
			PublicKey:         fastly.ToPointer(pgpPublicKey()),
		},
	}, nil
}

func listDigitalOceansError(_ context.Context, _ *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
	return nil, errTest
}

var listDigitalOceansShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listDigitalOceansVerboseOutput = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Service ID (via --service-id): 123

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

func getDigitalOceanOK(_ context.Context, i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return &fastly.DigitalOcean{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("logs"),
		BucketName:        fastly.ToPointer("my-logs"),
		Domain:            fastly.ToPointer("https://digitalocean.us-east-1.amazonaws.com"),
		AccessKey:         fastly.ToPointer("1234"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Path:              fastly.ToPointer("logs/"),
		Period:            fastly.ToPointer(3600),
		GzipLevel:         fastly.ToPointer(9),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Placement:         fastly.ToPointer("none"),
		PublicKey:         fastly.ToPointer(pgpPublicKey()),
		ProcessingRegion:  fastly.ToPointer("us"),
	}, nil
}

func getDigitalOceanError(_ context.Context, _ *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

var describeDigitalOceanOutput = "\n" + strings.TrimSpace(`
Access key: 1234
Bucket: my-logs
Domain: https://digitalocean.us-east-1.amazonaws.com
Format: %h %l %u %t "%r" %>s %b
Format version: 2
GZip level: 9
Message type: classic
Name: logs
Path: logs/
Period: 3600
Placement: none
Processing region: us
Public key: `+pgpPublicKey()+`
Response condition: Prevent default logging
Secret key: -----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA
Service ID: 123
Timestamp format: %Y-%m-%dT%H:%M:%S.000
Version: 1
`) + "\n"

func updateDigitalOceanOK(_ context.Context, i *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return &fastly.DigitalOcean{
		ServiceID:         fastly.ToPointer(i.ServiceID),
		ServiceVersion:    fastly.ToPointer(i.ServiceVersion),
		Name:              fastly.ToPointer("log"),
		BucketName:        fastly.ToPointer("my-logs"),
		Domain:            fastly.ToPointer("https://digitalocean.us-east-1.amazonaws.com"),
		AccessKey:         fastly.ToPointer("1234"),
		SecretKey:         fastly.ToPointer("-----BEGIN RSA PRIVATE KEY-----MIIEogIBAAKCA"),
		Path:              fastly.ToPointer("logs/"),
		Period:            fastly.ToPointer(3600),
		GzipLevel:         fastly.ToPointer(9),
		Format:            fastly.ToPointer(`%h %l %u %t "%r" %>s %b`),
		FormatVersion:     fastly.ToPointer(2),
		ResponseCondition: fastly.ToPointer("Prevent default logging"),
		MessageType:       fastly.ToPointer("classic"),
		TimestampFormat:   fastly.ToPointer("%Y-%m-%dT%H:%M:%S.000"),
		Placement:         fastly.ToPointer("none"),
		PublicKey:         fastly.ToPointer(pgpPublicKey()),
	}, nil
}

func updateDigitalOceanError(_ context.Context, _ *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

func deleteDigitalOceanOK(_ context.Context, _ *fastly.DeleteDigitalOceanInput) error {
	return nil
}

func deleteDigitalOceanError(_ context.Context, _ *fastly.DeleteDigitalOceanInput) error {
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
