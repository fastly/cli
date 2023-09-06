package cloudfiles_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCloudfilesCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging cloudfiles create --service-id 123 --version 1 --name log --user username --bucket log --access-key foo --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateCloudfilesFn: createCloudfilesOK,
			},
			wantOutput: "Created Cloudfiles logging endpoint log (service 123 version 4)",
		},
		{
			args: args("logging cloudfiles create --service-id 123 --version 1 --name log --user username --bucket log --access-key foo --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				CreateCloudfilesFn: createCloudfilesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging cloudfiles create --service-id 123 --version 1 --name log --user username --bucket log --access-key foo --compression-codec zstd --gzip-level 9 --autoclone"),
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
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestCloudfilesList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: args("logging cloudfiles list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListCloudfilesFn: listCloudfilesOK,
			},
			wantOutput: listCloudfilesShortOutput,
		},
		{
			args: args("logging cloudfiles list --service-id 123 --version 1 --verbose"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListCloudfilesFn: listCloudfilesOK,
			},
			wantOutput: listCloudfilesVerboseOutput,
		},
		{
			args: args("logging cloudfiles list --service-id 123 --version 1 -v"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListCloudfilesFn: listCloudfilesOK,
			},
			wantOutput: listCloudfilesVerboseOutput,
		},
		{
			args: args("logging cloudfiles --verbose list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListCloudfilesFn: listCloudfilesOK,
			},
			wantOutput: listCloudfilesVerboseOutput,
		},
		{
			args: args("logging -v cloudfiles list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListCloudfilesFn: listCloudfilesOK,
			},
			wantOutput: listCloudfilesVerboseOutput,
		},
		{
			args: args("logging cloudfiles list --service-id 123 --version 1"),
			api: mock.API{
				ListVersionsFn:   testutil.ListVersions,
				ListCloudfilesFn: listCloudfilesError,
			},
			wantError: errTest.Error(),
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestCloudfilesDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging cloudfiles describe --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging cloudfiles describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetCloudfilesFn: getCloudfilesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging cloudfiles describe --service-id 123 --version 1 --name logs"),
			api: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				GetCloudfilesFn: getCloudfilesOK,
			},
			wantOutput: describeCloudfilesOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestCloudfilesUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging cloudfiles update --service-id 123 --version 1 --new-name log"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging cloudfiles update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateCloudfilesFn: updateCloudfilesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging cloudfiles update --service-id 123 --version 1 --name logs --new-name log --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				UpdateCloudfilesFn: updateCloudfilesOK,
			},
			wantOutput: "Updated Cloudfiles logging endpoint log (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestCloudfilesDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("logging cloudfiles delete --service-id 123 --version 1"),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: args("logging cloudfiles delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteCloudfilesFn: deleteCloudfilesError,
			},
			wantError: errTest.Error(),
		},
		{
			args: args("logging cloudfiles delete --service-id 123 --version 1 --name logs --autoclone"),
			api: mock.API{
				ListVersionsFn:     testutil.ListVersions,
				CloneVersionFn:     testutil.CloneVersionResult(4),
				DeleteCloudfilesFn: deleteCloudfilesOK,
			},
			wantOutput: "Deleted Cloudfiles logging endpoint logs (service 123 version 4)",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createCloudfilesOK(i *fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error) {
	s := fastly.Cloudfiles{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
	}

	if *i.Name != "" {
		s.Name = *i.Name
	}

	return &s, nil
}

func createCloudfilesError(i *fastly.CreateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return nil, errTest
}

func listCloudfilesOK(i *fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error) {
	return []*fastly.Cloudfiles{
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "logs",
			User:              "username",
			AccessKey:         "1234",
			BucketName:        "my-logs",
			Path:              "logs/",
			Region:            "ORD",
			Placement:         "none",
			Period:            3600,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			MessageType:       "classic",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			PublicKey:         pgpPublicKey(),
		},
		{
			ServiceID:         i.ServiceID,
			ServiceVersion:    i.ServiceVersion,
			Name:              "analytics",
			User:              "username",
			AccessKey:         "1234",
			BucketName:        "analytics",
			Path:              "logs/",
			Region:            "ORD",
			Placement:         "none",
			Period:            86400,
			GzipLevel:         9,
			Format:            `%h %l %u %t "%r" %>s %b`,
			FormatVersion:     2,
			ResponseCondition: "Prevent default logging",
			MessageType:       "classic",
			TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
			PublicKey:         pgpPublicKey(),
		},
	}, nil
}

