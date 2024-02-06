package configstoreentry_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/configstoreentry"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/threadsafe"
)

func TestCreateEntryCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "key"
		itemValue = "the-value"
	)
	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstoreentry.RootName + " create --key a-key --value a-value"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --key %s --value %s", configstoreentry.RootName, storeID, itemKey, itemValue)),
			API: mock.API{
				CreateConfigStoreItemFn: func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --key %s --value %s", configstoreentry.RootName, storeID, itemKey, itemValue)),
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
			Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --key %s --value %s --json", configstoreentry.RootName, storeID, itemKey, itemValue)),
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

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstoreentry.RootName + " delete --key a-key"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args:      testutil.Args(configstoreentry.RootName + " delete --store-id " + storeID),
			WantError: "invalid command, neither --all or --key provided",
		},
		{
			Args:      testutil.Args(configstoreentry.RootName + " delete --json --all --store-id " + storeID),
			WantError: "invalid flag combination, --all and --json",
		},
		{
			Args:      testutil.Args(configstoreentry.RootName + " delete --key a-key --all --store-id " + storeID),
			WantError: "invalid flag combination, --all and --key",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --key %s", configstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --key %s", configstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key '%s' from Config Store '%s'", itemKey, storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --key %s --json", configstoreentry.RootName, storeID, itemKey)),
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
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --all --auto-yes", configstoreentry.RootName, storeID)),
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
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --all --auto-yes", configstoreentry.RootName, storeID)),
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

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout threadsafe.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
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

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstoreentry.RootName + " describe --key a-key"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s describe --store-id %s --key %s", configstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				GetConfigStoreItemFn: func(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s describe --store-id %s --key %s", configstoreentry.RootName, storeID, itemKey)),
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
			Args: testutil.Args(fmt.Sprintf("%s describe --store-id %s --key %s --json", configstoreentry.RootName, storeID, itemKey)),
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

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstoreentry.RootName + " list"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list --store-id %s", configstoreentry.RootName, storeID)),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list --store-id %s", configstoreentry.RootName, storeID)),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
			},
			WantOutput: printConfigStoreItemsTbl(testItems),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list --store-id %s --json", configstoreentry.RootName, storeID)),
			API: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(testItems),
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

func TestUpdateEntryCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "key"
		itemValue = "the-value"
	)
	now := time.Now()

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(configstoreentry.RootName + " update --key a-key --value a-value"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s update --store-id %s --key %s --value %s", configstoreentry.RootName, storeID, itemKey, itemValue)),
			API: mock.API{
				UpdateConfigStoreItemFn: func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s update --store-id %s --key %s --value %s", configstoreentry.RootName, storeID, itemKey, itemValue)),
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
			Args: testutil.Args(fmt.Sprintf("%s update --store-id %s --key %s --value %s --json", configstoreentry.RootName, storeID, itemKey, itemValue+"updated")),
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
