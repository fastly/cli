package edgedictionary_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/go-fastly/fastly"
)

func TestDictionaryDescribe(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionary", "describe", "--version", "1", "--service-id", "123"},
			api:       mock.API{GetDictionaryFn: describeDictionaryOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       []string{"dictionary", "describe", "--version", "1", "--service-id", "123", "--name", "dict-1"},
			api:        mock.API{GetDictionaryFn: describeDictionaryOK},
			wantOutput: describeDictionaryOutput,
		},
		{
			args:       []string{"dictionary", "describe", "--version", "1", "--service-id", "123", "--name", "dict-1"},
			api:        mock.API{GetDictionaryFn: describeDictionaryOKDeleted},
			wantOutput: describeDictionaryOutputDeleted,
		},
		{
			args: []string{"dictionary", "describe", "--version", "1", "--service-id", "123", "--name", "dict-1", "--verbose"},
			api: mock.API{
				GetDictionaryFn:       describeDictionaryOK,
				GetDictionaryInfoFn:   getDictionaryInfoOK,
				ListDictionaryItemsFn: listDictionaryItemsOK,
			},
			wantOutput: describeDictionaryOutputVerbose,
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
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

func TestDictionaryCreate(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionary", "create", "--version", "1", "--service-id", "123"},
			api:       mock.API{CreateDictionaryFn: createDictionaryOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       []string{"dictionary", "create", "--version", "1", "--service-id", "123", "--name", "denylist"},
			api:        mock.API{CreateDictionaryFn: createDictionaryOK},
			wantOutput: createDictionaryOutput,
		},
		{
			args:      []string{"dictionary", "create", "--version", "1", "--service-id", "123", "--name", "denylist"},
			api:       mock.API{CreateDictionaryFn: createDictionaryDuplicate},
			wantError: "Duplicate record",
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
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

func TestDeleteDictionary(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:      []string{"dictionary", "delete", "--service-id", "123", "--version", "1"},
			api:       mock.API{DeleteDictionaryFn: deleteDictionaryOK},
			wantError: "error parsing arguments: required flag --name not provided",
		},
		{
			args:       []string{"dictionary", "delete", "--service-id", "123", "--version", "1", "--name", "allowlist"},
			api:        mock.API{DeleteDictionaryFn: deleteDictionaryOK},
			wantOutput: deleteDictionaryOutput,
		},
		{
			args:      []string{"dictionary", "delete", "--service-id", "123", "--version", "1", "--name", "allowlist"},
			api:       mock.API{DeleteDictionaryFn: deleteDictionaryError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                           = testcase.args
				env                            = config.Environment{}
				file                           = config.File{}
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

func describeDictionaryOK(i *fastly.GetDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID: i.Service,
		Version:   i.Version,
		Name:      i.Name,
		CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly: false,
		ID:        "456",
		UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

func describeDictionaryOKDeleted(i *fastly.GetDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID: i.Service,
		Version:   i.Version,
		Name:      i.Name,
		CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly: false,
		ID:        "456",
		UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		DeletedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:08Z"),
	}, nil
}

func createDictionaryOK(i *fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return &fastly.Dictionary{
		ServiceID: i.Service,
		Version:   i.Version,
		Name:      i.Name,
		CreatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
		WriteOnly: false,
		ID:        "456",
		UpdatedAt: testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
	}, nil
}

// These two are responses from other packages that will need to update if
// the responses ever change.
func getDictionaryInfoOK(i *fastly.GetDictionaryInfoInput) (*fastly.DictionaryInfo, error) {
	return &fastly.DictionaryInfo{
		ItemCount:   2,
		LastUpdated: testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		Digest:      "digest_hash",
	}, nil
}

func listDictionaryItemsOK(i *fastly.ListDictionaryItemsInput) ([]*fastly.DictionaryItem, error) {
	return []*fastly.DictionaryItem{
		&fastly.DictionaryItem{
			ServiceID:    i.Service,
			DictionaryID: "456",
			ItemKey:      "foo",
			ItemValue:    "bar",
			CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
		},
		&fastly.DictionaryItem{
			ServiceID:    i.Service,
			DictionaryID: "456",
			ItemKey:      "baz",
			ItemValue:    "bear",
			CreatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:06Z"),
			UpdatedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:05:07Z"),
			DeletedAt:    testutil.MustParseTimeRFC3339("2001-02-03T04:06:08Z"),
		},
	}, nil
}

func createDictionaryDuplicate(*fastly.CreateDictionaryInput) (*fastly.Dictionary, error) {
	return nil, errors.New("Duplicate record")
}

func deleteDictionaryOK(*fastly.DeleteDictionaryInput) error {
	return nil
}

func deleteDictionaryError(*fastly.DeleteDictionaryInput) error {
	return errTest
}

var errTest = errors.New("an expected error ocurred")

var createDictionaryOutput = "\nSUCCESS: Created dictionary denylist (service 123 version 1)\n"
var deleteDictionaryOutput = "\nSUCCESS: Deleted dictionary allowlist (service 123 version 1)\n"

var describeDictionaryOutput = strings.TrimSpace(`
Service ID: 123
Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
`) + "\n"

var describeDictionaryOutputDeleted = strings.TrimSpace(`
Service ID: 123
Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
Deleted (UTC): 2001-02-03 04:05
`) + "\n"

var describeDictionaryOutputVerbose = strings.TrimSpace(`
Fastly API token not provided
Fastly API endpoint: https://api.fastly.com
Service ID: 123
Version: 1
ID: 456
Name: dict-1
Write Only: false
Created (UTC): 2001-02-03 04:05
Last edited (UTC): 2001-02-03 04:05
Digest: digest_hash
Item Count: 2
Item 1/2:
	Item Key: foo
	Item Value: bar
Item 2/2:
	Item Key: baz
	Item Value: bear
`) + "\n"
