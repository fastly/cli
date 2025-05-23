package kvstore_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/kvstore"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestCreateStoreCommand(t *testing.T) {
	const (
		storeName     = "test123"
		storeLocation = "EU"
		storeID       = "store-id-123"
	)
	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: fmt.Sprintf("--name %s", storeName),
			API: mock.API{
				CreateKVStoreFn: func(_ *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--name %s", storeName),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						StoreID: storeID,
						Name:    i.Name,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Created KV Store '%s' (%s)", storeName, storeID),
		},
		{
			Args: fmt.Sprintf("--name %s --json", storeName),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						StoreID:   storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.KVStore{
				StoreID:   storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
		{
			// NOTE: The following tests only validate support for the --location flag.
			// Location/region indicators are not exposed for us to validate.
			Args: fmt.Sprintf("--name %s --location %s", storeName, storeLocation),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						StoreID: storeID,
						Name:    i.Name,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Created KV Store '%s' (%s)", storeName, storeID),
		},
		{
			// NOTE: The following tests only validate support for the --location flag.
			// Location/region indicators are not exposed for us to validate.
			Args: fmt.Sprintf("--name %s --location %s --json", storeName, storeLocation),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						StoreID:   storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.KVStore{
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
				DeleteKVStoreFn: func(i *fastly.DeleteKVStoreInput) error {
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
				DeleteKVStoreFn: func(i *fastly.DeleteKVStoreInput) error {
					if i.StoreID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted KV Store '%s'\n", storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				DeleteKVStoreFn: func(i *fastly.DeleteKVStoreInput) error {
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
				GetKVStoreFn: func(_ *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				GetKVStoreFn: func(i *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						StoreID:   i.StoreID,
						Name:      storeName,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fmtStore(
				&fastly.KVStore{
					StoreID:   storeID,
					Name:      storeName,
					CreatedAt: &now,
					UpdatedAt: &now,
				},
			),
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				GetKVStoreFn: func(i *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						StoreID:   i.StoreID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.KVStore{
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

	stores := &fastly.ListKVStoresResponse{
		Data: []fastly.KVStore{
			{StoreID: storeID, Name: storeName, CreatedAt: &now, UpdatedAt: &now},
			{StoreID: storeID + "+1", Name: storeName + "+1", CreatedAt: &now, UpdatedAt: &now},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			API: mock.API{
				ListKVStoresFn: func(_ *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return nil, nil
				},
			},
		},
		{
			API: mock.API{
				ListKVStoresFn: func(_ *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			API: mock.API{
				ListKVStoresFn: func(_ *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return stores, nil
				},
			},
			WantOutput: fmtStores(stores),
		},
		{
			Args: "--json",
			API: mock.API{
				ListKVStoresFn: func(_ *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return stores, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(stores),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func fmtStore(ks *fastly.KVStore) string {
	var b bytes.Buffer
	text.PrintKVStore(&b, "", ks)
	return b.String()
}

func fmtStores(ks *fastly.ListKVStoresResponse) string {
	var b bytes.Buffer
	for _, o := range ks.Data {
		// avoid gosec loop aliasing check :/
		o := o
		text.PrintKVStore(&b, "", &o)
	}
	return b.String()
}
