package configstore_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/configstore"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestCreateStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)
	now := time.Now()

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "create",
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: fmt.Sprintf("create --name %s", storeName),
			api: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("create --name %s", storeName),
			api: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:   storeID,
						Name: i.Name,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Created config store %s (name %s)", storeID, storeName),
		},
		{
			args: fmt.Sprintf("create --name %s --json", storeName),
			api: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstore.RootName+" "+testcase.args), &stdout)

			f := testcase.api.CreateConfigStoreFn
			var apiInvoked bool
			testcase.api.CreateConfigStoreFn = func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API CreateConfigStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDeleteStoreCommand(t *testing.T) {
	const storeID = "test123"
	errStoreNotFound := errors.New("store not found")

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "delete",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: "delete --store-id DOES-NOT-EXIST",
			api: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantError:      errStoreNotFound.Error(),
		},
		{
			args: fmt.Sprintf("delete --store-id %s", storeID),
			api: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Deleted config store %s\n", storeID),
		},
		{
			args: fmt.Sprintf("delete --store-id %s --json", storeID),
			api: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.JSON(`{"id": %q, "deleted": true}`, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstore.RootName+" "+testcase.args), &stdout)

			f := testcase.api.DeleteConfigStoreFn
			var apiInvoked bool
			testcase.api.DeleteConfigStoreFn = func(i *fastly.DeleteConfigStoreInput) error {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DeleteConfigStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDescribeStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "get",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("get --store-id %s", storeID),
			api: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("get --store-id %s", storeID),
			api: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fmtStore(configstore.ConfigStoreWithMetadata{
				ConfigStore: &fastly.ConfigStore{
					ID:        storeID,
					Name:      storeName,
					CreatedAt: &now,
				},
			}),
		},
		{
			args: fmt.Sprintf("get --store-id %s --metadata", storeID),
			api: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
				GetConfigStoreMetadataFn: func(i *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error) {
					return &fastly.ConfigStoreMetadata{
						ItemCount: 42,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fmtStore(configstore.ConfigStoreWithMetadata{
				ConfigStore: &fastly.ConfigStore{
					ID:        storeID,
					Name:      storeName,
					CreatedAt: &now,
				},
				Metdata: &fastly.ConfigStoreMetadata{
					ItemCount: 42,
				},
			}),
		},
		{
			args: fmt.Sprintf("get --store-id %s --json", storeID),
			api: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstore.RootName+" "+testcase.args), &stdout)

			f := testcase.api.GetConfigStoreFn
			var apiInvoked bool
			testcase.api.GetConfigStoreFn = func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API GetConfigStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestListStoresCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	stores := []*fastly.ConfigStore{
		{ID: storeID, Name: storeName, CreatedAt: &now},
		{ID: storeID + "+1", Name: storeName + "+1", CreatedAt: &now},
	}

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args: "list",
			api: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return nil, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtStores(nil),
		},
		{
			args: "list",
			api: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return nil, errors.New("unknown error")
				},
			},
			wantAPIInvoked: true,
			wantError:      "unknown error",
		},
		{
			args: "list",
			api: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return stores, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtStores(stores),
		},
		{
			args: "list --json",
			api: mock.API{
				ListConfigStoresFn: func() ([]*fastly.ConfigStore, error) {
					return stores, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.EncodeJSON(stores),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstore.RootName+" "+testcase.args), &stdout)

			f := testcase.api.ListConfigStoresFn
			var apiInvoked bool
			testcase.api.ListConfigStoresFn = func() ([]*fastly.ConfigStore, error) {
				apiInvoked = true
				return f()
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListConfigStores invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestListStoreServicesCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	services := []*fastly.Service{
		{ID: "abc1", Name: "test1", Type: "wasm"},
		{ID: "abc2", Name: "test2", Type: "vcl"},
	}

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args: fmt.Sprintf("list-services --store-id %s", storeID),
			api: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return nil, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtServices(nil),
		},
		{
			args: fmt.Sprintf("list-services --store-id %s", storeID),
			api: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return nil, errors.New("unknown error")
				},
			},
			wantAPIInvoked: true,
			wantError:      "unknown error",
		},
		{
			args: fmt.Sprintf("list-services --store-id %s", storeID),
			api: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return services, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtServices(services),
		},
		{
			args: fmt.Sprintf("list-services --store-id %s --json", storeID),
			api: mock.API{
				ListConfigStoreServicesFn: func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return services, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.EncodeJSON(services),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstore.RootName+" "+testcase.args), &stdout)

			f := testcase.api.ListConfigStoreServicesFn
			var apiInvoked bool
			testcase.api.ListConfigStoreServicesFn = func(i *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListConfigStoreServices invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestUpdateStoreCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		storeName = "test123"
	)
	now := time.Now()

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      fmt.Sprintf("update --store-id %s", storeID),
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args: fmt.Sprintf("update --store-id %s --name %s", storeID, storeName),
			api: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("update --store-id %s --name %s", storeID, storeName),
			api: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Updated config store %s (name %s)", storeID, storeName),
		},
		{
			args: fmt.Sprintf("update --store-id %s --name %s --json", storeID, storeName),
			api: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstore.RootName+" "+testcase.args), &stdout)

			f := testcase.api.UpdateConfigStoreFn
			var apiInvoked bool
			testcase.api.UpdateConfigStoreFn = func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API UpdateConfigStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}