func listCloudfilesError(i *fastly.ListCloudfilesInput) ([]*fastly.Cloudfiles, error) {
	return nil, errTest
}

var listCloudfilesShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME
123      1        logs
123      1        analytics
`) + "\n"

var listCloudfilesVerboseOutput = strings.TrimSpace(`
Fastly API token provided via config file (profile: user)
Fastly API endpoint: https://api.fastly.com

Service ID (via --service-id): 123

Version: 1
	Cloudfiles 1/2
		Service ID: 123
		Version: 1
		Name: logs
		User: username
		Access key: 1234
		Bucket: my-logs
		Path: logs/
		Region: ORD
		Placement: none
		Period: 3600
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Public key: `+pgpPublicKey()+`
	Cloudfiles 2/2
		Service ID: 123
		Version: 1
		Name: analytics
		User: username
		Access key: 1234
		Bucket: analytics
		Path: logs/
		Region: ORD
		Placement: none
		Period: 86400
		GZip level: 9
		Format: %h %l %u %t "%r" %>s %b
		Format version: 2
		Response condition: Prevent default logging
		Message type: classic
		Timestamp format: %Y-%m-%dT%H:%M:%S.000
		Public key: `+pgpPublicKey()+`
`) + "\n\n"

func getCloudfilesOK(i *fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error) {
	return &fastly.Cloudfiles{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "logs",
		User:              "username",
		AccessKey:         "1234",
		BucketName:        "my-logs",
		Path:              "logs/",
		Region:            "ORD",
		Placement:         "none",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		PublicKey:         pgpPublicKey(),
	}, nil
}

func getCloudfilesError(i *fastly.GetCloudfilesInput) (*fastly.Cloudfiles, error) {
	return nil, errTest
}

var describeCloudfilesOutput = "\n" + strings.TrimSpace(`
Access key: 1234
Bucket: my-logs
Format: %h %l %u %t "%r" %>s %b
Format version: 2
GZip level: 9
Message type: classic
Name: logs
Path: logs/
Period: 3600
Placement: none
Public key: `+pgpPublicKey()+`
Region: ORD
Response condition: Prevent default logging
Service ID: 123
Timestamp format: %Y-%m-%dT%H:%M:%S.000
User: username
Version: 1
`) + "\n"

func updateCloudfilesOK(i *fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return &fastly.Cloudfiles{
		ServiceID:         i.ServiceID,
		ServiceVersion:    i.ServiceVersion,
		Name:              "log",
		User:              "username",
		AccessKey:         "1234",
		BucketName:        "my-logs",
		Path:              "logs/",
		Region:            "ORD",
		Placement:         "none",
		Period:            3600,
		GzipLevel:         9,
		Format:            `%h %l %u %t "%r" %>s %b`,
		FormatVersion:     2,
		ResponseCondition: "Prevent default logging",
		MessageType:       "classic",
		TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
		PublicKey:         pgpPublicKey(),
	}, nil
}

func updateCloudfilesError(i *fastly.UpdateCloudfilesInput) (*fastly.Cloudfiles, error) {
	return nil, errTest
}

func deleteCloudfilesOK(i *fastly.DeleteCloudfilesInput) error {
	return nil
}

func deleteCloudfilesError(i *fastly.DeleteCloudfilesInput) error {
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
