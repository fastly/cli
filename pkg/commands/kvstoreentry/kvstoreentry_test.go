package kvstoreentry_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v11/fastly"

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

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--key a-key --value a-value",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, _ *fastly.InsertKVStoreKeyInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, _ *fastly.InsertKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --json", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, _ *fastly.InsertKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"id": %q, "key": %q}`, storeID, itemKey),
		},
		{
			Name: "validate --add flag",
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --add", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if !i.Add {
						return errors.New("expected Add flag to be true")
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Name: "validate --append flag",
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --append", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if !i.Append {
						return errors.New("expected Append flag to be true")
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Name: "validate --prepend flag",
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --prepend", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if !i.Prepend {
						return errors.New("expected Prepend flag to be true")
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Name: "validate --background-fetch flag",
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --background-fetch", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if !i.BackgroundFetch {
						return errors.New("expected BackgroundFetch flag to be true")
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Name: "validate --if-generation-match flag",
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --if-generation-match 42", storeID, itemKey, itemValue),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if i.IfGenerationMatch != 42 {
						return fmt.Errorf("expected IfGenerationMatch to be 42, got %d", i.IfGenerationMatch)
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Name:      "validate --if-generation-match flag with invalid value",
			Args:      fmt.Sprintf("--store-id %s --key %s --value %s --if-generation-match invalid", storeID, itemKey, itemValue),
			WantError: "invalid generation value: invalid",
		},
		{
			Name: "validate --metadata flag",
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --metadata %s", storeID, itemKey, itemValue, "test-metadata"),
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if i.Metadata == nil || *i.Metadata != "test-metadata" {
						return errors.New("expected Metadata to be 'test-metadata'")
					}
					return nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in KV Store '%s'", itemKey, storeID),
		},
		{
			Args:  fmt.Sprintf("--store-id %s --stdin", storeID),
			Stdin: []string{`{"key":"example","value":"VkFMVUU="}`},
			API: mock.API{
				BatchModifyKVStoreKeyFn: func(_ context.Context, _ *fastly.BatchModifyKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "SUCCESS: Inserted keys into KV Store\n",
		},
		{
			Args: fmt.Sprintf("--store-id %s --file %s", storeID, filepath.Join("testdata", "data.json")),
			API: mock.API{
				BatchModifyKVStoreKeyFn: func(_ context.Context, _ *fastly.BatchModifyKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "SUCCESS: Inserted keys into KV Store\n",
		},
		{
			Args:  fmt.Sprintf("--store-id %s --dir %s", storeID, filepath.Join("testdata", "example")),
			Stdin: []string{"y"},
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if i.Key == "foo.txt" {
						return nil
					}
					return errors.New("invalid request")
				},
			},
			WantOutput: "SUCCESS: Inserted 1 keys into KV Store",
		},
		{
			Args:  fmt.Sprintf("--store-id %s --dir %s --dir-allow-hidden", storeID, filepath.Join("testdata", "example")),
			Stdin: []string{"y"},
			API: mock.API{
				InsertKVStoreKeyFn: func(_ context.Context, i *fastly.InsertKVStoreKeyInput) error {
					if i.Key == "foo.txt" || i.Key == ".hiddenfile" {
						return nil
					}
					return errors.New("invalid request")
				},
			},
			WantOutput: "SUCCESS: Inserted 2 keys into KV Store",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDeleteCommand(t *testing.T) {
	const (
		storeID = "store-id-123"
		itemKey = "foo"
	)

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--key a-key",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args:      "--store-id " + storeID,
			WantError: "invalid command, neither --all or --key provided",
		},
		{
			Args:      "--json --all --store-id " + storeID,
			WantError: "invalid flag combination, --all and --json",
		},
		{
			Args:      "--key a-key --all --store-id " + storeID,
			WantError: "invalid flag combination, --all and --key",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key '%s' from KV Store '%s'", itemKey, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"key": "%s", "store_id": "%s", "deleted": true}`, itemKey, storeID),
		},
		{
			Name: "validate --force flag with any key",
			Args: fmt.Sprintf("--store-id %s --key %s --force", storeID, "myFakeKey"),
			API: mock.API{
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key '%s' from KV Store '%s'", "myFakeKey", storeID),
		},
		{
			Name: "validate --if-generation-match with matching generation",
			Args: fmt.Sprintf("--store-id %s --key %s --if-generation-match 123", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{Generation: 123}, nil
				},
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key '%s' from KV Store '%s'", itemKey, storeID),
		},
		{
			Name: "validate --if-generation-match with invalid generation value",
			Args: fmt.Sprintf("--store-id %s --key %s --if-generation-match invalid", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{Generation: 123}, nil
				},
			},
			WantError: `invalid generation value: invalid`,
		},
		{
			Args: fmt.Sprintf("--store-id %s --all --auto-yes", storeID),
			API: mock.API{
				NewListKVStoreKeysPaginatorFn: func(_ context.Context, _ *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries {
					return &mockKVStoresEntriesPaginator{
						next: true,
						keys: []string{"foo", "bar", "baz"},
					}
				},
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: "Deleting keys...",
		},
		{
			Args: fmt.Sprintf("--store-id %s --all --auto-yes", storeID),
			API: mock.API{
				NewListKVStoreKeysPaginatorFn: func(_ context.Context, _ *fastly.ListKVStoreKeysInput) fastly.PaginatorKVStoreEntries {
					return &mockKVStoresEntriesPaginator{
						next: true,
						keys: []string{"foo", "bar", "baz"},
					}
				},
				DeleteKVStoreKeyFn: func(_ context.Context, _ *fastly.DeleteKVStoreKeyInput) error {
					return errors.New("whoops")
				},
			},
			WantError: "failed to delete 3 keys",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestGetCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "foo"
		itemValue = "a value"
	)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --store-id flag",
			Args:      "--key a-key",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Name:      "validate missing --key flag",
			Args:      "--store-id " + storeID,
			WantError: "error parsing arguments: required flag --key not provided",
		},
		{
			Name: "validate API error handling",
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{}, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Name: "validate successful get operation",
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Value:      io.NopCloser(strings.NewReader(itemValue)),
					}, nil
				},
			},
			WantOutput: itemValue,
		},
		{
			Name: "validate --json flag output",
			Args: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Value:      io.NopCloser(strings.NewReader(itemValue)),
					}, nil
				},
			},
			WantOutput: fmt.Sprintf(`{"%s": "%s"}`, itemKey, base64.StdEncoding.EncodeToString([]byte(itemValue))) + "\n",
		},
		{
			Name:      "validate --verbose flag error",
			Args:      fmt.Sprintf("--store-id %s --key %s --verbose", storeID, itemKey),
			WantError: "the 'get' command does not support the --verbose flag",
		},
		{
			Name: "validate --if-generation-match with matching generation",
			Args: fmt.Sprintf("--store-id %s --key %s --if-generation-match 123", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Value:      io.NopCloser(strings.NewReader(itemValue)),
					}, nil
				},
			},
			WantOutput: itemValue,
		},
		{
			Name: "validate --if-generation-match with non-matching generation",
			Args: fmt.Sprintf("--store-id %s --key %s --if-generation-match 123", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 456,
						Value:      io.NopCloser(strings.NewReader(itemValue)),
					}, nil
				},
			},
			WantError: "generation value does not match: expected 456, got 123",
		},
		{
			Name: "validate --if-generation-match with invalid generation value",
			Args: fmt.Sprintf("--store-id %s --key %s --if-generation-match invalid", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Value:      io.NopCloser(strings.NewReader(itemValue)),
					}, nil
				},
			},
			WantError: "invalid generation value: invalid",
		},
		{
			Name: "validate handling of nil value reader",
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Value:      nil,
					}, nil
				},
			},
			WantOutput: "",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "get"}, scenarios)
}

