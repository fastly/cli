package user_test

import (
	"bytes"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestCreate(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			Args:      args("user create"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("user create --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("user create --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate CreateUser API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user create --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateUser API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return &fastly.User{
						ServiceID: i.ServiceID,
					}, nil
				},
			},
			Args:       args("user create --service-id 123 --version 3"),
			WantOutput: "Created <...> '456' (service: 123)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateUserFn: func(i *fastly.CreateUserInput) (*fastly.User, error) {
					return &fastly.VCL{
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("user create --autoclone --service-id 123 --version 1"),
			WantOutput: "Created <...> 'foo' (service: 123, version: 4)",
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
			Name:      "validate missing --version flag",
			Args:      args("User delete"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("user delete --version 1"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("User delete ---service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate DeleteUser API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return testutil.Err
				},
			},
			Args:      args("user delete --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteUser API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return nil
				},
			},
			Args:       args("user delete --service-id 123 --version 3"),
			WantOutput: "Deleted <...> '456' (service: 123)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteUserFn: func(i *fastly.DeleteUserInput) error {
					return nil
				},
			},
			Args:       args("user delete --autoclone --service-id 123 --version 1"),
			WantOutput: "Deleted <...> 'foo' (service: 123, version: 4)",
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
			Name:      "validate missing --version flag",
			Args:      args("User describe"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("user describe --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetUser API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetUserFn: func(i *fastly.GetUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user describe --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetUser API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetUserFn: getUser,
			},
			Args:       args("user describe --service-id 123 --version 3"),
			WantOutput: "<...>",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetVCLFn:       getVCL,
			},
			Args:       args("User describe --service-id 123 --version 1"),
			WantOutput: "<...>",
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
			Name:      "validate missing --version flag",
			Args:      args("User list"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("user list --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListUsers API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListUsersFn: func(i *fastly.ListUsersInput) ([]*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user list --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListUsers API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListUsersFn: listUsers,
			},
			Args:       args("user list --service-id 123 --version 3"),
			WantOutput: "<...>",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListUsersFn:     listUsers,
			},
			Args:       args("user list --service-id 123 --version 1"),
			WantOutput: "<...>",
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListUsersFn: listUsers,
			},
			Args:       args("user list --acl-id 123 --service-id 123 --verbose"),
			WantOutput: "<...>",
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
			Name:      "validate missing --name flag",
			Args:      args("user update --version 3"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      args("user update --name foobar"),
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      args("user update --name foobar --version 3"),
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Args:      args("user update --name foobar --service-id 123 --version 1"),
			WantError: "service version 1 is not editable",
		},
		{
			Name: "validate UpdateUser API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateUserFn: func(i *fastly.UpdateUserInput) (*fastly.User, error) {
					return nil, testutil.Err
				},
			},
			Args:      args("user update --name foobar --service-id 123 --version 3"),
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateUser API success with --new-name",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateUserFn: func(i *fastly.UpdateUserInput) (*fastly.User, error) {
					return &fastly.User{
						Name:           *i.NewName,
						ServiceID:      i.ServiceID,
						ServiceVersion: i.ServiceVersion,
					}, nil
				},
			},
			Args:       args("user update --name foobar --new-name beepboop --service-id 123 --version 3"),
			WantOutput: "Updated <...> 'beepboop' (previously: 'foobar', service: 123, version: 3)",
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
		ServiceID: i.ServiceID,

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listUsers(i *fastly.ListUsersInput) ([]*fastly.User, error) {
	t := testutil.Date
	vs := []*fastly.User{
		{
			ServiceID: i.ServiceID,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			ServiceID: i.ServiceID,

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
