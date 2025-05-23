package configstore_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/configstore"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)
	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: fmt.Sprintf("--name %s", storeName),
			API: mock.API{
				CreateConfigStoreFn: func(_ *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--name %s", storeName),
			API: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID: storeID,
						Name:    i.Name,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Created Config Store '%s' (%s)", storeName, storeID),
		},
		{
			Args: fmt.Sprintf("--name %s --json", storeName),
			API: mock.API{
				CreateConfigStoreFn: func(i *fastly.CreateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID:   storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				StoreID:   storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDeleteStoreCommand(t *testing.T) {
	const storeID = "test123"
	errStoreNotFound := errors.New("store not found")

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: "--store-id DOES-NOT-EXIST",
			API: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.StoreID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantError: errStoreNotFound.Error(),
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.StoreID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted Config Store '%s'\n", storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				DeleteConfigStoreFn: func(i *fastly.DeleteConfigStoreInput) error {
					if i.StoreID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, storeID),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestGetStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				GetConfigStoreFn: func(_ *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID:   i.StoreID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fmtStore(
				&fastly.ConfigStore{
					StoreID:   storeID,
					Name:      storeName,
					CreatedAt: &now,
				},
				nil,
			),
		},
		{
			Args: fmt.Sprintf("--store-id %s --metadata", storeID),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID:   i.StoreID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
				GetConfigStoreMetadataFn: func(_ *fastly.GetConfigStoreMetadataInput) (*fastly.ConfigStoreMetadata, error) {
					return &fastly.ConfigStoreMetadata{
						ItemCount: 42,
					}, nil
				},
			},
			WantOutput: fmtStore(
				&fastly.ConfigStore{
					StoreID:   storeID,
					Name:      storeName,
					CreatedAt: &now,
				},
				&fastly.ConfigStoreMetadata{
					ItemCount: 42,
				},
			),
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				GetConfigStoreFn: func(i *fastly.GetConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID:   i.StoreID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				StoreID:   storeID,
				Name:      storeName,
				CreatedAt: &now,
			}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "get"}, scenarios)
}

func TestListStoresCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	stores := []*fastly.ConfigStore{
		{StoreID: storeID, Name: storeName, CreatedAt: &now},
		{StoreID: storeID + "+1", Name: storeName + "+1", CreatedAt: &now},
	}

	scenarios := []testutil.CLIScenario{
		{
			API: mock.API{
				ListConfigStoresFn: func(_ *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
					return nil, nil
				},
			},
			WantOutput: fmtStores(nil),
		},
		{
			API: mock.API{
				ListConfigStoresFn: func(_ *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			API: mock.API{
				ListConfigStoresFn: func(_ *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
					return stores, nil
				},
			},
			WantOutput: fmtStores(stores),
		},
		{
			Args: "--json",
			API: mock.API{
				ListConfigStoresFn: func(_ *fastly.ListConfigStoresInput) ([]*fastly.ConfigStore, error) {
					return stores, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(stores),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestListStoreServicesCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	services := []*fastly.Service{
		{ServiceID: fastly.ToPointer("abc1"), Name: fastly.ToPointer("test1"), Type: fastly.ToPointer("wasm")},
		{ServiceID: fastly.ToPointer("abc2"), Name: fastly.ToPointer("test2"), Type: fastly.ToPointer("vcl")},
	}

	scenarios := []testutil.CLIScenario{
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListConfigStoreServicesFn: func(_ *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return nil, nil
				},
			},
			WantOutput: fmtServices(nil),
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListConfigStoreServicesFn: func(_ *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListConfigStoreServicesFn: func(_ *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return services, nil
				},
			},
			WantOutput: fmtServices(services),
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				ListConfigStoreServicesFn: func(_ *fastly.ListConfigStoreServicesInput) ([]*fastly.Service, error) {
					return services, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(services),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list-services"}, scenarios)
}

func TestUpdateStoreCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		storeName = "test123"
	)
	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			Args:      fmt.Sprintf("--store-id %s", storeID),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s --name %s", storeID, storeName),
			API: mock.API{
				UpdateConfigStoreFn: func(_ *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --name %s", storeID, storeName),
			API: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID:   storeID,
						Name:      i.Name,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Updated Config Store '%s' (%s)", storeName, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --name %s --json", storeID, storeName),
			API: mock.API{
				UpdateConfigStoreFn: func(i *fastly.UpdateConfigStoreInput) (*fastly.ConfigStore, error) {
					return &fastly.ConfigStore{
						StoreID:   storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStore{
				StoreID:   storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}
