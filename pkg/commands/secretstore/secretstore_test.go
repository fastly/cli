package secretstore_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/secretstore"
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
						ID:   storeID,
						Name: i.Name,
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
						ID:        storeID,
						Name:      i.Name,
						CreatedAt: now,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.JSON(`{"id": %q, "name": %q, "created_at": %q}`, storeID, storeName, now.Format(time.RFC3339Nano)),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameStore+" "+testcase.args), &stdout)

			f := testcase.api.CreateSecretStoreFn
			var apiInvoked bool
			testcase.api.CreateSecretStoreFn = func(i *fastly.CreateSecretStoreInput) (*fastly.SecretStore, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

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
				DeleteSecretStoreFn: func(i *fastly.DeleteSecretStoreInput) error {
					if i.ID != storeID {
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
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameStore+" "+testcase.args), &stdout)

			f := testcase.api.DeleteSecretStoreFn
			var apiInvoked bool
			testcase.api.DeleteSecretStoreFn = func(i *fastly.DeleteSecretStoreInput) error {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

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
						ID:   i.ID,
						Name: storeName,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fmtStore(&fastly.SecretStore{
				ID:   storeID,
				Name: storeName,
			}),
		},
		{
			args: fmt.Sprintf("get --store-id %s --json", storeID),
			api: mock.API{
				GetSecretStoreFn: func(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
					return &fastly.SecretStore{
						ID:   i.ID,
						Name: storeName,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.SecretStore{
				ID:   storeID,
				Name: storeName,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameStore+" "+testcase.args), &stdout)

			f := testcase.api.GetSecretStoreFn
			var apiInvoked bool
			testcase.api.GetSecretStoreFn = func(i *fastly.GetSecretStoreInput) (*fastly.SecretStore, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

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
			{ID: storeID, Name: storeName},
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
			opts := testutil.NewRunOpts(testutil.Args(secretstore.RootNameStore+" "+testcase.args), &stdout)

			f := testcase.api.ListSecretStoresFn
			var apiInvoked bool
			testcase.api.ListSecretStoresFn = func(i *fastly.ListSecretStoresInput) (*fastly.SecretStores, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListSecretStores invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}
