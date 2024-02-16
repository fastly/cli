package serviceauth_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
)

func TestServiceAuthCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth create"),
			wantError: "error parsing arguments: required flag --user-id not provided",
		},
		{
			args:      args("service-auth create --user-id 123 --service-id 123"),
			api:       mock.API{CreateServiceAuthorizationFn: createServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth create --user-id 123 --service-id 123"),
			api:        mock.API{CreateServiceAuthorizationFn: createServiceAuthOK},
			wantOutput: "Created service authorization 12345",
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

func TestServiceAuthList(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth list --verbose --json"),
			wantError: "invalid flag combination, --verbose and --json",
		},
		{
			args:      args("service-auth list"),
			api:       mock.API{ListServiceAuthorizationsFn: listServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth list"),
			api:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			wantOutput: "AUTH ID  USER ID  SERVICE ID  PERMISSION\n123      456      789         read_only\n",
		},
		{
			args: args("service-auth list --json"),
			api:  mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			wantOutput: `{
  "Info": {
    "links": {},
    "meta": {}
  },
  "Items": [
    {
      "CreatedAt": null,
      "DeletedAt": null,
      "ID": "123",
      "Permission": "read_only",
      "Service": {
        "ID": "789"
      },
      "UpdatedAt": null,
      "User": {
        "ID": "456"
      }
    }
  ]
}`,
		},
		{
			args:       args("service-auth list --verbose"),
			api:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			wantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nAuth ID: 123\nUser ID: 456\nService ID: 789\nPermission: read_only\n",
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
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceAuthDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth describe"),
			wantError: "error parsing arguments: required flag --id not provided",
		},
		{
			args:      args("service-auth describe --id 123 --verbose --json"),
			wantError: "invalid flag combination, --verbose and --json",
		},
		{
			args:      args("service-auth describe --id 123"),
			api:       mock.API{GetServiceAuthorizationFn: describeServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth describe --id 123"),
			api:        mock.API{GetServiceAuthorizationFn: describeServiceAuthOK},
			wantOutput: "Auth ID: 12345\nUser ID: 456\nService ID: 789\nPermission: read_only\n",
		},
		{
			args: args("service-auth describe --id 123 --json"),
			api:  mock.API{GetServiceAuthorizationFn: describeServiceAuthOK},
			wantOutput: `{
  "CreatedAt": null,
  "DeletedAt": null,
  "ID": "12345",
  "Permission": "read_only",
  "Service": {
    "ID": "789"
  },
  "UpdatedAt": null,
  "User": {
    "ID": "456"
  }
}`,
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
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

func TestServiceAuthUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth update --permission full"),
			wantError: "error parsing arguments: required flag --id not provided",
		},
		{
			args:      args("service-auth update --id 123"),
			wantError: "error parsing arguments: required flag --permission not provided",
		},
		{
			args:      args("service-auth update --id 123 --permission full"),
			api:       mock.API{UpdateServiceAuthorizationFn: updateServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth update --id 123 --permission full"),
			api:        mock.API{UpdateServiceAuthorizationFn: updateServiceAuthOK},
			wantOutput: "Updated service authorization 123",
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

func TestServiceAuthDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("service-auth delete"),
			wantError: "error parsing arguments: required flag --id not provided",
		},
		{
			args:      args("service-auth delete --id 123"),
			api:       mock.API{DeleteServiceAuthorizationFn: deleteServiceAuthError},
			wantError: errTest.Error(),
		},
		{
			args:       args("service-auth delete --id 123"),
			api:        mock.API{DeleteServiceAuthorizationFn: deleteServiceAuthOK},
			wantOutput: "Deleted service authorization 123",
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

func createServiceAuthError(*fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func createServiceAuthOK(_ *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func listServiceAuthError(*fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
	return nil, errTest
}

func listServiceAuthOK(_ *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
	return &fastly.ServiceAuthorizations{
		Items: []*fastly.ServiceAuthorization{
			{
				ID: "123",
				User: &fastly.SAUser{
					ID: "456",
				},
				Service: &fastly.SAService{
					ID: "789",
				},
				Permission: "read_only",
			},
		},
	}, nil
}

func describeServiceAuthError(*fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func describeServiceAuthOK(_ *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
		User: &fastly.SAUser{
			ID: "456",
		},
		Service: &fastly.SAService{
			ID: "789",
		},
		Permission: "read_only",
	}, nil
}

func updateServiceAuthError(*fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func updateServiceAuthOK(_ *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func deleteServiceAuthError(*fastly.DeleteServiceAuthorizationInput) error {
	return errTest
}

func deleteServiceAuthOK(_ *fastly.DeleteServiceAuthorizationInput) error {
	return nil
}
