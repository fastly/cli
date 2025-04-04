package configstoreentry_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v10/fastly"

	root "github.com/fastly/cli/pkg/commands/configstoreentry"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestCreateEntryCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "key"
		itemValue = "the-value"
	)
	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--key a-key --value a-value",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				CreateConfigStoreItemFn: func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				CreateConfigStoreItemFn: func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID: i.StoreID,
						Key:     i.Key,
						Value:   i.Value,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Created key '%s' in Config Store '%s'", itemKey, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --json", storeID, itemKey, itemValue),
			API: mock.API{
				CreateConfigStoreItemFn: func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID:   i.StoreID,
						Key:       i.Key,
						Value:     i.Value,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStoreItem{
				StoreID:   storeID,
				Key:       itemKey,
				Value:     itemValue,
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "create"}, scenarios)
}

func TestDeleteEntryCommand(t *testing.T) {
	const (
		storeID = "store-id-123"
		itemKey = "key"
	)

	now := time.Now()

	testItems := make([]*fastly.ConfigStoreItem, 3)
	for i := range testItems {
		testItems[i] = &fastly.ConfigStoreItem{
			StoreID:   storeID,
			Key:       fmt.Sprintf("key-%02d", i),
			Value:     fmt.Sprintf("value %02d", i),
			CreatedAt: &now,
			UpdatedAt: &now,
		}
	}

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
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key '%s' from Config Store '%s'", itemKey, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(struct {
				StoreID string `json:"store_id"`
				Key     string `json:"key"`
				Deleted bool   `json:"deleted"`
			}{
				storeID,
				itemKey,
				true,
			}),
		},
		{
			Args: fmt.Sprintf("--store-id %s --all --auto-yes", storeID),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return nil
				},
			},
			WantOutput: fmt.Sprintf(`Deleting key: key-00
Deleting key: key-01
Deleting key: key-02

SUCCESS: Deleted all keys from Config Store '%s'
`, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --all --auto-yes", storeID),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return errors.New("whoops")
				},
			},
			WantError: "failed to delete keys: key-00, key-01, key-02",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "delete"}, scenarios)
}

func TestDescribeEntryCommand(t *testing.T) {
	const (
		storeID = "store-id-123"
		itemKey = "key"
	)
	now := time.Now()

	testItem := &fastly.ConfigStoreItem{
		StoreID:   storeID,
		Key:       itemKey,
		Value:     "a value",
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--key a-key",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetConfigStoreItemFn: func(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s", storeID, itemKey),
			API: mock.API{
				GetConfigStoreItemFn: func(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID:   i.StoreID,
						Key:       i.Key,
						Value:     "a value",
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: printConfigStoreItem(testItem),
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --json", storeID, itemKey),
			API: mock.API{
				GetConfigStoreItemFn: func(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID:   i.StoreID,
						Key:       i.Key,
						Value:     "a value",
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(testItem),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "describe"}, scenarios)
}

func TestListEntriesCommand(t *testing.T) {
	const storeID = "store-id-123"

	now := time.Now()

	testItems := make([]*fastly.ConfigStoreItem, 3)
	for i := range testItems {
		testItems[i] = &fastly.ConfigStoreItem{
			StoreID:   storeID,
			Key:       fmt.Sprintf("key-%02d", i),
			Value:     fmt.Sprintf("value %02d", i),
			CreatedAt: &now,
			UpdatedAt: &now,
		}
	}

	scenarios := []testutil.CLIScenario{
		{
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s", storeID),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
			},
			WantOutput: printConfigStoreItemsTbl(testItems),
		},
		{
			Args: fmt.Sprintf("--store-id %s --json", storeID),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(testItems),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "list"}, scenarios)
}

func TestUpdateEntryCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "key"
		itemValue = "the-value"
	)
	now := time.Now()

	scenarios := []testutil.CLIScenario{
		{
			Args:      "--key a-key --value a-value",
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				UpdateConfigStoreItemFn: func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			API: mock.API{
				UpdateConfigStoreItemFn: func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID: i.StoreID,
						Key:     i.Key,
						Value:   i.Value,
					}, nil
				},
			},
			WantOutput: fstfmt.Success("Updated config store item %s in store %s", itemKey, storeID),
		},
		{
			Args: fmt.Sprintf("--store-id %s --key %s --value %s --json", storeID, itemKey, itemValue+"updated"),
			API: mock.API{
				UpdateConfigStoreItemFn: func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID:   i.StoreID,
						Key:       i.Key,
						Value:     i.Value,
						CreatedAt: &now,
						UpdatedAt: &now,
					}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(&fastly.ConfigStoreItem{
				StoreID:   storeID,
				Key:       itemKey,
				Value:     itemValue + "updated",
				CreatedAt: &now,
				UpdatedAt: &now,
			}),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "update"}, scenarios)
}

func printConfigStoreItem(i *fastly.ConfigStoreItem) string {
	var b bytes.Buffer
	text.PrintConfigStoreItem(&b, "", i)
	return b.String()
}

func printConfigStoreItemsTbl(i []*fastly.ConfigStoreItem) string {
	var b bytes.Buffer
	text.PrintConfigStoreItemsTbl(&b, i)
	return b.String()
}
