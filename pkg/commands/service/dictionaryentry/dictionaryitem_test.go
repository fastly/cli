package dictionaryentry_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/dictionaryentry"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v12/fastly"
)

func TestDictionaryItemDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123 --key foo",
			API:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			WantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			Args:      "--service-id 123 --dictionary-id 456",
			API:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			WantError: "error parsing arguments: required flag --key not provided",
		},
		{
			Args:       "--service-id 123 --dictionary-id 456 --key foo",
			API:        mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			WantOutput: describeDictionaryItemOutput,
		},
		{
			Args:       "--service-id 123 --dictionary-id 456 --key foo-deleted",
			API:        mock.API{GetDictionaryItemFn: describeDictionaryItemOKDeleted},
			WantOutput: describeDictionaryItemOutputDeleted,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestDictionaryItemsList(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			WantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			Args:      "--dictionary-id 456",
			WantError: "error reading service: no service ID found",
		},
		{
			API: mock.API{
				GetDictionaryItemsFn: func(ctx context.Context, _ *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem] {
					return fastly.NewPaginator[fastly.DictionaryItem](ctx, &mock.HTTPClient{
						Errors: []error{
							testutil.Err,
						},
						Responses: []*http.Response{nil},
					}, fastly.ListOpts{}, "/example")
				},
			},
			Args:      "--service-id 123 --dictionary-id 456",
			WantError: testutil.Err.Error(),
		},
		{
			API: mock.API{
				GetDictionaryItemsFn: func(ctx context.Context, _ *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem] {
					return fastly.NewPaginator[fastly.DictionaryItem](ctx, &mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[
                  {
                    "dictionary_id": "123",
                    "item_key": "foo",
                    "item_value": "bar",
                    "created_at": "2021-06-15T23:00:00Z",
                    "updated_at": "2021-06-15T23:00:00Z"
                  },
                  {
                    "dictionary_id": "456",
                    "item_key": "baz",
                    "item_value": "qux",
                    "created_at": "2021-06-15T23:00:00Z",
                    "updated_at": "2021-06-15T23:00:00Z",
                    "deleted_at": "2021-06-15T23:00:00Z"
                  }
                ]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			Args:       "--service-id 123 --dictionary-id 456 --per-page 1",
			WantOutput: listDictionaryItemsOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestDictionaryItemCreate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			API:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			WantError: "error parsing arguments: required flag ",
		},
		{
			Args:      "--service-id 123 --dictionary-id 456",
			API:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			WantError: "error parsing arguments: required flag ",
		},
		{
			Args:       "--service-id 123 --dictionary-id 456 --key foo --value bar",
			API:        mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			WantOutput: "SUCCESS: Created dictionary item foo (service 123, dictionary 456)\n",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "create"}, scenarios)
}

func TestDictionaryItemUpdate(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			API:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			WantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			Args:      "--service-id 123 --dictionary-id 456",
			API:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			WantError: "an empty value is not allowed for either the '--key' or '--value' flags",
		},
		{
			Args:       "--service-id 123 --dictionary-id 456 --key foo --value bar",
			API:        mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			WantOutput: updateDictionaryItemOutput,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)

	// File-based test: invalid json
	t.Run("invalid json file", func(t *testing.T) {
		filePath := testutil.MakeTempFile(t, `{invalid": "json"}`)
		defer os.RemoveAll(filePath)

		scenarios := []testutil.CLIScenario{
			{
				Args:      "--service-id 123 --dictionary-id 456 --file " + filePath,
				WantError: "invalid character 'i' looking for beginning of object key string",
			},
		}
		testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
	})

	// NOTE: We don't specify the full error value in the WantError field
	// because this would cause an error on different OS'. For example, Unix
	// systems report 'no such file or directory', while Windows will report
	// 'The system cannot find the file specified'.
	t.Run("missing file", func(t *testing.T) {
		scenarios := []testutil.CLIScenario{
			{
				Args:      "--service-id 123 --dictionary-id 456 --file missingPath",
				WantError: "open missingPath:",
			},
		}
		testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
	})

	// File-based test: batch modify error
	t.Run("batch modify error", func(t *testing.T) {
		filePath := testutil.MakeTempFile(t, dictionaryItemBatchModifyInputOK)
		defer os.RemoveAll(filePath)

		scenarios := []testutil.CLIScenario{
			{
				Args:      "--service-id 123 --dictionary-id 456 --file " + filePath,
				API:       mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsError},
				WantError: errTest.Error(),
			},
		}
		testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
	})

	// File-based test: batch modify success
	t.Run("batch modify success", func(t *testing.T) {
		filePath := testutil.MakeTempFile(t, dictionaryItemBatchModifyInputOK)
		defer os.RemoveAll(filePath)

		scenarios := []testutil.CLIScenario{
			{
				Args:       "--service-id 123 --dictionary-id 456 --file " + filePath,
				API:        mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsOK},
				WantOutput: "SUCCESS: Made 4 modifications of Dictionary 456 on service 123\n",
			},
		}
		testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
	})
}

