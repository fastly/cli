package configstoreentry_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/configstoreentry"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

func TestCreateEntryCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "key"
		itemValue = "the-value"
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
			args:      "create --key a-key --value a-value",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("create --store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			api: mock.API{
				CreateConfigStoreItemFn: func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("create --store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			api: mock.API{
				CreateConfigStoreItemFn: func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID: i.StoreID,
						Key:     i.Key,
						Value:   i.Value,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Created config store item %s in store %s", itemKey, storeID),
		},
		{
			args: fmt.Sprintf("create --store-id %s --key %s --value %s --json", storeID, itemKey, itemValue),
			api: mock.API{
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
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.ConfigStoreItem{
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
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstoreentry.RootName+" "+testcase.args), &stdout)

			f := testcase.api.CreateConfigStoreItemFn
			var apiInvoked bool
			testcase.api.CreateConfigStoreItemFn = func(i *fastly.CreateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API CreateConfigStoreItem invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
		})
	}
}

func TestDeleteEntryCommand(t *testing.T) {
	const (
		storeID = "store-id-123"
		itemKey = "key"
	)

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "delete --key a-key",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("delete --store-id %s --key %s", storeID, itemKey),
			api: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("delete --store-id %s --key %s", storeID, itemKey),
			api: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Deleted config store item %s from store %s", itemKey, storeID),
		},
		{
			args: fmt.Sprintf("delete --store-id %s --key %s --json", storeID, itemKey),
			api: mock.API{
				DeleteConfigStoreItemFn: func(i *fastly.DeleteConfigStoreItemInput) error {
					return nil
				},
			},
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(struct {
				StoreID string `json:"store_id"`
				Key     string `json:"item_key"`
				Deleted bool   `json:"deleted"`
			}{
				storeID,
				itemKey,
				true,
			}),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstoreentry.RootName+" "+testcase.args), &stdout)

			f := testcase.api.DeleteConfigStoreItemFn
			var apiInvoked bool
			testcase.api.DeleteConfigStoreItemFn = func(i *fastly.DeleteConfigStoreItemInput) error {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API DeleteConfigStoreItem invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
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

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "describe --key a-key",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("describe --store-id %s --key %s", storeID, itemKey),
			api: mock.API{
				GetConfigStoreItemFn: func(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("describe --store-id %s --key %s", storeID, itemKey),
			api: mock.API{
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
			wantAPIInvoked: true,
			wantOutput:     printConfigStoreItem(testItem),
		},
		{
			args: fmt.Sprintf("describe --store-id %s --key %s --json", storeID, itemKey),
			api: mock.API{
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
			wantAPIInvoked: true,
			wantOutput:     fstfmt.EncodeJSON(testItem),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstoreentry.RootName+" "+testcase.args), &stdout)

			f := testcase.api.GetConfigStoreItemFn
			var apiInvoked bool
			testcase.api.GetConfigStoreItemFn = func(i *fastly.GetConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API GetConfigStoreItem invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
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

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "list",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("list --store-id %s", storeID),
			api: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("list --store-id %s", storeID),
			api: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     printConfigStoreItemsTbl(testItems),
		},
		{
			args: fmt.Sprintf("list --store-id %s --json", storeID),
			api: mock.API{
				ListConfigStoreItemsFn: func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
					return testItems, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.EncodeJSON(testItems),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstoreentry.RootName+" "+testcase.args), &stdout)

			f := testcase.api.ListConfigStoreItemsFn
			var apiInvoked bool
			testcase.api.ListConfigStoreItemsFn = func(i *fastly.ListConfigStoreItemsInput) ([]*fastly.ConfigStoreItem, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API ListConfigStoreItems invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
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

	scenarios := []struct {
		args           string
		api            mock.API
		wantAPIInvoked bool
		wantError      string
		wantOutput     string
	}{
		{
			args:      "update --key a-key --value a-value",
			wantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			args: fmt.Sprintf("update --store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			api: mock.API{
				UpdateConfigStoreItemFn: func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return nil, errors.New("invalid request")
				},
			},
			wantAPIInvoked: true,
			wantError:      "invalid request",
		},
		{
			args: fmt.Sprintf("update --store-id %s --key %s --value %s", storeID, itemKey, itemValue),
			api: mock.API{
				UpdateConfigStoreItemFn: func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
					return &fastly.ConfigStoreItem{
						StoreID: i.StoreID,
						Key:     i.Key,
						Value:   i.Value,
					}, nil
				},
			},
			wantAPIInvoked: true,
			wantOutput:     fstfmt.Success("Updated config store item %s in store %s", itemKey, storeID),
		},
		{
			args: fmt.Sprintf("update --store-id %s --key %s --value %s --json", storeID, itemKey, itemValue+"updated"),
			api: mock.API{
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
			wantAPIInvoked: true,
			wantOutput: fstfmt.EncodeJSON(&fastly.ConfigStoreItem{
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
		t.Run(testcase.args, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testutil.Args(configstoreentry.RootName+" "+testcase.args), &stdout)

			f := testcase.api.UpdateConfigStoreItemFn
			var apiInvoked bool
			testcase.api.UpdateConfigStoreItemFn = func(i *fastly.UpdateConfigStoreItemInput) (*fastly.ConfigStoreItem, error) {
				apiInvoked = true
				return f(i)
			}

			opts.APIClient = mock.APIClient(testcase.api)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
			if apiInvoked != testcase.wantAPIInvoked {
				t.Fatalf("API UpdateConfigStoreItem invoked = %v, want %v", apiInvoked, testcase.wantAPIInvoked)
			}
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
