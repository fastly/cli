package resourcelink_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/resourcelink"
	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateServiceResourceCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		// Missing required arguments.
		{
			Args:      "--service-id abc --resource-id 123",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args:      "--service-id abc --version latest",
			WantError: "error parsing arguments: required flag --resource-id not provided",
		},
		{
			Args:      "--resource-id abc --version latest",
			WantError: "error reading service: no service ID found",
		},
		// Success.
		{
			Args: "--resource-id abc --service-id 123 --version 42",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				CreateResourceFn: func(_ context.Context, i *fastly.CreateResourceInput) (*fastly.Resource, error) {
					if got, want := *i.ResourceID, "abc"; got != want {
						return nil, fmt.Errorf("ResourceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := *i.Name, ""; got != want {
						return nil, fmt.Errorf("Name: got %q, want %q", got, want)
					}
					now := time.Now()
					return &fastly.Resource{
						LinkID:         fastly.ToPointer("rand-id"),
						Name:           fastly.ToPointer("the-name"),
						ResourceID:     fastly.ToPointer("abc"),
						ServiceID:      fastly.ToPointer("123"),
						ServiceVersion: fastly.ToPointer(42),
						CreatedAt:      &now,
						UpdatedAt:      &now,
					}, nil
				},
			},
			WantOutput: `SUCCESS: Created service resource link "the-name" (rand-id) on service 123 version 42`,
		},
		// Success with --name.
		{
			Args: "--resource-id abc --service-id 123 --version 42 --name testing",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				CreateResourceFn: func(_ context.Context, i *fastly.CreateResourceInput) (*fastly.Resource, error) {
					if got, want := *i.ResourceID, "abc"; got != want {
						return nil, fmt.Errorf("ResourceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := *i.Name, "testing"; got != want {
						return nil, fmt.Errorf("Name: got %q, want %q", got, want)
					}
					now := time.Now()
					return &fastly.Resource{
						LinkID:         fastly.ToPointer("rand-id"),
						Name:           fastly.ToPointer("a-name"),
						ResourceID:     fastly.ToPointer("abc"),
						ServiceID:      fastly.ToPointer("123"),
						ServiceVersion: fastly.ToPointer(42),
						CreatedAt:      &now,
						UpdatedAt:      &now,
					}, nil
				},
			},
			WantOutput: `SUCCESS: Created service resource link "a-name" (rand-id) on service 123 version 42`,
		},
		// Success with --autoclone.
		{
			Args: "--resource-id abc --service-id 123 --version=latest --autoclone",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					// Specified version is active, meaning a service clone will be attempted.
					return []*fastly.Version{{Active: fastly.ToPointer(true), Number: fastly.ToPointer(42)}}, nil
				},
				CloneVersionFn: func(_ context.Context, _ *fastly.CloneVersionInput) (*fastly.Version, error) {
					return &fastly.Version{Number: fastly.ToPointer(43)}, nil
				},
				CreateResourceFn: func(_ context.Context, i *fastly.CreateResourceInput) (*fastly.Resource, error) {
					if got, want := *i.ResourceID, "abc"; got != want {
						return nil, fmt.Errorf("ResourceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := *i.Name, ""; got != want {
						return nil, fmt.Errorf("Name: got %q, want %q", got, want)
					}
					now := time.Now()
					return &fastly.Resource{
						LinkID:         fastly.ToPointer("rand-id"),
						Name:           fastly.ToPointer("cloned"),
						ResourceID:     fastly.ToPointer("abc"),
						ServiceID:      fastly.ToPointer("123"),
						ServiceVersion: fastly.ToPointer(43), // Cloned version.
						CreatedAt:      &now,
						UpdatedAt:      &now,
					}, nil
				},
			},
			WantOutput: `SUCCESS: Created service resource link "cloned" (rand-id) on service 123 version 43`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestDeleteServiceResourceCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		// Missing required arguments.
		{
			Args:      "--id LINK-ID --service-id abc",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args:      "--id LINK-ID --version 123",
			WantError: "error reading service: no service ID found",
		},
		{
			Args:      "--service-id abc --version 123",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		// Success.
		{
			Args: "--service-id 123 --version 42 --id LINKID",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				DeleteResourceFn: func(_ context.Context, i *fastly.DeleteResourceInput) error {
					if got, want := i.ResourceID, "LINKID"; got != want {
						return fmt.Errorf("ID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceVersion, 42; got != want {
						return fmt.Errorf("ServiceVersion: got %d, want %d", got, want)
					}
					return nil
				},
			},
			WantOutput: "SUCCESS: Deleted service resource link LINKID from service 123 version 42",
		},
		// Success with --autoclone.
		{
			Args: "--service-id 123 --version 42 --id LINKID --autoclone",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					// Specified version is active, meaning a service clone will be attempted.
					return []*fastly.Version{{Active: fastly.ToPointer(true), Number: fastly.ToPointer(42)}}, nil
				},
				CloneVersionFn: func(_ context.Context, _ *fastly.CloneVersionInput) (*fastly.Version, error) {
					return &fastly.Version{Number: fastly.ToPointer(43)}, nil
				},
				DeleteResourceFn: func(_ context.Context, i *fastly.DeleteResourceInput) error {
					if got, want := i.ResourceID, "LINKID"; got != want {
						return fmt.Errorf("ID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceVersion, 43; got != want {
						return fmt.Errorf("ServiceVersion: got %d, want %d", got, want)
					}
					return nil
				},
			},
			WantOutput: "SUCCESS: Deleted service resource link LINKID from service 123 version 43",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func TestDescribeServiceResourceCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		// Missing required arguments.
		{
			Args:      "--id LINK-ID --service-id abc",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args:      "--id LINK-ID --version 123",
			WantError: "error reading service: no service ID found",
		},
		{
			Args:      "--service-id abc --version 123",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		// Success.
		{
			Args: "--service-id 123 --version 42 --id LINKID",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				GetResourceFn: func(_ context.Context, i *fastly.GetResourceInput) (*fastly.Resource, error) {
					if got, want := i.ResourceID, "LINKID"; got != want {
						return nil, fmt.Errorf("ID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceVersion, 42; got != want {
						return nil, fmt.Errorf("ServiceVersion: got %d, want %d", got, want)
					}
					now := time.Unix(1697372322, 0)
					return &fastly.Resource{
						LinkID:         fastly.ToPointer("LINKID"),
						ResourceID:     fastly.ToPointer("abc"),
						ResourceType:   fastly.ToPointer("secret-store"),
						Name:           fastly.ToPointer("test-name"),
						ServiceID:      fastly.ToPointer("123"),
						ServiceVersion: fastly.ToPointer(42),
						CreatedAt:      &now,
						UpdatedAt:      &now,
					}, nil
				},
			},
			WantOutput: `Service ID: 123
Service Version: 42
ID: LINKID
Name: test-name
Service ID: 123
Service Version: 42
Resource ID: abc
Resource Type: secret-store
Created (UTC): 2023-10-15 12:18
Last edited (UTC): 2023-10-15 12:18`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestListServiceResourceCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		// Missing required arguments.
		{
			Args:      "--service-id abc",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args:      "--version 123",
			WantError: "error reading service: no service ID found",
		},
		// Success.
		{
			Args: "--service-id 123 --version 42",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				ListResourcesFn: func(_ context.Context, i *fastly.ListResourcesInput) ([]*fastly.Resource, error) {
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceVersion, 42; got != want {
						return nil, fmt.Errorf("ServiceVersion: got %d, want %d", got, want)
					}

					now := time.Unix(1697372322, 0)
					resources := make([]*fastly.Resource, 3)
					for i := range resources {
						resources[i] = &fastly.Resource{
							LinkID:         fastly.ToPointer(fmt.Sprintf("LINKID-%02d", i)),
							ResourceID:     fastly.ToPointer("abc"),
							ResourceType:   fastly.ToPointer("secret-store"),
							Name:           fastly.ToPointer("test-name"),
							ServiceID:      fastly.ToPointer("123"),
							ServiceVersion: fastly.ToPointer(42),
							CreatedAt:      &now,
							UpdatedAt:      &now,
						}
					}
					return resources, nil
				},
			},
			WantOutput: `Service ID: 123
Service Version: 42
Resource Link 1/3
  ID: LINKID-00
  Name: test-name
  Service ID: 123
  Service Version: 42
  Resource ID: abc
  Resource Type: secret-store
  Created (UTC): 2023-10-15 12:18
  Last edited (UTC): 2023-10-15 12:18

Resource Link 2/3
  ID: LINKID-01
  Name: test-name
  Service ID: 123
  Service Version: 42
  Resource ID: abc
  Resource Type: secret-store
  Created (UTC): 2023-10-15 12:18
  Last edited (UTC): 2023-10-15 12:18

Resource Link 3/3
  ID: LINKID-02
  Name: test-name
  Service ID: 123
  Service Version: 42
  Resource ID: abc
  Resource Type: secret-store
  Created (UTC): 2023-10-15 12:18
  Last edited (UTC): 2023-10-15 12:18`,
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestUpdateServiceResourceCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		// Missing required arguments.
		{
			Args:      "--id LINK-ID --name new-name --service-id abc",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Args:      "--id LINK-ID --name new-name --version 123",
			WantError: "error reading service: no service ID found",
		},
		{
			Args:      "--id LINK-ID --service-id abc --version 123",
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args:      "--name new-name --service-id abc --version 123",
			WantError: "error parsing arguments: required flag --id not provided",
		},
		// Success.
		{
			Args: "--id LINK-ID --name new-name --service-id 123 --version 42",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				UpdateResourceFn: func(_ context.Context, i *fastly.UpdateResourceInput) (*fastly.Resource, error) {
					if got, want := i.ResourceID, "LINK-ID"; got != want {
						return nil, fmt.Errorf("ID: got %q, want %q", got, want)
					}
					if got, want := *i.Name, "new-name"; got != want {
						return nil, fmt.Errorf("Name: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceVersion, 42; got != want {
						return nil, fmt.Errorf("ServiceVersion: got %d, want %d", got, want)
					}

					now := time.Now()
					return &fastly.Resource{
						LinkID:         fastly.ToPointer("LINK-ID"),
						ResourceID:     fastly.ToPointer("abc"),
						ResourceType:   fastly.ToPointer("secret-store"),
						Name:           fastly.ToPointer("new-name"),
						ServiceID:      fastly.ToPointer("123"),
						ServiceVersion: fastly.ToPointer(42),
						CreatedAt:      &now,
						UpdatedAt:      &now,
					}, nil
				},
			},
			WantOutput: "SUCCESS: Updated service resource link LINK-ID on service 123 version 42",
		},
		// Success with --autoclone.
		{
			Args: "--id LINK-ID --name new-name --service-id 123 --version 42 --autoclone",
			API: mock.API{
				ListVersionsFn: func(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					// Specified version is active, meaning a service clone will be attempted.
					return []*fastly.Version{{Active: fastly.ToPointer(true), Number: fastly.ToPointer(42)}}, nil
				},
				CloneVersionFn: func(_ context.Context, _ *fastly.CloneVersionInput) (*fastly.Version, error) {
					return &fastly.Version{Number: fastly.ToPointer(43)}, nil
				},
				UpdateResourceFn: func(_ context.Context, i *fastly.UpdateResourceInput) (*fastly.Resource, error) {
					if got, want := i.ResourceID, "LINK-ID"; got != want {
						return nil, fmt.Errorf("ID: got %q, want %q", got, want)
					}
					if got, want := *i.Name, "new-name"; got != want {
						return nil, fmt.Errorf("Name: got %q, want %q", got, want)
					}
					if got, want := i.ServiceID, "123"; got != want {
						return nil, fmt.Errorf("ServiceID: got %q, want %q", got, want)
					}
					if got, want := i.ServiceVersion, 43; got != want {
						return nil, fmt.Errorf("ServiceVersion: got %d, want %d", got, want)
					}

					now := time.Now()
					return &fastly.Resource{
						LinkID:         fastly.ToPointer("LINK-ID"),
						ResourceID:     fastly.ToPointer("abc"),
						ResourceType:   fastly.ToPointer("secret-store"),
						Name:           fastly.ToPointer("new-name"),
						ServiceID:      fastly.ToPointer("123"),
						ServiceVersion: fastly.ToPointer(43),
						CreatedAt:      &now,
						UpdatedAt:      &now,
					}, nil
				},
			},
			WantOutput: "SUCCESS: Updated service resource link LINK-ID on service 123 version 43",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
