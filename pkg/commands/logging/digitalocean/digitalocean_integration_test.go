package digitalocean_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDigitalOceanCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging digitalocean create --service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key abc --autoclone"),
			api: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				CreateDigitalOceanFn: createDigitalOceanOK,
			},
			wantOutput: "Created DigitalOcean Spaces logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging digitalocean create --service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key abc --autoclone"),
			api: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				CreateDigitalOceanFn: createDigitalOceanError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging digitalocean create --service-id 123 --version 1 --name log --bucket log --access-key foo --secret-key abc --compression-codec zstd --gzip-level 9 --autoclone"),
			api: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
			},
			wantError: "error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestDigitalOceanList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging digitalocean list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			wantOutput: listDigitalOceansShortOutput,
		},
		{
			args: args("logging digitalocean list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args: args("logging digitalocean list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args: args("logging digitalocean --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args: args("logging -v digitalocean list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansOK,
			},
			wantOutput: listDigitalOceansVerboseOutput,
		},
		{
			args: args("logging digitalocean list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:      testutil.ListVersions,
				ListDigitalOceansFn: listDigitalOceansError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDigitalOceanDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging digitalocean describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging digitalocean describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetDigitalOceanFn: getDigitalOceanError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging digitalocean describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetDigitalOceanFn: getDigitalOceanOK,
			},
			wantOutput: describeDigitalOceanOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDigitalOceanUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging digitalocean update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging digitalocean update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				UpdateDigitalOceanFn: updateDigitalOceanError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging digitalocean update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				UpdateDigitalOceanFn: updateDigitalOceanOK,
			},
			wantOutput: "Updated DigitalOcean Spaces logging endpoint log (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestDigitalOceanDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging digitalocean delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging digitalocean delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				DeleteDigitalOceanFn: deleteDigitalOceanError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging digitalocean delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:       testutil.ListVersions,
				CloneVersionFn:       testutil.CloneVersionResult(4),
				DeleteDigitalOceanFn: deleteDigitalOceanOK,
			},
			wantOutput: "Deleted DigitalOcean Spaces logging endpoint logs (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createDigitalOceanOK(i *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	s := fastly.DigitalOcean{
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
	}

	if i.Name != nil {
		s.Name = i.Name
	}

	return &s, nil
}

func createDigitalOceanError(_ *fastly.CreateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

func listDigitalOceansOK(i *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
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

func listDigitalOceansError(_ *fastly.ListDigitalOceansInput) ([]*fastly.DigitalOcean, error) {
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

func getDigitalOceanOK(i *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
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

func getDigitalOceanError(_ *fastly.GetDigitalOceanInput) (*fastly.DigitalOcean, error) {
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

func updateDigitalOceanOK(i *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
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

func updateDigitalOceanError(_ *fastly.UpdateDigitalOceanInput) (*fastly.DigitalOcean, error) {
	return nil, errTest
}

func deleteDigitalOceanOK(_ *fastly.DeleteDigitalOceanInput) error {
	return nil
}

func deleteDigitalOceanError(_ *fastly.DeleteDigitalOceanInput) error {
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
