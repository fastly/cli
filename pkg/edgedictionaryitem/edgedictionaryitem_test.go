package edgedictionaryitem_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestDictionaryItemDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionaryitem", "describe", "--service-id", "123", "--key", "foo"},
			api:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:      []string{"dictionaryitem", "describe", "--service-id", "123", "--dictionary-id", "456"},
			api:       mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantError: "error parsing arguments: required flag --key not provided",
		},
		{
			args:       []string{"dictionaryitem", "describe", "--service-id", "123", "--dictionary-id", "456", "--key", "foo"},
			api:        mock.API{GetDictionaryItemFn: describeDictionaryItemOK},
			wantOutput: describeDictionaryItemOutput,
		},
		{
			args:       []string{"dictionaryitem", "describe", "--service-id", "123", "--dictionary-id", "456", "--key", "foo-deleted"},
			api:        mock.API{GetDictionaryItemFn: describeDictionaryItemOKDeleted},
			wantOutput: describeDictionaryItemOutputDeleted,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestDictionaryItemsList(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionaryitem", "list", "--service-id", "123"},
			api:       mock.API{ListDictionaryItemsFn: listDictionaryItemsOK},
			wantError: "error parsing arguments: required flag --dictionary-id not provided",
		},
		{
			args:       []string{"dictionaryitem", "list", "--service-id", "123", "--dictionary-id", "456"},
			api:        mock.API{ListDictionaryItemsFn: listDictionaryItemsOK},
			wantOutput: listDictionaryItemsOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestDictionaryItemCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionaryitem", "create", "--service-id", "123"},
			api:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      []string{"dictionaryitem", "create", "--service-id", "123", "--dictionary-id", "456"},
			api:       mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       []string{"dictionaryitem", "create", "--service-id", "123", "--dictionary-id", "456", "--key", "foo", "--value", "bar"},
			api:        mock.API{CreateDictionaryItemFn: createDictionaryItemOK},
			wantOutput: "\nSUCCESS: Created dictionary item foo (service 123, dictionary 456)\n",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestDictionaryItemUpdate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionaryitem", "update", "--service-id", "123"},
			api:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      []string{"dictionaryitem", "update", "--service-id", "123", "--dictionary-id", "456"},
			api:       mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       []string{"dictionaryitem", "update", "--service-id", "123", "--dictionary-id", "456", "--key", "foo", "--value", "bar"},
			api:        mock.API{UpdateDictionaryItemFn: updateDictionaryItemOK},
			wantOutput: describeDictionaryItemOutput,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestDictionaryItemDelete(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionaryitem", "delete", "--service-id", "123"},
			api:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      []string{"dictionaryitem", "delete", "--service-id", "123", "--dictionary-id", "456"},
			api:       mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:       []string{"dictionaryitem", "delete", "--service-id", "123", "--dictionary-id", "456", "--key", "foo"},
			api:        mock.API{DeleteDictionaryItemFn: deleteDictionaryItemOK},
			wantOutput: "\nSUCCESS: Deleted dictionary item foo (service 123, dicitonary 456)\n",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
		})
	}
}

func TestDictionaryItemBatchModify(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		fileData   string
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionaryitem", "batchmodify", "--service-id", "123"},
			wantError: "error parsing arguments: required flag ",
		},
		{
			args:      []string{"dictionaryitem", "batchmodify", "--service-id", "123", "--dictionary-id", "456"},
			wantError: "error parsing arguments: required flag --file not provided",
		},
		{
			fileData:  `{invalid": "json"}`,
			args:      []string{"dictionaryitem", "batchmodify", "--service-id", "123", "--dictionary-id", "456", "--file", "filePath"},
			wantError: "invalid character 'i' looking for beginning of object key string",
		},
		{
			fileData:  `{"valid": "json"}`,
			args:      []string{"dictionaryitem", "batchmodify", "--service-id", "123", "--dictionary-id", "456", "--file", "filePath"},
			wantError: "item key not found in file ",
		},
		{
			args:      []string{"dictionaryitem", "batchmodify", "--service-id", "123", "--dictionary-id", "456", "--file", "missingFile"},
			wantError: "open missingFile",
		},
		{
			fileData:  dictionaryItemBatchModifyInputOK,
			args:      []string{"dictionaryitem", "batchmodify", "--service-id", "123", "--dictionary-id", "456", "--file", "filePath"},
			api:       mock.API{BatchModifyDictionaryItemsFn: batchModifyDictionaryItemsError},
			wantError: errTest.Error(),
		},
		{
			fileData:   dictionaryItemBatchModifyInputOK,
			args:       []string{"dictionaryitem", "batchmodify", "--service-id", "123", "--dictionary-id", "456", "--file", "filePath"},
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

			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.ConfigFile{}
				appConfigFile                  = "/dev/null"
				clientFactory                  = mock.APIClient(testcase.api)
				httpClient                     = http.DefaultClient
				versioner     update.Versioner = nil
				in            io.Reader        = nil
				out           bytes.Buffer
			)
			err := app.Run(args, env, file, appConfigFile, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertString(t, testcase.wantOutput, out.String())
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

var describeDictionaryItemOutput = strings.TrimSpace(`
Service ID: 123
Dictionary ID: 456
Item Key: foo
Item Value: bar
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`) + "\n"

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

var describeDictionaryItemOutputDeleted = strings.TrimSpace(`
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

var listDictionaryItemsOutput = strings.TrimSpace(`
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
