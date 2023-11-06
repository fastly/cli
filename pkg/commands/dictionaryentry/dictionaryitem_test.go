package dictionaryentry_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestDictionaryItemDescribe(t *testing.T) {
	args := testutil.Args
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
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.APIClient = mock.APIClient(testcase.api)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, stdout.String())
		})
	}
}

type mockDictionaryItemPaginator struct {
	count         int
	maxPages      int
	numOfPages    int
	requestedPage int
	returnErr     bool
}

func (p *mockDictionaryItemPaginator) HasNext() bool {
	if p.count > p.maxPages {
		return false
	}
	p.count++
	return true
}

func (p mockDictionaryItemPaginator) Remaining() int {
	return 1
}

func (p *mockDictionaryItemPaginator) GetNext() (di []*fastly.DictionaryItem, err error) {
	if p.returnErr {
		err = testutil.Err
	}
	pageOne := fastly.DictionaryItem{
		ServiceID:    "123",
		DictionaryID: "456",
		ItemKey:      "foo",
		ItemValue:    "bar",
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}
	pageTwo := fastly.DictionaryItem{
		ServiceID:    "123",
		DictionaryID: "456",
		ItemKey:      "baz",
		ItemValue:    "bear",
		CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		DeletedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:06:08Z"),
	}
	if p.count == 1 {
		di = append(di, &pageOne)
	}
	if p.count == 2 {
		di = append(di, &pageTwo)
	}
	if p.requestedPage > 0 && p.numOfPages == 1 {
		p.count = p.maxPages + 1 // forces only one result to be displayed
	}
	return di, err
}

func TestDictionaryItemsList(t *testing.T) {
	args := testutil.Args
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
				NewListDictionaryItemsPaginatorFn: func(i *fastly.ListDictionaryItemsInput) fastly.PaginatorDictionaryItems {
					return &mockDictionaryItemPaginator{returnErr: true}
				},
			},
			args:      args("dictionary-entry list --service-id 123 --dictionary-id 456"),
			wantError: testutil.Err.Error(),
		},
		// NOTE: Our mock paginator defines two dictionary items, and so even when
		// setting --per-page 1 we expect the final output to display both items.
		{
			api: mock.API{
				NewListDictionaryItemsPaginatorFn: func(i *fastly.ListDictionaryItemsInput) fastly.PaginatorDictionaryItems {
					return &mockDictionaryItemPaginator{numOfPages: i.PerPage, maxPages: 2}
				},
			},
			args:       args("dictionary-entry list --service-id 123 --dictionary-id 456 --per-page 1"),
			wantOutput: listDictionaryItemsOutput,
		},
		// In the following test, we set --page 1 and as there's only one record
		// displayed per page we expect only the first record to be displayed.
		{
			api: mock.API{
				NewListDictionaryItemsPaginatorFn: func(i *fastly.ListDictionaryItemsInput) fastly.PaginatorDictionaryItems {
					return &mockDictionaryItemPaginator{count: i.Page - 1, requestedPage: i.Page, numOfPages: i.PerPage, maxPages: 2}
				},
			},
			args:       args("dictionary-entry list --service-id 123 --dictionary-id 456 --page 1 --per-page 1"),
			wantOutput: listDictionaryItemsPageOneOutput,
		},
		// In the following test, we set --page 2 and as there's only one record
		// displayed per page we expect only the second record to be displayed.
		{
			api: mock.API{
				NewListDictionaryItemsPaginatorFn: func(i *fastly.ListDictionaryItemsInput) fastly.PaginatorDictionaryItems {
					return &mockDictionaryItemPaginator{count: i.Page - 1, requestedPage: i.Page, numOfPages: i.PerPage, maxPages: 2}
				},
			},
			args:       args("dictionary-entry list --service-id 123 --dictionary-id 456 --page 2 --per-page 1"),
			wantOutput: listDictionaryItemsPageTwoOutput,
		},
	}
	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
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

var updateDictionaryItemOutput = `SUCCESS: Updated dictionary item (service 123)

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

var listDictionaryItemsPageOneOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Item: 1/1
	Dictionary ID: 456
	Item Key: foo
	Item Value: bar
	Created (UTC): 2001-02-03 04:05
	Last edited (UTC): 2001-02-03 04:05
`) + "\n\n"

var listDictionaryItemsPageTwoOutput = "\n" + strings.TrimSpace(`
Service ID: 123
Item: 1/1
	Dictionary ID: 456
	Item Key: baz
	Item Value: bear
	Created (UTC): 2001-02-03 04:05
	Last edited (UTC): 2001-02-03 04:05
	Deleted (UTC): 2001-02-03 04:06
`) + "\n\n"

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
