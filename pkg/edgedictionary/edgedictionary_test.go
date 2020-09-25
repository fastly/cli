package edgedictionary_test

import (
	"bytes"
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
