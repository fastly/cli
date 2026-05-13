package acl_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/acl"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestACLCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foo",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foo --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate CreateACL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateACLFn: func(_ context.Context, _ *fastly.CreateACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foo --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate CreateACL API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				CreateACLFn: func(_ context.Context, i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--name foo --service-id 123 --version 3",
			WantOutput: "Created ACL 'foo' (id: 456, service: 123, version: 3)",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				CreateACLFn: func(_ context.Context, i *fastly.CreateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.Name,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Created ACL 'foo' (id: 456, service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestACLDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate DeleteACL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteACLFn: func(_ context.Context, _ *fastly.DeleteACLInput) error {
					return testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DeleteACL API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteACLFn: func(_ context.Context, _ *fastly.DeleteACLInput) error {
					return nil
				},
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "Deleted ACL 'foobar' (service: 123, version: 3)",
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteACLFn: func(_ context.Context, i *fastly.DeleteACLInput) error {
					return fmt.Errorf("Cannot update version %d. Versions that have been activated cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been activated cannot be updated",
		},
		{
			Name: "validate API error when modifying locked version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				DeleteACLFn: func(_ context.Context, i *fastly.DeleteACLInput) error {
					return fmt.Errorf("Cannot update version %d. Versions that have been locked cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been locked cannot be updated",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteACLFn: func(_ context.Context, _ *fastly.DeleteACLInput) error {
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 1",
			WantOutput: "Deleted ACL 'foo' (service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteACLFn: func(_ context.Context, i *fastly.DeleteACLInput) error {
					// Verify operation happens on the cloned version (4), not original (2)
					if i.ServiceVersion != 4 {
						return fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 2",
			WantOutput: "Deleted ACL 'foo' (service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				DeleteACLFn: func(_ context.Context, i *fastly.DeleteACLInput) error {
					// Verify operation happens on the cloned version (4), not original (3)
					if i.ServiceVersion != 4 {
						return fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return nil
				},
			},
			Args:       "--autoclone --name foo --service-id 123 --version 3",
			WantOutput: "Deleted ACL 'foo' (service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestACLDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate GetACL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetACLFn: func(_ context.Context, _ *fastly.GetACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate GetACL API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetACLFn:     getACL,
			},
			Args:       "--name foobar --service-id 123 --version 3",
			WantOutput: "\nService ID: 123\nService Version: 3\n\nName: foobar\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				GetACLFn:     getACL,
			},
			Args:       "--name foobar --service-id 123 --version 1",
			WantOutput: "\nService ID: 123\nService Version: 1\n\nName: foobar\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestACLList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate ListACLs API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListACLsFn: func(_ context.Context, _ *fastly.ListACLsInput) ([]*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate ListACLs API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListACLsFn:   listACLs,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "SERVICE ID  VERSION  NAME  ID\n123         3        foo   456\n123         3        bar   789\n",
		},
		{
			Name: "validate missing --autoclone flag is OK",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListACLsFn:   listACLs,
			},
			Args:       "--service-id 123 --version 1",
			WantOutput: "SERVICE ID  VERSION  NAME  ID\n123         1        foo   456\n123         1        bar   789\n",
		},
		{
			Name: "validate --verbose flag",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				ListACLsFn:   listACLs,
			},
			Args:       "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (auth: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nID: 456\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\nName: bar\nID: 789\n\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\n\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestACLUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --name flag",
			Args:      "--new-name beepboop --version 3",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Name:      "validate missing --new-name flag",
			Args:      "--name foobar --version 3",
			WantError: "error parsing arguments: required flag --new-name not provided",
		},
		{
			Name:      "validate missing --version flag",
			Args:      "--name foobar --new-name beepboop",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--name foobar --new-name beepboop --version 3",
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate UpdateACL API error",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateACLFn: func(_ context.Context, _ *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate UpdateACL API success",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateACLFn: func(_ context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 3)",
		},
		{
			Name: "validate API error when modifying active version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateACLFn: func(_ context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return nil, fmt.Errorf("Cannot update version %d. Versions that have been activated cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been activated cannot be updated",
		},
		{
			Name: "validate API error when modifying locked version",
			API: &mock.API{
				GetVersionFn: testutil.GetVersion,
				UpdateACLFn: func(_ context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return nil, fmt.Errorf("Cannot update version %d. Versions that have been locked cannot be updated", i.ServiceVersion)
				},
			},
			Args:      "--name foobar --new-name beepboop --service-id 123 --version 3",
			WantError: "Cannot update version 3. Versions that have been locked cannot be updated",
		},
		{
			Name: "validate --autoclone results in cloned service version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateACLFn: func(_ context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 1",
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on locked version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateACLFn: func(_ context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					// Verify operation happens on the cloned version (4), not original (2)
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 2",
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 4)",
		},
		{
			Name: "validate --autoclone on editable version",
			API: &mock.API{
				GetVersionFn:   testutil.GetVersion,
				CloneVersionFn: testutil.CloneVersionResult(4),
				UpdateACLFn: func(_ context.Context, i *fastly.UpdateACLInput) (*fastly.ACL, error) {
					// Verify operation happens on the cloned version (4), not original (3)
					if i.ServiceVersion != 4 {
						return nil, fmt.Errorf("expected operation on cloned version 4, got %d", i.ServiceVersion)
					}
					return &fastly.ACL{
						ACLID:          fastly.ToPointer("456"),
						Name:           i.NewName,
						ServiceID:      fastly.ToPointer(i.ServiceID),
						ServiceVersion: fastly.ToPointer(i.ServiceVersion),
					}, nil
				},
			},
			Args:       "--autoclone --name foobar --new-name beepboop --service-id 123 --version 3",
			WantOutput: "Updated ACL 'beepboop' (previously: 'foobar', service: 123, version: 4)",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}

func getACL(_ context.Context, i *fastly.GetACLInput) (*fastly.ACL, error) {
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

func listACLs(_ context.Context, i *fastly.ListACLsInput) ([]*fastly.ACL, error) {
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
