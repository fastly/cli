package kvstore_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/kvstore"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestCreateStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)
	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(kvstore.RootName + " create"),
			WantError: "error parsing arguments: required flag --name not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --name %s", kvstore.RootName, storeName)),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --name %s", kvstore.RootName, storeName)),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						ID:   storeID,
						Name: i.Name,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Created KV Store '%s' (%s)", storeName, storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --name %s --json", kvstore.RootName, storeName)),
			API: mock.API{
				CreateKVStoreFn: func(i *fastly.CreateKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.KVStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDeleteStoreCommand(t *testing.T) {
	const storeID = "test123"
	errStoreNotFound := errors.New("store not found")

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(kvstore.RootName + " delete"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(kvstore.RootName + " delete --store-id DOES-NOT-EXIST"),
			API: mock.API{
				DeleteKVStoreFn: func(i *fastly.DeleteKVStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantError: errStoreNotFound.Error(),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s", kvstore.RootName, storeID)),
			API: mock.API{
				DeleteKVStoreFn: func(i *fastly.DeleteKVStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted KV Store '%s'\n", storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --json", kvstore.RootName, storeID)),
			API: mock.API{
				DeleteKVStoreFn: func(i *fastly.DeleteKVStoreInput) error {
					if i.ID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "deleted": true}`, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDescribeStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(kvstore.RootName + " get"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s", kvstore.RootName, storeID)),
			API: mock.API{
				GetKVStoreFn: func(i *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s", kvstore.RootName, storeID)),
			API: mock.API{
				GetKVStoreFn: func(i *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fmtStore(
				&fastly.KVStore{
					ID:        storeID,
					Name:      storeName,
					CreatedAt: &now,
					UpdatedAt: &now,
				},
			),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s get --store-id %s --json", kvstore.RootName, storeID)),
			API: mock.API{
				GetKVStoreFn: func(i *fastly.GetKVStoreInput) (*fastly.KVStore, error) {
					return &fastly.KVStore{
						ID:        i.ID,
						Name:      storeName,
						CreatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.KVStore{
				ID:        storeID,
				Name:      storeName,
				CreatedAt: &now,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestListStoresCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	now := time.Now()

	stores := &fastly.ListKVStoresResponse{
		Data: []fastly.KVStore{
			{ID: storeID, Name: storeName, CreatedAt: &now, UpdatedAt: &now},
			{ID: storeID + "+1", Name: storeName + "+1", CreatedAt: &now, UpdatedAt: &now},
		},
	}

	scenarios := []testutil.TestScenario{
		{
			Args: testutil.Args(kvstore.RootName + " list"),
			API: mock.API{
				ListKVStoresFn: func(i *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return nil, nil
				},
			},
			WantOutput: "",
		},
		{
			Args: testutil.Args(kvstore.RootName + " list"),
			API: mock.API{
				ListKVStoresFn: func(i *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return nil, errors.New("unknown error")
				},
			},
			WantError: "unknown error",
		},
		{
			Args: testutil.Args(kvstore.RootName + " list"),
			API: mock.API{
				ListKVStoresFn: func(i *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return stores, nil
				},
			},
			WantOutput: fmtStores(stores),
		},
		{
			Args: testutil.Args(kvstore.RootName + " list --json"),
			API: mock.API{
				ListKVStoresFn: func(i *fastly.ListKVStoresInput) (*fastly.ListKVStoresResponse, error) {
					return stores, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(stores),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
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
