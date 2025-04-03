package user_test

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/user"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestUserCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "validate CreateUser API error",
			API: mock.API{
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--login foo@example.com --name foobar",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateUser API success",
			API: mock.API{
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return &fastly.User{
						Name: i.Name,
						Role: fastly.ToPointer("user"),
					}, nil
				},
			},
			Args:       "--login foo@example.com --name foobar",
			WantOutput: "Created user 'foobar' (role: user)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestUserDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name: "validate DeleteUser API error",
			API: mock.API{
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return testutil.Err
				},
			},
			Args:      "--id foo123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteUser API success",
			API: mock.API{
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return nil
				},
			},
			Args:       "--id foo123",
			WantOutput: "Deleted user (id: foo123)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestUserDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: must provide --id flag",
		},
		{
			Name: "validate GetUser API error",
			API: mock.API{
				GetUserFn: func(i *fastly.GetUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id 123",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetCurrentUser API error",
			API: mock.API{
				GetCurrentUserFn: func() (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--current",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetUser API success",
			API: mock.API{
				GetUserFn: getUser,
			},
			Args:       "--id 123",
			WantOutput: describeUserOutput(),
		},
		{
			Name: "validate GetCurrentUser API success",
			API: mock.API{
				GetCurrentUserFn: getCurrentUser,
			},
			Args:       "--current",
			WantOutput: describeCurrentUserOutput(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestUserList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --customer-id flag",
			WantError: "error reading customer ID: no customer ID found",
		},
		{
			Name: "validate ListUsers API error",
			API: mock.API{
				ListCustomerUsersFn: func(i *fastly.ListCustomerUsersInput) ([]*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--customer-id abc",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListUsers API success",
			API: mock.API{
				ListCustomerUsersFn: listUsers,
			},
			Args:       "--customer-id abc",
			WantOutput: listOutput(),
		},
		{
			Name: "validate ListUsers API success with verbose mode",
			API: mock.API{
				ListCustomerUsersFn: listUsers,
			},
			Args:       "--customer-id abc --verbose",
			WantOutput: listVerboseOutput(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestUserUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --id flag",
			WantError: "error parsing arguments: must provide --id flag",
		},
		{
			Name:      "validate missing --name and --role flags",
			Args:      "--id 123",
			WantError: "error parsing arguments: must provide either the --name or --role with the --id flag",
		},
		{
			Name:      "validate missing --login flag with --password-reset",
			Args:      "--password-reset",
			WantError: "error parsing arguments: must provide --login when requesting a password reset",
		},
		{
			Name:      "validate invalid --role value",
			Args:      "--id 123 --role foobar",
			WantError: "error parsing arguments: enum value must be one of user,billing,engineer,superuser, got 'foobar'",
		},
		{
			Name: "validate UpdateUser API error",
			API: mock.API{
				UpdateUserFn: func(i *fastly.UpdateUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--id 123 --name foo",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ResetUserPassword API error",
			API: mock.API{
				ResetUserPasswordFn: func(i *fastly.ResetUserPasswordInput) error {
					return testutil.Err
				},
			},
			Args:      "--id 123 --login foo@example.com --password-reset",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateUser API success",
			API: mock.API{
				UpdateUserFn: func(i *fastly.UpdateUserInput) (*fastly.User, error) {
					return &fastly.User{
						UserID: fastly.ToPointer(i.UserID),
						Name:   i.Name,
						Role:   i.Role,
					}, nil
				},
			},
			Args:       "--id 123 --name foo --role engineer",
			WantOutput: "Updated user 'foo' (role: engineer)",
		},
		{
			Name: "validate ResetUserPassword API success",
			API: mock.API{
				ResetUserPasswordFn: func(i *fastly.ResetUserPasswordInput) error {
					return nil
				},
			},
			Args:       "--id 123 --login foo@example.com --password-reset",
			WantOutput: "Reset user password (login: foo@example.com)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func getUser(i *fastly.GetUserInput) (*fastly.User, error) {
	t := testutil.Date

	return &fastly.User{
		UserID:                 fastly.ToPointer(i.UserID),
		Login:                  fastly.ToPointer("foo@example.com"),
		Name:                   fastly.ToPointer("foo"),
		Role:                   fastly.ToPointer("user"),
		CustomerID:             fastly.ToPointer("abc"),
		EmailHash:              fastly.ToPointer("example-hash"),
		LimitServices:          fastly.ToPointer(true),
		Locked:                 fastly.ToPointer(true),
		RequireNewPassword:     fastly.ToPointer(true),
		TwoFactorAuthEnabled:   fastly.ToPointer(true),
		TwoFactorSetupRequired: fastly.ToPointer(true),
		CreatedAt:              &t,
		DeletedAt:              &t,
		UpdatedAt:              &t,
	}, nil
}

func getCurrentUser() (*fastly.User, error) {
	t := testutil.Date

	return &fastly.User{
		UserID:                 fastly.ToPointer("current123"),
		Login:                  fastly.ToPointer("bar@example.com"),
		Name:                   fastly.ToPointer("bar"),
		Role:                   fastly.ToPointer("superuser"),
		CustomerID:             fastly.ToPointer("abc"),
		EmailHash:              fastly.ToPointer("example-hash2"),
		LimitServices:          fastly.ToPointer(false),
		Locked:                 fastly.ToPointer(false),
		RequireNewPassword:     fastly.ToPointer(false),
		TwoFactorAuthEnabled:   fastly.ToPointer(false),
		TwoFactorSetupRequired: fastly.ToPointer(false),
		CreatedAt:              &t,
		DeletedAt:              &t,
		UpdatedAt:              &t,
	}, nil
}

func listUsers(_ *fastly.ListCustomerUsersInput) ([]*fastly.User, error) {
	user, _ := getUser(&fastly.GetUserInput{UserID: "123"})
	userCurrent, _ := getCurrentUser()
	vs := []*fastly.User{
		user,
		userCurrent,
	}
	return vs, nil
}

func describeUserOutput() string {
	return `
ID: 123
Login: foo@example.com
Name: foo
Role: user
Customer ID: abc
Email Hash: example-hash
Limit Services: true
Locked: true
Require New Password: true
Two Factor Auth Enabled: true
Two Factor Setup Required: true

Created at: 2021-06-15 23:00:00 +0000 UTC
Updated at: 2021-06-15 23:00:00 +0000 UTC
Deleted at: 2021-06-15 23:00:00 +0000 UTC
`
}

func describeCurrentUserOutput() string {
	return `
ID: current123
Login: bar@example.com
Name: bar
Role: superuser
Customer ID: abc
Email Hash: example-hash2
Limit Services: false
Locked: false
Require New Password: false
Two Factor Auth Enabled: false
Two Factor Setup Required: false

Created at: 2021-06-15 23:00:00 +0000 UTC
Updated at: 2021-06-15 23:00:00 +0000 UTC
Deleted at: 2021-06-15 23:00:00 +0000 UTC
`
}

func listOutput() string {
	return `LOGIN            NAME  ROLE       LOCKED  ID
foo@example.com  foo   user       true    123
bar@example.com  bar   superuser  false   current123
`
}

func listVerboseOutput() string {
	return fmt.Sprintf(`Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

%s%s`, describeUserOutput(), describeCurrentUserOutput())
}
