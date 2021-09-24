package backend_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v5/fastly"
)

func TestBackendCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("backend create --version 1 --service-id 123 --address example.com"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		// The following test specifies a service version that's 'active', and
		// subsequently we expect it to not be cloned as we don't provide the
		// --autoclone flag and trying to add a backend to an activated service
		// should cause an error.
		{
			Args: args("backend create --service-id 123 --version 1 --address example.com --name www.test.com"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			WantError: "service version 1 is not editable",
		},
		// The following test is the same as the above but it appends --autoclone
		// so we can be sure the backend creation error still occurs.
		{
			Args: args("backend create --service-id 123 --version 1 --address example.com --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateBackendFn: createBackendError,
			},
			WantError: errTest.Error(),
		},
		// The following test is the same as above but with an IP address for the
		// --address flag instead of a hostname.
		{
			Args: args("backend create --service-id 123 --version 1 --address 127.0.0.1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateBackendFn: createBackendError,
			},
			WantError: errTest.Error(),
		},
		// The following test is the same as above but mocks a successful backend
		// creation so we can validate the correct service version was utilised.
		{
			Args: args("backend create --service-id 123 --version 1 --address 127.0.0.1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateBackendFn: createBackendOK,
			},
			WantOutput: "Created backend www.test.com (service 123 version 4)",
		},
		// The following test is the same as above but appends both --use-ssl and
		// --verbose so we may validate the expected output message regarding a
		// missing port is displayed.
		{
			Args: args("backend create --service-id 123 --version 1 --address 127.0.0.1 --name www.test.com --autoclone --use-ssl --verbose"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				CreateBackendFn: createBackendOK,
			},
			WantOutput: "Use-ssl was set but no port was specified, using default port 443",
		},
		// The following test specifies a service version that's 'inactive', and
		// subsequently we expect it to be the same editable version.
		{
			Args: args("backend create --service-id 123 --version 3 --address 127.0.0.1 --name www.test.com"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CreateBackendFn: createBackendOK,
			},
			WantOutput: "Created backend www.test.com (service 123 version 3)",
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestBackendList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args: args("backend list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListBackendsFn: listBackendsOK,
			},
			WantOutput: listBackendsShortOutput,
		},
		{
			Args: args("backend list --service-id 123 --version 1 --verbose"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListBackendsFn: listBackendsOK,
			},
			WantOutput: listBackendsVerboseOutput,
		},
		{
			Args: args("backend list --service-id 123 --version 1 -v"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListBackendsFn: listBackendsOK,
			},
			WantOutput: listBackendsVerboseOutput,
		},
		{
			Args: args("backend --verbose list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListBackendsFn: listBackendsOK,
			},
			WantOutput: listBackendsVerboseOutput,
		},
		{
			Args: args("-v backend list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListBackendsFn: listBackendsOK,
			},
			WantOutput: listBackendsVerboseOutput,
		},
		{
			Args: args("backend list --service-id 123 --version 1"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListBackendsFn: listBackendsError,
			},
			WantError: errTest.Error(),
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestBackendDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("backend describe --service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("backend describe --service-id 123 --version 1 --name www.test.com"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetBackendFn:   getBackendError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("backend describe --service-id 123 --version 1 --name www.test.com"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetBackendFn:   getBackendOK,
			},
			WantOutput: describeBackendOutput,
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestBackendUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("backend update --service-id 123 --version 2 --new-name www.test.com --comment "),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("backend update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				GetBackendFn:   getBackendError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("backend update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				GetBackendFn:    getBackendOK,
				UpdateBackendFn: updateBackendError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("backend update --service-id 123 --version 1 --name www.test.com --new-name www.example.com --comment  --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				GetBackendFn:    getBackendOK,
				UpdateBackendFn: updateBackendOK,
			},
			WantOutput: "Updated backend www.example.com (service 123 version 4)",
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

func TestBackendDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Args:      args("backend delete --service-id 123 --version 1"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: args("backend delete --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteBackendFn: deleteBackendError,
			},
			WantError: errTest.Error(),
		},
		{
			Args: args("backend delete --service-id 123 --version 1 --name www.test.com --autoclone"),
			API: mock.API{
				ListVersionsFn:  testutil.ListVersions,
				CloneVersionFn:  testutil.CloneVersionResult(4),
				DeleteBackendFn: deleteBackendOK,
			},
			WantOutput: "Deleted backend www.test.com (service 123 version 4)",
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}

var errTest = errors.New("fixture error")

func createBackendOK(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           i.Name,
		Comment:        i.Comment,
	}, nil
}

func createBackendError(i *fastly.CreateBackendInput) (*fastly.Backend, error) {
	return nil, errTest
}

func listBackendsOK(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return []*fastly.Backend{
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "test.com",
			Address:        "www.test.com",
			Port:           80,
			Comment:        "test",
		},
		{
			ServiceID:      i.ServiceID,
			ServiceVersion: i.ServiceVersion,
			Name:           "example.com",
			Address:        "www.example.com",
			Port:           443,
			Comment:        "example",
		},
	}, nil
}

func listBackendsError(i *fastly.ListBackendsInput) ([]*fastly.Backend, error) {
	return nil, errTest
}

var listBackendsShortOutput = strings.TrimSpace(`
SERVICE  VERSION  NAME         ADDRESS          PORT  COMMENT
123      1        test.com     www.test.com     80    test
123      1        example.com  www.example.com  443   example
`) + "\n"

var listBackendsVerboseOutput = strings.Join([]string{
	"Fastly API token not provided",
	"Fastly API endpoint: https://api.fastly.com",
	"Service ID (via --service-id): 123",
	"",
	"Version: 1",
	"	Backend 1/2",
	"		Name: test.com",
	"		Comment: test",
	"		Address: www.test.com",
	"		Port: 80",
	"		Override host: ",
	"		Connect timeout: 0",
	"		Max connections: 0",
	"		First byte timeout: 0",
	"		Between bytes timeout: 0",
	"		Auto loadbalance: false",
	"		Weight: 0",
	"		Healthcheck: ",
	"		Shield: ",
	"		Use SSL: false",
	"		SSL check cert: false",
	"		SSL CA cert: ",
	"		SSL client cert: ",
	"		SSL client key: ",
	"		SSL cert hostname: ",
	"		SSL SNI hostname: ",
	"		Min TLS version: ",
	"		Max TLS version: ",
	"		SSL ciphers: ",
	"	Backend 2/2",
	"		Name: example.com",
	"		Comment: example",
	"		Address: www.example.com",
	"		Port: 443",
	"		Override host: ",
	"		Connect timeout: 0",
	"		Max connections: 0",
	"		First byte timeout: 0",
	"		Between bytes timeout: 0",
	"		Auto loadbalance: false",
	"		Weight: 0",
	"		Healthcheck: ",
	"		Shield: ",
	"		Use SSL: false",
	"		SSL check cert: false",
	"		SSL CA cert: ",
	"		SSL client cert: ",
	"		SSL client key: ",
	"		SSL cert hostname: ",
	"		SSL SNI hostname: ",
	"		Min TLS version: ",
	"		Max TLS version: ",
	"		SSL ciphers: ",
}, "\n") + "\n\n"

func getBackendOK(i *fastly.GetBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           "test.com",
		Address:        "www.test.com",
		Port:           80,
		Comment:        "test",
	}, nil
}

func getBackendError(i *fastly.GetBackendInput) (*fastly.Backend, error) {
	return nil, errTest
}

var describeBackendOutput = strings.Join([]string{
	"Service ID: 123",
	"Version: 1",
	"Name: test.com",
	"Comment: test",
	"Address: www.test.com",
	"Port: 80",
	"Override host: ",
	"Connect timeout: 0",
	"Max connections: 0",
	"First byte timeout: 0",
	"Between bytes timeout: 0",
	"Auto loadbalance: false",
	"Weight: 0",
	"Healthcheck: ",
	"Shield: ",
	"Use SSL: false",
	"SSL check cert: false",
	"SSL CA cert: ",
	"SSL client cert: ",
	"SSL client key: ",
	"SSL cert hostname: ",
	"SSL SNI hostname: ",
	"Min TLS version: ",
	"Max TLS version: ",
	"SSL ciphers: ",
}, "\n") + "\n"

func updateBackendOK(i *fastly.UpdateBackendInput) (*fastly.Backend, error) {
	return &fastly.Backend{
		ServiceID:      i.ServiceID,
		ServiceVersion: i.ServiceVersion,
		Name:           *i.NewName,
		Comment:        *i.Comment,
	}, nil
}

func updateBackendError(i *fastly.UpdateBackendInput) (*fastly.Backend, error) {
	return nil, errTest
}

func deleteBackendOK(i *fastly.DeleteBackendInput) error {
	return nil
}

func deleteBackendError(i *fastly.DeleteBackendInput) error {
	return errTest
}
