package user_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v6/fastly"
)

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --login flag",
			Args:      args("user create --name foobar"),
			WantError: "error parsing arguments: required flag --login not provided",
		},
		{
			Name:      "validate missing --name flag",
			Args:      args("user create --login foo@example.com"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --token flag",
			Args:      args("user create --login foo@example.com --name foobar"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name: "validate CreateUser API error",
			API: mock.API{
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user create --login foo@example.com --name foobar --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateUser API success",
			API: mock.API{
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return &fastly.User{
						Name: i.Name,
						Role: "user",
					}, nil
				},
			},
			Args:       args("user create --login foo@example.com --name foobar --token 123"),
			WantOutput: "Created user 'foobar' (role: user)",
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

func TestDelete(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --id flag",
			Args:      args("user delete"),
			WantError: "error parsing arguments: required flag --id not provided",
		},
		{
			Name:      "validate missing --token flag",
			Args:      args("user delete --id foo123"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name: "validate DeleteUser API error",
			API: mock.API{
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return testutil.Err
				},
			},
			Args:      args("user delete --id foo123 --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteUser API success",
			API: mock.API{
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return nil
				},
			},
			Args:       args("user delete --id foo123 --token 123"),
			WantOutput: "Deleted user (id: foo123)",
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

func TestDescribe(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --token flag",
			Args:      args("user describe"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name:      "validate missing --id flag",
			Args:      args("user describe --token 123"),
			WantError: "error parsing arguments: must provide --id flag",
		},
		{
			Name: "validate GetUser API error",
			API: mock.API{
				GetUserFn: func(i *fastly.GetUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user describe --id 123 --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetCurrentUser API error",
			API: mock.API{
				GetCurrentUserFn: func() (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user describe --current --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetUser API success",
			API: mock.API{
				GetUserFn: getUser,
			},
			Args:       args("user describe --id 123 --token 123"),
			WantOutput: describeUserOutput(),
		},
		{
			Name: "validate GetCurrentUser API success",
			API: mock.API{
				GetCurrentUserFn: getCurrentUser,
			},
			Args:       args("user describe --current --token 123"),
			WantOutput: describeCurrentUserOutput(),
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

func TestList(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --token flag",
			Args:      args("user list --customer-id abc"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name:      "validate missing --customer-id flag",
			Args:      args("user list --token 123"),
			WantError: "error reading customer ID: no customer ID found",
		},
		{
			Name: "validate ListUsers API error",
			API: mock.API{
				ListCustomerUsersFn: func(i *fastly.ListCustomerUsersInput) ([]*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user list --customer-id abc --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListUsers API success",
			API: mock.API{
				ListCustomerUsersFn: listUsers,
			},
			Args:       args("user list --customer-id abc --token 123"),
			WantOutput: listOutput(),
		},
		{
			Name: "validate ListUsers API success with verbose mode",
			API: mock.API{
				ListCustomerUsersFn: listUsers,
			},
			Args:       args("user list --customer-id abc --token 123 --verbose"),
			WantOutput: listVerboseOutput(),
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

func TestUpdate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --token flag",
			Args:      args("user update --id 123"),
			WantError: errors.ErrNoToken.Inner.Error(),
		},
		{
			Name:      "validate missing --id flag",
			Args:      args("user update --token 123"),
			WantError: "error parsing arguments: must provide --id flag",
		},
		{
			Name:      "validate missing --name and --role flags",
			Args:      args("user update --id 123 --token 123"),
			WantError: "error parsing arguments: must provide either the --name or --role with the --id flag",
		},
		{
			Name:      "validate missing --login flag with --password-reset",
			Args:      args("user update --password-reset --token 123"),
			WantError: "error parsing arguments: must provide --login when requesting a password reset",
		},
		{
			Name:      "validate invalid --role value",
			Args:      args("user update --id 123 --role foobar --token 123"),
			WantError: "error parsing arguments: enum value must be one of user,billing,engineer,superuser, got 'foobar'",
		},
		{
			Name: "validate UpdateUser API error",
			API: mock.API{
				UpdateUserFn: func(i *fastly.UpdateUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user update --id 123 --name foo --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ResetUserPassword API error",
			API: mock.API{
				ResetUserPasswordFn: func(i *fastly.ResetUserPasswordInput) error {
					return testutil.Err
				},
			},
			Args:      args("user update --id 123 --login foo@example.com --password-reset --token 123"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateUser API success",
			API: mock.API{
				UpdateUserFn: func(i *fastly.UpdateUserInput) (*fastly.User, error) {
					return &fastly.User{
						ID:   i.ID,
						Name: *i.Name,
						Role: *i.Role,
					}, nil
				},
			},
			Args:       args("user update --id 123 --name foo --role engineer --token 123"),
			WantOutput: "Updated user 'foo' (role: engineer)",
		},
		{
			Name: "validate ResetUserPassword API success",
			API: mock.API{
				ResetUserPasswordFn: func(i *fastly.ResetUserPasswordInput) error {
					return nil
				},
			},
			Args:       args("user update --id 123 --login foo@example.com --password-reset --token 123"),
			WantOutput: "Reset user password (login: foo@example.com)",
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

func getUser(i *fastly.GetUserInput) (*fastly.User, error) {
	t := testutil.Date

	return &fastly.User{
		ID:                     i.ID,
		Login:                  "foo@example.com",
		Name:                   "foo",
		Role:                   "user",
		CustomerID:             "abc",
		EmailHash:              "example-hash",
		LimitServices:          true,
		Locked:                 true,
		RequireNewPassword:     true,
		TwoFactorAuthEnabled:   true,
		TwoFactorSetupRequired: true,
		CreatedAt:              &t,
		DeletedAt:              &t,
		UpdatedAt:              &t,
	}, nil
}

func getCurrentUser() (*fastly.User, error) {
	t := testutil.Date

	return &fastly.User{
		ID:                     "current123",
		Login:                  "bar@example.com",
		Name:                   "bar",
		Role:                   "superuser",
		CustomerID:             "abc",
		EmailHash:              "example-hash2",
		LimitServices:          false,
		Locked:                 false,
		RequireNewPassword:     false,
		TwoFactorAuthEnabled:   false,
		TwoFactorSetupRequired: false,
		CreatedAt:              &t,
		DeletedAt:              &t,
		UpdatedAt:              &t,
	}, nil
}

func listUsers(i *fastly.ListCustomerUsersInput) ([]*fastly.User, error) {
	user, _ := getUser(&fastly.GetUserInput{ID: "123"})
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
	return fmt.Sprintf(`Fastly API token provided via --token
Fastly API endpoint: https://api.fastly.com
%s%s`, describeUserOutput(), describeCurrentUserOutput())
}
