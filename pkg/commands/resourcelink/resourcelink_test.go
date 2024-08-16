package resourcelink_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/resourcelink"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateServiceResourceCommand(t *testing.T) {
	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		// Missing required arguments.
		{
			args:           "create --service-id abc --resource-id 123",
			wantError:      "error parsing arguments: required flag --version not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "create --service-id abc --version latest",
			wantError:      "error parsing arguments: required flag --resource-id not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "create --resource-id abc --version latest",
			wantError:      "error reading service: no service ID found",
			wantAPIInvoked: false,
		},
		// Success.
		{
			args: "create --resource-id abc --service-id 123 --version 42",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				CreateResourceFn: func(i *fastly.CreateResourceInput) (*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput:     `SUCCESS: Created service resource link "the-name" (rand-id) on service 123 version 42`,
		},
		// Success with --name.
		{
			args: "create --resource-id abc --service-id 123 --version 42 --name testing",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				CreateResourceFn: func(i *fastly.CreateResourceInput) (*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput:     `SUCCESS: Created service resource link "a-name" (rand-id) on service 123 version 42`,
		},
		// Success with --autoclone.
		{
			args: "create --resource-id abc --service-id 123 --version=latest --autoclone",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					// Specified version is active, meaning a service clone will be attempted.
					return []*fastly.Version{{Active: fastly.ToPointer(true), Number: fastly.ToPointer(42)}}, nil
				},
				CloneVersionFn: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
					return &fastly.Version{Number: fastly.ToPointer(43)}, nil
				},
				CreateResourceFn: func(i *fastly.CreateResourceInput) (*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput:     `SUCCESS: Created service resource link "cloned" (rand-id) on service 123 version 43`,
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(resourcelink.RootName + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.CreateResourceFn
			var apiInvoked bool
			testcase.api.CreateResourceFn = func(i *fastly.CreateResourceInput) (*fastly.Resource, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, strings.TrimSpace(stdout.String()))
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API CreateResource invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDeleteServiceResourceCommand(t *testing.T) {
	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		// Missing required arguments.
		{
			args:           "delete --id LINK-ID --service-id abc",
			wantError:      "error parsing arguments: required flag --version not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "delete --id LINK-ID --version 123",
			wantError:      "error reading service: no service ID found",
			wantAPIInvoked: false,
		},
		{
			args:           "delete --service-id abc --version 123",
			wantError:      "error parsing arguments: required flag --id not provided",
			wantAPIInvoked: false,
		},
		// Success.
		{
			args: "delete --service-id 123 --version 42 --id LINKID",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				DeleteResourceFn: func(i *fastly.DeleteResourceInput) error {
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
			wantAPIInvoked: true,
			wantOutput:     "SUCCESS: Deleted service resource link LINKID from service 123 version 42",
		},
		// Success with --autoclone.
		{
			args: "delete --service-id 123 --version 42 --id LINKID --autoclone",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					// Specified version is active, meaning a service clone will be attempted.
					return []*fastly.Version{{Active: fastly.ToPointer(true), Number: fastly.ToPointer(42)}}, nil
				},
				CloneVersionFn: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
					return &fastly.Version{Number: fastly.ToPointer(43)}, nil
				},
				DeleteResourceFn: func(i *fastly.DeleteResourceInput) error {
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
			wantAPIInvoked: true,
			wantOutput:     "SUCCESS: Deleted service resource link LINKID from service 123 version 43",
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(resourcelink.RootName + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.DeleteResourceFn
			var apiInvoked bool
			testcase.api.DeleteResourceFn = func(i *fastly.DeleteResourceInput) error {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, strings.TrimSpace(stdout.String()))
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DeleteResource invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDescribeServiceResourceCommand(t *testing.T) {
	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		// Missing required arguments.
		{
			args:           "describe --id LINK-ID --service-id abc",
			wantError:      "error parsing arguments: required flag --version not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "describe --id LINK-ID --version 123",
			wantError:      "error reading service: no service ID found",
			wantAPIInvoked: false,
		},
		{
			args:           "describe --service-id abc --version 123",
			wantError:      "error parsing arguments: required flag --id not provided",
			wantAPIInvoked: false,
		},
		// Success.
		{
			args: "describe --service-id 123 --version 42 --id LINKID",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				GetResourceFn: func(i *fastly.GetResourceInput) (*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput: `Service ID: 123
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

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(resourcelink.RootName + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.GetResourceFn
			var apiInvoked bool
			testcase.api.GetResourceFn = func(i *fastly.GetResourceInput) (*fastly.Resource, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, strings.TrimSpace(stdout.String()))
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DescribeResource invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestListServiceResourceCommand(t *testing.T) {
	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		// Missing required arguments.
		{
			args:           "list --service-id abc",
			wantError:      "error parsing arguments: required flag --version not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "list --version 123",
			wantError:      "error reading service: no service ID found",
			wantAPIInvoked: false,
		},
		// Success.
		{
			args: "list --service-id 123 --version 42",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				ListResourcesFn: func(i *fastly.ListResourcesInput) ([]*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput: `Service ID: 123
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

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(resourcelink.RootName + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.ListResourcesFn
			var apiInvoked bool
			testcase.api.ListResourcesFn = func(i *fastly.ListResourcesInput) ([]*fastly.Resource, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, strings.ReplaceAll(strings.TrimSpace(stdout.String()), "\t", "  "))
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListResources invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestUpdateServiceResourceCommand(t *testing.T) {
	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		// Missing required arguments.
		{
			args:           "update --id LINK-ID --name new-name --service-id abc",
			wantError:      "error parsing arguments: required flag --version not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "update --id LINK-ID --name new-name --version 123",
			wantError:      "error reading service: no service ID found",
			wantAPIInvoked: false,
		},
		{
			args:           "update --id LINK-ID --service-id abc --version 123",
			wantError:      "error parsing arguments: required flag --name not provided",
			wantAPIInvoked: false,
		},
		{
			args:           "update --name new-name --service-id abc --version 123",
			wantError:      "error parsing arguments: required flag --id not provided",
			wantAPIInvoked: false,
		},
		// Success.
		{
			args: "update --id LINK-ID --name new-name --service-id 123 --version 42",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					return []*fastly.Version{{Number: fastly.ToPointer(42)}}, nil
				},
				UpdateResourceFn: func(i *fastly.UpdateResourceInput) (*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput:     "SUCCESS: Updated service resource link LINK-ID on service 123 version 42",
		},
		// Success with --autoclone.
		{
			args: "update --id LINK-ID --name new-name --service-id 123 --version 42 --autoclone",
			api: mock.API{
				ListVersionsFn: func(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
					// Specified version is active, meaning a service clone will be attempted.
					return []*fastly.Version{{Active: fastly.ToPointer(true), Number: fastly.ToPointer(42)}}, nil
				},
				CloneVersionFn: func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
					return &fastly.Version{Number: fastly.ToPointer(43)}, nil
				},
				UpdateResourceFn: func(i *fastly.UpdateResourceInput) (*fastly.Resource, error) {
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
			wantAPIInvoked: true,
			wantOutput:     "SUCCESS: Updated service resource link LINK-ID on service 123 version 43",
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(resourcelink.RootName + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.UpdateResourceFn
			var apiInvoked bool
			testcase.api.UpdateResourceFn = func(i *fastly.UpdateResourceInput) (*fastly.Resource, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, strings.TrimSpace(stdout.String()))
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API UpdateResource invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}