func TestDictionaryItemDelete(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "--service-id 123",
			API:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			WantError: "error parsing arguments: required flag ",
		},
		{
			Args:      "--service-id 123 --dictionary-id 456",
			API:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			WantError: "error parsing arguments: required flag ",
		},
		{
			Args:       "--service-id 123 --dictionary-id 456 --key foo",
			API:        mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			WantOutput: "SUCCESS: Deleted dictionary item foo (service 123, dictionary 456)\n",
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "delete"}, scenarios)
}

func describeDictionaryItemOK(_ context.Context, i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      fastly.ToPointer(i.ItemKey),
		ItemValue:    fastly.ToPointer("bar"),
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

var describeDictionaryItemOutput = "\n" + `Service ID: 123
Dictionary ID: 456
Item Key: foo
Item Value: bar
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`

var updateDictionaryItemOutput = `SUCCESS: Updated dictionary item (service 123)

Dictionary ID: 456
Item Key: foo
Item Value: bar
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`

func describeDictionaryItemOKDeleted(_ context.Context, i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      fastly.ToPointer(i.ItemKey),
		ItemValue:    fastly.ToPointer("bar"),
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		DeletedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:06:08Z"),
	}, nil
}

var describeDictionaryItemOutputDeleted = "\n" + strings.TrimSpace(`
Service ID: 123
Dictionary ID: 456
Item Key: foo-deleted
Item Value: bar
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
Deleted (UTC): 2001-02-03 04:06
`) + "\n"

var listDictionaryItemsOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Item: 1/2
	Dictionary ID: 123
	Item Key: foo
	Item Value: bar
	Created (UTC): 2021-06-15 23:00
	Last edited (UTC): 2021-06-15 23:00

Item: 2/2
	Dictionary ID: 456
	Item Key: baz
	Item Value: qux
	Created (UTC): 2021-06-15 23:00
	Last edited (UTC): 2021-06-15 23:00
	Deleted (UTC): 2021-06-15 23:00
`) + "\n\n"

func createDictionaryItemOK(_ context.Context, i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      i.ItemKey,
		ItemValue:    i.ItemValue,
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func updateDictionaryItemOK(_ context.Context, i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      fastly.ToPointer(i.ItemKey),
		ItemValue:    fastly.ToPointer(i.ItemValue),
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func deleteDictionaryItemOK(_ context.Context, _ *fastly.DeleteDictionaryItemInput) error {
	return nil
}

var dictionaryItemBatchModifyInputOK = `
{
	"items": [
		{
		  "op": "create",
		  "item_key": "some_key",
		  "item_value": "new_value"
		},
		{
		  "op": "update",
		  "item_key": "some_key",
		  "item_value": "new_value"
		},
		{
		  "op": "upsert",
		  "item_key": "some_key",
		  "item_value": "new_value"
		},
		{
		  "op": "delete",
		  "item_key": "some_key"
		}
	]
}`

func batchModifyDictionaryItemsOK(_ context.Context, _ *fastly.BatchModifyDictionaryItemsInput) error {
	return nil
}

func batchModifyDictionaryItemsError(_ context.Context, _ *fastly.BatchModifyDictionaryItemsInput) error {
	return errTest
}

var errTest = errors.New("an expected error occurred")
