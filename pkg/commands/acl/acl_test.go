package acl_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/acl"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestACLCreate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foo",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foo --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foo --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foo --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate CreateACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateACLFn: func(i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--name foo --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CreateACLFn: func(i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:        "--name foo --service-id 123 --version 3",
			WantOutput: "Created ACL 'foo' (id: 456, service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateACLFn: func(i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:        "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Created ACL 'foo' (id: 456, service: 123, version: 4)",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestACLDelete(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foobar --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foo --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate DeleteACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteACLFn: func(i *fastly.DeleteACLInput) error {
					return testutil.Err
				},
			},
			Arg:       "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				DeleteACLFn: func(i *fastly.DeleteACLInput) error {
					return nil
				},
			},
			Arg:        "--name foobar --service-id 123 --version 3",
			WantOutput: "Deleted ACL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteACLFn: func(i *fastly.DeleteACLInput) error {
					return nil
				},
			},
			Arg:        "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Deleted ACL 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestACLDescribe(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foobar --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetACLFn: func(i *fastly.GetACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetACLFn:       getACL,
			},
			Arg:        "--name foobar --service-id 123 --version 3",
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetACLFn:       getACL,
			},
			Arg:        "--name foobar --service-id 123 --version 1",
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestACLList(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListACLs API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn: func(i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListACLs API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn:     listACLs,
			},
			Arg:        "--service-id 123 --version 3",
			WantOutput: "SERVICE ID  VERSION  NAME  ID\n123         3        foo   456\n123         3        bar   789\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn:     listACLs,
			},
			Arg:        "--service-id 123 --version 1",
			WantOutput: "SERVICE ID  VERSION  NAME  ID\n123         1        foo   456\n123         1        bar   789\n",
		},
		{
			Name: "validate --verbose flag",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				ListACLsFn:     listACLs,
			},
			Arg:        "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nID: 789\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestACLUpdate(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing --name flag",
			Arg:       "--new-name beepboop --version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --new-name flag",
			Arg:       "--name foobar --version 3",
			WantError: "error parsing arguments: required flag --new-name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Arg:       "--name foobar --new-name beepboop",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Arg:       "--name foobar --new-name beepboop --version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate missing --autoclone flag with 'active' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foo --new-name beepboop --service-id 123 --version 1",
			WantError: "service version 1 is active",
		},
		{
			Name: "validate missing --autoclone flag with 'locked' service",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
			},
			Arg:       "--name foo --new-name beepboop --service-id 123 --version 2",
			WantError: "service version 2 is locked",
		},
		{
			Name: "validate UpdateACL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateACLFn: func(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Arg:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateACL API success",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				UpdateACLFn: func(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:        "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateACLFn: func(i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Arg:        "--autoclone --name foobar --new-name beepboop --service-id 123 --version 1",
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 4)",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func getACL(i *fastly.GetACLInput) (*fastly.ACL, error) {
	t := testutil.Date

	return &fastly.ACL{
		ACLID:          fastly.ToPointer("456"),
		Name:           fastly.ToPointer(i.Name),
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),

		CreatedAt: &t,
		DeletedAt: &t,
		UpdatedAt: &t,
	}, nil
}

func listACLs(i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
	t := testutil.Date
	vs := []*fastly.ACL{
		{
			ACLID:          fastly.ToPointer("456"),
			Name:           fastly.ToPointer("foo"),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
		{
			ACLID:          fastly.ToPointer("789"),
			Name:           fastly.ToPointer("bar"),
			ServiceID:      fastly.ToPointer(i.ServiceID),
			ServiceVersion: fastly.ToPointer(i.ServiceVersion),

			CreatedAt: &t,
			DeletedAt: &t,
			UpdatedAt: &t,
		},
	}
	return vs, nil
}
