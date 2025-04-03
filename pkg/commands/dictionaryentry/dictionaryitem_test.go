package dictionaryentry_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDictionaryItemDescribe(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-entry describe --service-id 123 --key foo"),
			api:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:      args("dictionary-entry describe --service-id 123 --dictionary-id 456"),
			api:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantError: "error parsing arguments: required flag --key not provided",
		},
		{
			args:       args("dictionary-entry describe --service-id 123 --dictionary-id 456 --key foo"),
			api:        mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantOutput: describeDictionaryItemOutput,
		},
		{
			args:       args("dictionary-entry describe --service-id 123 --dictionary-id 456 --key foo-deleted"),
			api:        mock.API{GetDictionaryItemFn: describeDictionaryItemOKDeleted},
			wantOutput: describeDictionaryItemOutputDeleted,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemsList(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-entry list --service-id 123"),
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:      args("dictionary-entry list --dictionary-id 456"),
			wantError: "error reading service: no service ID found",
		},
		{
			api: mock.API{
				GetDictionaryItemsFn: func(i *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem] {
					return fastly.NewPaginator[fastly.DictionaryItem](&mock.HTTPClient{
						Errors: []error{
							testutil.Err,
						},
						Responses: []*http.Response{nil},
					}, fastly.ListOpts{}, "/example")
				},
			},
			args:      args("dictionary-entry list --service-id 123 --dictionary-id 456"),
			wantError: testutil.Err.Error(),
		},
		{
			api: mock.API{
				GetDictionaryItemsFn: func(i *fastly.GetDictionaryItemsInput) *fastly.ListPaginator[fastly.DictionaryItem] {
					return fastly.NewPaginator[fastly.DictionaryItem](&mock.HTTPClient{
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
			args:       args("dictionary-entry list --service-id 123 --dictionary-id 456 --per-page 1"),
			wantOutput: listDictionaryItemsOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemCreate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-entry create --service-id 123"),
			api:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      args("dictionary-entry create --service-id 123 --dictionary-id 456"),
			api:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       args("dictionary-entry create --service-id 123 --dictionary-id 456 --key foo --value bar"),
			api:        mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantOutput: "SUCCESS: Created dictionary item foo (service 123, dictionary 456)\n",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemUpdate(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		fileData   string
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-entry update --service-id 123"),
			api:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:      args("dictionary-entry update --service-id 123 --dictionary-id 456"),
			api:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantError: "an empty value is not allowed for either the '--key' or '--value' flags",
		},
		{
			args:       args("dictionary-entry update --service-id 123 --dictionary-id 456 --key foo --value bar"),
			api:        mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantOutput: updateDictionaryItemOutput,
		},
		{
			args:      args("dictionary-entry update --service-id 123 --dictionary-id 456 --file filePath"),
			fileData:  `{invalid": "json"}`,
			wantError: "invalid character 'i' looking for beginning of object key string",
		},
		// NOTE: We don't specify the full error value in the wantError field
		// because this would cause an error on different OS'. For example, Unix
		// systems report 'no such file or directory', while Windows will report
		// 'The system cannot find the file specified'.
		{
			args:      args("dictionary-entry update --service-id 123 --dictionary-id 456 --file missingPath"),
			wantError: "open missingPath:",
		},
		{
			args:      args("dictionary-entry update --service-id 123 --dictionary-id 456 --file filePath"),
			fileData:  dictionaryItemBatchModifyInputOK,
			api:       mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsError},
			wantError: errTest.Error(),
		},
		{
			args:       args("dictionary-entry update --service-id 123 --dictionary-id 456 --file filePath"),
			fileData:   dictionaryItemBatchModifyInputOK,
			api:        mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsOK},
			wantOutput: "SUCCESS: Made 4 modifications of Dictionary 456 on service 123\n",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var filePath string
			if testcase.fileData != "" {
				filePath = testutil.MakeTempFile(t, testcase.fileData)
				defer os.RemoveAll(filePath)
			}

			// Insert temp file path into args when "filePath" is present as placeholder
			for i, v := range testcase.args {
				if v == "filePath" {
					testcase.args[i] = filePath
				}
			}

			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemDelete(t *testing.T) {
	args := testutil.SplitArgs
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-entry delete --service-id 123"),
			api:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      args("dictionary-entry delete --service-id 123 --dictionary-id 456"),
			api:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       args("dictionary-entry delete --service-id 123 --dictionary-id 456 --key foo"),
			api:        mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantOutput: "SUCCESS: Deleted dictionary item foo (service 123, dictionary 456)\n",
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.api)
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func describeDictionaryItemOK(i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
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

func describeDictionaryItemOKDeleted(i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
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

func createDictionaryItemOK(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      i.ItemKey,
		ItemValue:    i.ItemValue,
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func updateDictionaryItemOK(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    fastly.ToPointer(i.ServiceID),
		DictionaryID: fastly.ToPointer(i.DictionaryID),
		ItemKey:      fastly.ToPointer(i.ItemKey),
		ItemValue:    fastly.ToPointer(i.ItemValue),
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func deleteDictionaryItemOK(_ *fastly.DeleteDictionaryItemInput) error {
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

func batchModifyDictionaryItemsOK(_ *fastly.BatchModifyDictionaryItemsInput) error {
	return nil
}

func batchModifyDictionaryItemsError(_ *fastly.BatchModifyDictionaryItemsInput) error {
	return errTest
}

var errTest = errors.New("an expected error occurred")
