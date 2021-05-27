package backend_test

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

func TestBackendCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"backend", "create", "--version", "1", "--service-id", "123", "--address", "example.com"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		// The following test mocks a service version that's 'active', and
		// subsequently we expect it to not be cloned as we don't provide the
		// --autoclone flag and trying to add a backend to an activated service
		// should cause an error.
		{
			args: []string{"backend", "create", "--service-id", "123", "--version", "1", "--address", "example.com", "--name", "www.test.com"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				CreateBackendFn: createBackendError,
			},
			wantError: errTest.Error(),
		},
		// The following test is the same as above but with an IP address for the
		// --address flag instead of a hostname.
		{
			args: []string{"backend", "create", "--service-id", "123", "--version", "1", "--address", "127.0.0.1", "--name", "www.test.com"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				CreateBackendFn: createBackendError,
			},
			wantError: errTest.Error(),
		},
		// The following test mocks a service version that's 'inactive', and
		// subsequently we expect it to be the same editable version 1.
		{
			args: []string{"backend", "create", "--service-id", "123", "--version", "1", "--address", "127.0.0.1", "--name", "www.test.com"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetInactiveVersionOK,
				CloneVersionFn:  testutil.CloneVersionOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: "Created backend www.test.com (service 123 version 1)",
		},
		// The following test mocks a service version that's 'active', and
		// subsequently we expect it to be cloned to version 2.
		{
			args: []string{"backend", "create", "--service-id", "123", "--version", "1", "--address", "127.0.0.1", "--name", "www.test.com", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				CloneVersionFn:  testutil.CloneVersionOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: "Created backend www.test.com (service 123 version 2)",
		},
		{
			args: []string{"backend", "create", "--service-id", "123", "--version", "1", "--address", "127.0.0.1", "--name", "www.test.com", "--use-ssl", "--verbose"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				CreateBackendFn: createBackendOK,
			},
			wantOutput: "Use-ssl was set but no port was specified, so default port 443 will be used",
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

func TestBackendList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args: []string{"backend", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				ListBackendsFn: listBackendsOK,
			},
			wantOutput: listBackendsShortOutput,
		},
		{
			args: []string{"backend", "list", "--service-id", "123", "--version", "1", "--verbose"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				ListBackendsFn: listBackendsOK,
			},
			wantOutput: listBackendsVerboseOutput,
		},
		{
			args: []string{"backend", "list", "--service-id", "123", "--version", "1", "-v"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				ListBackendsFn: listBackendsOK,
			},
			wantOutput: listBackendsVerboseOutput,
		},
		{
			args: []string{"backend", "--verbose", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				ListBackendsFn: listBackendsOK,
			},
			wantOutput: listBackendsVerboseOutput,
		},
		{
			args: []string{"-v", "backend", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				ListBackendsFn: listBackendsOK,
			},
			wantOutput: listBackendsVerboseOutput,
		},
		{
			args: []string{"backend", "list", "--service-id", "123", "--version", "1"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				ListBackendsFn: listBackendsError,
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

func TestBackendDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"backend", "describe", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"backend", "describe", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				GetBackendFn:   getBackendError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"backend", "describe", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				GetBackendFn:   getBackendOK,
			},
			wantOutput: describeBackendOutput,
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

func TestBackendUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"backend", "update", "--service-id", "123", "--version", "2", "--new-name", "www.test.com", "--comment", ""},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"backend", "update", "--service-id", "123", "--version", "2", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				ListVersionsFn: testutil.ListVersionsOk,
				GetVersionFn:   testutil.GetActiveVersionOK,
				GetBackendFn:   getBackendError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"backend", "update", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--new-name", "www.example.com"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				GetBackendFn:    getBackendOK,
				UpdateBackendFn: updateBackendError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"backend", "update", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--new-name", "www.example.com", "--comment", "", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				CloneVersionFn:  testutil.CloneVersionOK,
				GetBackendFn:    getBackendOK,
				UpdateBackendFn: updateBackendOK,
			},
			wantOutput: "Updated backend www.example.com (service 123 version 2)",
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

func TestBackendDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"backend", "delete", "--service-id", "123", "--version", "1"},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: []string{"backend", "delete", "--service-id", "123", "--version", "1", "--name", "www.test.com"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				DeleteBackendFn: deleteBackendError,
			},
			wantError: errTest.Error(),
		},
		{
			args: []string{"backend", "delete", "--service-id", "123", "--version", "1", "--name", "www.test.com", "--autoclone"},
			api: mock.API{
				ListVersionsFn:  testutil.ListVersionsOk,
				GetVersionFn:    testutil.GetActiveVersionOK,
				CloneVersionFn:  testutil.CloneVersionOK,
				DeleteBackendFn: deleteBackendOK,
			},
			wantOutput: "Deleted backend www.test.com (service 123 version 2)",
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
	"Service ID: 123",
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
	"		SSL ciphers: []",
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
	"		SSL ciphers: []",
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
	"SSL ciphers: []",
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
