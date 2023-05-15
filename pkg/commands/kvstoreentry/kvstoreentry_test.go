package kvstoreentry_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/kvstoreentry"
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

	type ts struct {
		testutil.TestScenario

		PartialMatch bool
		Stdin        io.Reader
	}

	scenarios := []ts{
		{
			TestScenario: testutil.TestScenario{
				Args:      testutil.Args(kvstoreentry.RootName + " create --key a-key --value a-value"),
				WantError: "error parsing arguments: required flag --store-id not provided",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --key %s --value %s", kvstoreentry.RootName, storeID, itemKey, itemValue)),
				API: mock.API{
					InsertKVStoreKeyFn: func(i *fastly.InsertKVStoreKeyInput) error {
						return errors.New("invalid request")
					},
				},
				WantError: "invalid request",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --key %s --value %s", kvstoreentry.RootName, storeID, itemKey, itemValue)),
				API: mock.API{
					InsertKVStoreKeyFn: func(i *fastly.InsertKVStoreKeyInput) error {
						return nil
					},
				},
				WantOutput: fstfmt.Success("Inserted key %s into KV Store %s", itemKey, storeID),
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --key %s --value %s --json", kvstoreentry.RootName, storeID, itemKey, itemValue)),
				API: mock.API{
					InsertKVStoreKeyFn: func(i *fastly.InsertKVStoreKeyInput) error {
						return nil
					},
				},
				WantOutput: fstfmt.JSON(`{"id": %q, "key": %q}`, storeID, itemKey),
			},
		},
		{
			Stdin: strings.NewReader(`{"key":"example","value":"VkFMVUU="}`),
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --stdin", kvstoreentry.RootName, storeID)),
				API: mock.API{
					BatchModifyKVStoreKeyFn: func(i *fastly.BatchModifyKVStoreKeyInput) error {
						return nil
					},
				},
				WantOutput: "\nSUCCESS: Inserted keys into KV Store\n",
			},
		},
		{
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --file %s", kvstoreentry.RootName, storeID, filepath.Join("testdata", "data.json"))),
				API: mock.API{
					BatchModifyKVStoreKeyFn: func(i *fastly.BatchModifyKVStoreKeyInput) error {
						return nil
					},
				},
				WantOutput: "\nSUCCESS: Inserted keys into KV Store\n",
			},
		},
		{
			PartialMatch: true,
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --dir %s", kvstoreentry.RootName, storeID, filepath.Join("testdata", "example"))),
				API: mock.API{
					InsertKVStoreKeyFn: func(i *fastly.InsertKVStoreKeyInput) error {
						return nil
					},
				},
				WantOutput: "SUCCESS: Inserted 1 keys into KV Store",
			},
		},
		{
			PartialMatch: true,
			TestScenario: testutil.TestScenario{
				Args: testutil.Args(fmt.Sprintf("%s create --store-id %s --dir %s --dir-allow-hidden", kvstoreentry.RootName, storeID, filepath.Join("testdata", "example"))),
				API: mock.API{
					InsertKVStoreKeyFn: func(i *fastly.InsertKVStoreKeyInput) error {
						return nil
					},
				},
				WantOutput: "SUCCESS: Inserted 2 keys into KV Store",
			},
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			if testcase.Stdin != nil {
				opts.Stdin = testcase.Stdin
			}

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.WantError)

			if testcase.PartialMatch {
				testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
			} else {
				testutil.AssertString(t, testcase.WantOutput, stdout.String())
			}
		})
	}
}

func TestDeleteCommand(t *testing.T) {
	const (
		storeID = "store-id-123"
		itemKey = "foo"
	)

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(kvstoreentry.RootName + " delete --key a-key"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --key %s", kvstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				DeleteKVStoreKeyFn: func(i *fastly.DeleteKVStoreKeyInput) error {
					return errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --key %s", kvstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				DeleteKVStoreKeyFn: func(i *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.Success("Deleted key %s from KV Store %s", itemKey, storeID),
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s delete --store-id %s --key %s --json", kvstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				DeleteKVStoreKeyFn: func(i *fastly.DeleteKVStoreKeyInput) error {
					return nil
				},
			},
			WantOutput: fstfmt.JSON(`{"key": "%s", "id": "%s", "deleted": true}`, itemKey, storeID),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)

			opts.APIClient = mock.APIClient(testcase.API)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestDescribeCommand(t *testing.T) {
	const (
		storeID   = "store-id-123"
		itemKey   = "foo"
		itemValue = "a value"
	)

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(kvstoreentry.RootName + " describe --key a-key"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s describe --store-id %s --key %s", kvstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				GetKVStoreKeyFn: func(i *fastly.GetKVStoreKeyInput) (string, error) {
					return "", errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s describe --store-id %s --key %s", kvstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				GetKVStoreKeyFn: func(i *fastly.GetKVStoreKeyInput) (string, error) {
					return itemValue, nil
				},
			},
			WantOutput: itemValue,
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s describe --store-id %s --key %s --json", kvstoreentry.RootName, storeID, itemKey)),
			API: mock.API{
				GetKVStoreKeyFn: func(i *fastly.GetKVStoreKeyInput) (string, error) {
					return itemValue, nil
				},
			},
			WantOutput: fmt.Sprintf(`{"%s": "%s"}`, itemKey, itemValue) + "\n",
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)

			opts.APIClient = mock.APIClient(testcase.API)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}

func TestListCommand(t *testing.T) {
	const storeID = "store-id-123"

	testItems := make([]string, 3)
	for i := range testItems {
		testItems[i] = fmt.Sprintf("key-%02d", i)
	}

	scenarios := []testutil.TestScenario{
		{
			Args:      testutil.Args(kvstoreentry.RootName + " list"),
			WantError: "error parsing arguments: required flag --store-id not provided",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list --store-id %s", kvstoreentry.RootName, storeID)),
			API: mock.API{
				ListKVStoreKeysFn: func(i *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return nil, errors.New("invalid request")
				},
			},
			WantError: "invalid request",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list --store-id %s", kvstoreentry.RootName, storeID)),
			API: mock.API{
				ListKVStoreKeysFn: func(i *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: testItems}, nil
				},
			},
			WantOutput: strings.Join(testItems, "\n") + "\n",
		},
		{
			Args: testutil.Args(fmt.Sprintf("%s list --store-id %s --json", kvstoreentry.RootName, storeID)),
			API: mock.API{
				ListKVStoreKeysFn: func(i *fastly.ListKVStoreKeysInput) (*fastly.ListKVStoreKeysResponse, error) {
					return &fastly.ListKVStoreKeysResponse{Data: testItems}, nil
				},
			},
			WantOutput: fstfmt.EncodeJSON(testItems),
		},
	}

	for _, testcase := range scenarios {
		testcase := testcase
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)

			opts.APIClient = mock.APIClient(testcase.API)

			err := app.Run(opts)

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertString(t, testcase.WantOutput, stdout.String())
		})
	}
}
