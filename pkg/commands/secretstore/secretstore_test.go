package secretstore_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/secretstore"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
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
				CreateSecretStoreFn: func(i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("create --name %s", storeName),
			api: mock.API{
				CreateSecretStoreFn: func(i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
					return &fastly.SecretStore{
						StoreID: storeID,
						Name:    i.Name,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Created Secret Store '%s' (%s)", storeName, storeID),
		},
		{
			args: fmt.Sprintf("create --name %s --json", storeName),
			api: mock.API{
				CreateSecretStoreFn: func(i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
					return &fastly.SecretStore{
						StoreID:   storeID,
						Name:      i.Name,
						CreatedAt: now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.JSON(`{"created_at": %q, "name": %q, "id": %q}`, now.Format(time.RFC3339Nano), storeName, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstore.RootNameStore + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.CreateSecretStoreFn
			var apiInvoked bool
			testcase.api.CreateSecretStoreFn = func(i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API CreateSecretStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
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
				DeleteSecretStoreFn: func(i *fastly.DeleteSecretStoreInput) error {
					if i.StoreID != storeID {
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
				DeleteSecretStoreFn: func(i *fastly.DeleteSecretStoreInput) error {
					if i.StoreID != storeID {
						return errStoreNotFound
					}
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Deleted Secret Store '%s'\n", storeID),
		},
		{
			args: fmt.Sprintf("delete --store-id %s --json", storeID),
			api: mock.API{
				DeleteSecretStoreFn: func(i *fastly.DeleteSecretStoreInput) error {
					if i.StoreID != storeID {
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
			args := testutil.SplitArgs(secretstore.RootNameStore + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.DeleteSecretStoreFn
			var apiInvoked bool
			testcase.api.DeleteSecretStoreFn = func(i *fastly.DeleteSecretStoreInput) error {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DeleteSecretStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDescribeStoreCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

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
				GetSecretStoreFn: func(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("get --store-id %s", storeID),
			api: mock.API{
				GetSecretStoreFn: func(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
					return &fastly.SecretStore{
						StoreID: i.StoreID,
						Name:    storeName,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fmtStore(&fastly.SecretStore{
				StoreID: storeID,
				Name:    storeName,
			}),
		},
		{
			args: fmt.Sprintf("get --store-id %s --json", storeID),
			api: mock.API{
				GetSecretStoreFn: func(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
					return &fastly.SecretStore{
						StoreID: i.StoreID,
						Name:    storeName,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.SecretStore{
				StoreID: storeID,
				Name:    storeName,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstore.RootNameStore + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.GetSecretStoreFn
			var apiInvoked bool
			testcase.api.GetSecretStoreFn = func(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API GetSecretStore invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestListStoresCommand(t *testing.T) {
	const (
		storeName = "test123"
		storeID   = "store-id-123"
	)

	stores := &fastly.SecretStores{
		Meta: fastly.SecretStoreMeta{
			Limit: 123,
		},
		Data: []fastly.SecretStore{
			{StoreID: storeID, Name: storeName},
		},
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
				ListSecretStoresFn: func(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
					return nil, nil
				},
			},
			wantAPIInvoked: true,
		},
		{
			args: "list",
			api: mock.API{
				ListSecretStoresFn: func(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
					return nil, errors.New("unknown error")
				},
			},
			wantAPIInvoked: true,
			wantError:      "unknown error",
		},
		{
			args: "list",
			api: mock.API{
				ListSecretStoresFn: func(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
					return stores, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fmtStores(stores.Data),
		},
		{
			args: "list --json",
			api: mock.API{
				ListSecretStoresFn: func(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
					return stores, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.EncodeJSON([]fastly.SecretStore{stores.Data[0]}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			args := testutil.SplitArgs(secretstore.RootNameStore + " " + testcase.args)
			opts := testutil.MockGlobalData(args, &stdout)

			f := testcase.api.ListSecretStoresFn
			var apiInvoked bool
			testcase.api.ListSecretStoresFn = func(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
				apiInvoked = true
				return f(i)
			}

			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(args, nil)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListSecretStores invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}
