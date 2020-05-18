package stats_test

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

func TestRegions(t *testing.T) {
	for _, testcase := range []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       []string{"stats", "regions"},
			api:        mock.API{GetRegionsFn: getRegionsOK},
			wantOutput: "foo\nbar\nbaz\n",
		},
		{
			args:      []string{"stats", "regions"},
			api:       mock.API{GetRegionsFn: getRegionsError},
			wantError: errTest.Error(),
		},
	} {
		t.Run(strings.Join(testcase.args, " "), func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(testcase.api)
				httpClient                      = http.DefaultClient
				versioner      update.Versioner = nil
				in             io.Reader        = nil
				out            bytes.Buffer
			)
			err := app.Run(args, env, file, configFileName, clientFactory, httpClient, versioner, in, &out)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, out.String(), testcase.wantOutput)
		})
	}
}

func getRegionsOK() (*fastly.RegionsResponse, error) {
	return &fastly.RegionsResponse{
		Data: []string{"foo", "bar", "baz"},
	}, nil
}

var errTest = errors.New("fixture error")

func getRegionsError() (*fastly.RegionsResponse, error) {
	return nil, errTest
}
