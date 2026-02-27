package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/auth"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestServiceAuthCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --user-id not provided",
		},
		{
			Args:      "--user-id 123 --service-id 123",
			API:       mock.API{CreateServiceAuthorizationFn: createServiceAuthError},
			WantError: errTest.Error(),
		},
		{
			Args:       "--user-id 123 --service-id 123",
			API:        mock.API{CreateServiceAuthorizationFn: createServiceAuthOK},
			WantOutput: "Created service authorization 12345",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestServiceAuthList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args:      "",
			API:       mock.API{ListServiceAuthorizationsFn: listServiceAuthError},
			WantError: errTest.Error(),
		},
		{
			Args:       "",
			API:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			WantOutput: "AUTH ID  USER ID  SERVICE ID  PERMISSION\n123      456      789         read_only\n",
		},
		{
			Args: "--json",
			API:  mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			WantOutput: `{
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
			Args:       "--verbose",
			API:        mock.API{ListServiceAuthorizationsFn: listServiceAuthOK},
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nAuth ID: 123\nUser ID: 456\nService ID: 789\nPermission: read_only\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestServiceAuthDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Args:      "--id 123 --verbose --json",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Args:      "--id 123",
			API:       mock.API{GetServiceAuthorizationFn: describeServiceAuthError},
			WantError: errTest.Error(),
		},
		{
			Args:       "--id 123",
			API:        mock.API{GetServiceAuthorizationFn: describeServiceAuthOK},
			WantOutput: "Auth ID: 12345\nUser ID: 456\nService ID: 789\nPermission: read_only\n",
		},
		{
			Args: "--id 123 --json",
			API:  mock.API{GetServiceAuthorizationFn: describeServiceAuthOK},
			WantOutput: `{
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

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestServiceAuthUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--permission full",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Args:      "--id 123",
			WantError: "error parsing arguments: required flag --permission not provided",
		},
		{
			Args:      "--id 123 --permission full",
			API:       mock.API{UpdateServiceAuthorizationFn: updateServiceAuthError},
			WantError: errTest.Error(),
		},
		{
			Args:       "--id 123 --permission full",
			API:        mock.API{UpdateServiceAuthorizationFn: updateServiceAuthOK},
			WantOutput: "Updated service authorization 123",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func TestServiceAuthDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Args:      "--id 123",
			API:       mock.API{DeleteServiceAuthorizationFn: deleteServiceAuthError},
			WantError: errTest.Error(),
		},
		{
			Args:       "--id 123",
			API:        mock.API{DeleteServiceAuthorizationFn: deleteServiceAuthOK},
			WantOutput: "Deleted service authorization 123",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

var errTest = errors.New("fixture error")

func createServiceAuthError(_ context.Context, _ *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func createServiceAuthOK(_ context.Context, _ *fastly.CreateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func listServiceAuthError(_ context.Context, _ *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
	return nil, errTest
}

func listServiceAuthOK(_ context.Context, _ *fastly.ListServiceAuthorizationsInput) (*fastly.ServiceAuthorizations, error) {
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

func describeServiceAuthError(_ context.Context, _ *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func describeServiceAuthOK(_ context.Context, _ *fastly.GetServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
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

func updateServiceAuthError(_ context.Context, _ *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return nil, errTest
}

func updateServiceAuthOK(_ context.Context, _ *fastly.UpdateServiceAuthorizationInput) (*fastly.ServiceAuthorization, error) {
	return &fastly.ServiceAuthorization{
		ID: "12345",
	}, nil
}

func deleteServiceAuthError(_ context.Context, _ *fastly.DeleteServiceAuthorizationInput) error {
	return errTest
}

func deleteServiceAuthOK(_ context.Context, _ *fastly.DeleteServiceAuthorizationInput) error {
	return nil
}
