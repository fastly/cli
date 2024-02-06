package stats_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestRegions(t *testing.T) {
	args := testutil.Args
	scenarios := []struct {
		args       []string
		api        mock.API
		wantError  string
		wantOutput string
	}{
		{
			args:       args("stats regions"),
			api:        mock.API{GetRegionsFn: getRegionsOK},
			wantOutput: "foo\nbar\nbaz\n",
		},
		{
			args:      args("stats regions"),
			api:       mock.API{GetRegionsFn: getRegionsError},
			wantError: errTest.Error(),
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
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
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
