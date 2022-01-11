package dictionaryitem_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v5/fastly"
)

func TestDictionaryItemDescribe(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-item describe --service-id 123 --key foo"),
			api:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:      args("dictionary-item describe --service-id 123 --dictionary-id 456"),
			api:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantError: "error parsing arguments: required flag --key not provided",
		},
		{
			args:       args("dictionary-item describe --service-id 123 --dictionary-id 456 --key foo"),
			api:        mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantOutput: describeDictionaryItemOutput,
		},
		{
			args:       args("dictionary-item describe --service-id 123 --dictionary-id 456 --key foo-deleted"),
			api:        mock.API{GetDictionaryItemFn: describeDictionaryItemOKDeleted},
			wantOutput: describeDictionaryItemOutputDeleted,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemsList(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-item list --service-id 123"),
			api:       mock.API{ListDictionaryItemsFn: listDictionaryItemsOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:       args("dictionary-item list --service-id 123 --dictionary-id 456"),
			api:        mock.API{ListDictionaryItemsFn: listDictionaryItemsOK},
			wantOutput: listDictionaryItemsOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemCreate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-item create --service-id 123"),
			api:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      args("dictionary-item create --service-id 123 --dictionary-id 456"),
			api:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       args("dictionary-item create --service-id 123 --dictionary-id 456 --key foo --value bar"),
			api:        mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantOutput: "\nSUCCESS: Created dictionary item foo (service 123, dictionary 456)\n",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemUpdate(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		fileData   string
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-item update --service-id 123"),
			api:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:      args("dictionary-item update --service-id 123 --dictionary-id 456"),
			api:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantError: "an empty value is not allowed for either the '--key' or '--value' flags",
		},
		{
			args:       args("dictionary-item update --service-id 123 --dictionary-id 456 --key foo --value bar"),
			api:        mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantOutput: updateDictionaryItemOutput,
		},
		{
			args:      args("dictionary-item update --service-id 123 --dictionary-id 456 --file filePath"),
			fileData:  `{invalid": "json"}`,
			wantError: "invalid character 'i' looking for beginning of object key string",
		},
		// NOTE: We don't specify the full error value in the wantError field
		// because this would cause an error on different OS'. For example, Unix
		// systems report 'no such file or directory', while Windows will report
		// 'The system cannot find the file specified'.
		{
			args:      args("dictionary-item update --service-id 123 --dictionary-id 456 --file missingPath"),
			wantError: "open missingPath:",
		},
		{
			args:      args("dictionary-item update --service-id 123 --dictionary-id 456 --file filePath"),
			fileData:  dictionaryItemBatchModifyInputOK,
			api:       mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsError},
			wantError: errTest.Error(),
		},
		{
			args:       args("dictionary-item update --service-id 123 --dictionary-id 456 --file filePath"),
			fileData:   dictionaryItemBatchModifyInputOK,
			api:        mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsOK},
			wantOutput: "\nSUCCESS: Made 4 modifications of Dictionary 456 on service 123\n",
		},
	} {
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
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func TestDictionaryItemDelete(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      args("dictionary-item delete --service-id 123"),
			api:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      args("dictionary-item delete --service-id 123 --dictionary-id 456"),
			api:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       args("dictionary-item delete --service-id 123 --dictionary-id 456 --key foo"),
			api:        mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantOutput: "\nSUCCESS: Deleted dictionary item foo (service 123, dicitonary 456)\n",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

func describeDictionaryItemOK(i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    i.ServiceID,
		DictionaryID: i.DictionaryID,
		ItemKey:      i.ItemKey,
		ItemValue:    "bar",
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

var updateDictionaryItemOutput = `
SUCCESS: Updated dictionary item (service 123)

Dictionary ID: 456
Item Key: foo
Item Value: bar
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`

func describeDictionaryItemOKDeleted(i *fastly.GetDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    i.ServiceID,
		DictionaryID: i.DictionaryID,
		ItemKey:      i.ItemKey,
		ItemValue:    "bar",
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

func listDictionaryItemsOK(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return []*fastly.DictionaryItem{
		{
			ServiceID:    i.ServiceID,
			DictionaryID: i.DictionaryID,
			ItemKey:      "foo",
			ItemValue:    "bar",
			CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		},
		{
			ServiceID:    i.ServiceID,
			DictionaryID: i.DictionaryID,
			ItemKey:      "baz",
			ItemValue:    "bear",
			CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
			DeletedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:06:08Z"),
		},
	}, nil
}

var listDictionaryItemsOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Item: 1/2
	Dictionary ID: 456
	Item Key: foo
	Item Value: bar
	Created (UTC): 2001-02-03 04:05
	Last edited (UTC): 2001-02-03 04:05

Item: 2/2
	Dictionary ID: 456
	Item Key: baz
	Item Value: bear
	Created (UTC): 2001-02-03 04:05
	Last edited (UTC): 2001-02-03 04:05
	Deleted (UTC): 2001-02-03 04:06
`) + "\n\n"

func createDictionaryItemOK(i *fastly.CreateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    i.ServiceID,
		DictionaryID: i.DictionaryID,
		ItemKey:      i.ItemKey,
		ItemValue:    i.ItemValue,
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func updateDictionaryItemOK(i *fastly.UpdateDictionaryItemInput) (*fastly.DictionaryItem, error) {
	return &fastly.DictionaryItem{
		ServiceID:    i.ServiceID,
		DictionaryID: i.DictionaryID,
		ItemKey:      i.ItemKey,
		ItemValue:    i.ItemValue,
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func deleteDictionaryItemOK(i *fastly.DeleteDictionaryItemInput) error {
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

func batchModifyDictionaryItemsOK(i *fastly.BatchModifyDictionaryItemsInput) error {
	return nil
}

func batchModifyDictionaryItemsError(i *fastly.BatchModifyDictionaryItemsInput) error {
	return errTest
}

var errTest = errors.New("an expected error ocurred")
