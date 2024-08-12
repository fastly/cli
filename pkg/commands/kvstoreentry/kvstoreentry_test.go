package kvstoreentry_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/kvstoreentry"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestCreateCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "foo"
		itemValue = "the-value"
	)

	scenarios := []testutil.TestScenario{
		{
			Arg:       "--key a-key --value a-value",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ *fastly.InsertKVStoreKeyInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ *fastly.InsertKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s --value %s --json", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ *fastly.InsertKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "key": %q}`, storeID, itemKey),
		},
		{
			Arg:   fmt.Sprintf("--store-id %s --stdin", storeID),
			Stdin: []string{`{"key":"example","value":"VkFMVUU="}`},
			API: mock.API{
				BatchModifyKVStoreKeyFn: func(_ *fastly.BatchModifyKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "SUCCESS: Inserted keys into KV Store\n",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --file %s", storeID, filepath.Join("testdata", "data.json")),
			API: mock.API{
				BatchModifyKVStoreKeyFn: func(_ *fastly.BatchModifyKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "SUCCESS: Inserted keys into KV Store\n",
		},
		{
			Arg:   fmt.Sprintf("--store-id %s --dir %s", storeID, filepath.Join("testdata", "example")),
			Stdin: []string{"y"},
			API: mock.API{
				InsertKVStoreKeyFn: func(_ *fastly.InsertKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "SUCCESS: Inserted 1 keys into KV Store",
		},
		{
			Arg:   fmt.Sprintf("--store-id %s --dir %s --dir-allow-hidden", storeID, filepath.Join("testdata", "example")),
			Stdin: []string{"y"},
			API: mock.API{
				InsertKVStoreKeyFn: func(_ *fastly.InsertKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "SUCCESS: Inserted 2 keys into KV Store",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDeleteCommand(t *testing.T) {
	const (
		storeID = "store-id-123"
		itemKey = "foo"
	)

	scenarios := []testutil.TestScenario{
		{
			Arg:       "--key a-key",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Arg:       "--store-id " + storeID,
			WantError: "invalid command, neither --all or --key provided",
		},
		{
			Arg:       "--json --all --store-id " + storeID,
			WantError: "invalid flag combination, --all and --json",
		},
		{
			Arg:       "--key a-key --all --store-id " + storeID,
			WantError: "invalid flag combination, --all and --key",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ *fastly.DeleteKVStoreKeyInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key '%s' from KV Store '%s'", itemKey, storeID),
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"key": "%s", "store_id": "%s", "deleted": true}`, itemKey, storeID),
		},
		{
			Arg: fmt.Sprintf("--store-id %s --all --auto-yes", storeID),
			API: mock.API{
				NewListKVStoreKeysPaginatorFn: func(_ *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries {
					return &mockKVStoresEntriesPaginator{
						next: true,
						keys: []string{"foo", "bar", "baz"},
					}
				},
				DeleteKVStoreKeyFn: func(_ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "Deleting keys...",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --all --auto-yes", storeID),
			API: mock.API{
				NewListKVStoreKeysPaginatorFn: func(_ *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries {
					return &mockKVStoresEntriesPaginator{
						next: true,
						keys: []string{"foo", "bar", "baz"},
					}
				},
				DeleteKVStoreKeyFn: func(_ *fastly.DeleteKVStoreKeyInput) error {
					return errors.New("whoops")
				},
			},
			WantError: "failed to delete 3 keys",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestGetCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "foo"
		itemValue = "a value"
	)

	scenarios := []testutil.TestScenario{
		{
			Arg:       "--key a-key",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreKeyFn: func(_ *fastly.GetKVStoreKeyInput) (string, error) {
					return "", errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreKeyFn: func(_ *fastly.GetKVStoreKeyInput) (string, error) {
					return itemValue, nil
				},
			},
			WantOutput: itemValue,
		},
		{
			Arg: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				GetKVStoreKeyFn: func(_ *fastly.GetKVStoreKeyInput) (string, error) {
					return itemValue, nil
				},
			},
			WantOutput: fmt.Sprintf(`{"%s": "%s"}`, itemKey, itemValue) + "\n",
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "get"}, scenarios)
}

func TestListCommand(t *testing.T) {
	const storeID = "store-id-123"

	testItems := make([]string, 3)
	for i := range testItems {
		testItems[i] = fmt.Sprintf("key-%02d", i)
	}

	scenarios := []testutil.TestScenario{
		{
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Arg: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Arg: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: testItems}, nil
				},
			},
			WantOutput: strings.Join(testItems, "\n") + "\n",
		},
		{
			Arg: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: testItems}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(testItems),
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

type mockKVStoresEntriesPaginator struct {
	next bool
	keys []string
	err  error
}

func (m *mockKVStoresEntriesPaginator) Next() bool {
	ret := m.next
	if m.next {
		m.next = false // allow one instance of true before stopping
	}
	return ret
}

func (m *mockKVStoresEntriesPaginator) Keys() []string {
	return m.keys
}

func (m *mockKVStoresEntriesPaginator) Err() error {
	return m.err
}