func TestDescribeCommand(t *testing.T) {
	const (
		storeID      = "store-id-123"
		itemKey      = "foo"
		itemMetadata = "test-metadata"
	)

	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --store-id flag",
			Args:      "--key a-key",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Name:      "validate missing --key flag",
			Args:      "--store-id " + storeID,
			WantError: "error parsing arguments: required flag --key not provided",
		},
		{
			Name: "validate API error handling",
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{}, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Name: "validate successful describe operation",
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Metadata:   itemMetadata,
					}, nil
				},
			},
			WantOutput: fmt.Sprintf("Key: %s\nGeneration: %d\nMetadata: %s\n", itemKey, 123, itemMetadata),
		},
		{
			Name: "validate --json flag output",
			Args: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Metadata:   itemMetadata,
					}, nil
				},
			},
			WantOutput: fmt.Sprintf("{\n  \"generation\": \"%d\",\n  \"key\": \"%s\",\n  \"metadata\": \"%s\"\n}\n", 123, itemKey, itemMetadata),
		},
		{
			Name: "validate --verbose flag output",
			Args: fmt.Sprintf("--store-id %s --key %s --verbose", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 123,
						Metadata:   itemMetadata,
					}, nil
				},
			},
			WantOutput: fmt.Sprintf("Key: %s\nGeneration: %d\nMetadata: %s\n", itemKey, 123, itemMetadata),
		},
		{
			Name: "validate handling of empty metadata",
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetKVStoreItemFn: func(_ context.Context, _ *fastly.GetKVStoreItemInput) (fastly.GetKVStoreItemOutput, error) {
					return fastly.GetKVStoreItemOutput{
						Generation: 456,
						Metadata:   "",
					}, nil
				},
			},
			WantOutput: fmt.Sprintf("Key: %s\nGeneration: %d\nMetadata: \n", itemKey, 456),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestListCommand(t *testing.T) {
	const storeID = "store-id-123"

	testItems := make([]string, 3)
	for i := range testItems {
		testItems[i] = fmt.Sprintf("key-%02d", i)
	}

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ context.Context, _ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ context.Context, _ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: testItems}, nil
				},
			},
			WantOutput: strings.Join(testItems, "\n") + "\n",
		},
		{
			Name: "validate --prefix param",
			Args: fmt.Sprintf("--store-id %s --prefix=foo", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ context.Context, _ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: []string{"foo-key1", "foo-key2"}}, nil
				},
			},
			WantOutput: "✓ Getting data\nfoo-key1\nfoo-key2\n",
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				ListKVStoreKeysFn: func(_ context.Context, _ *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: testItems}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(testItems),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
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
