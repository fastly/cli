package whoami_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestWhoami(t *testing.T) {
	args := testutil.Args
	for _, testcase := range []struct {
		name       string
		args       []string
		env        config.Environment
		client     api.HTTPClient
		wantError  string
		wantOutput string
	}{
		{
			name:       "basic response",
			args:       args("whoami"),
			client:     testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
			wantOutput: basicOutput,
		},
		{
			name:       "basic response verbose",
			args:       args("whoami -v"),
			client:     testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
			wantOutput: basicOutputVerbose,
		},
		{
			name:      "500 from API",
			args:      args("whoami"),
			client:    codeClient{code: http.StatusInternalServerError},
			wantError: "error executing API request: error response",
		},
		{
			name:      "local error",
			args:      args("whoami"),
			client:    errorClient{err: errors.New("some network failure")},
			wantError: "error executing API request: some network failure",
		},
		{
			name:   "alternative endpoint from flag",
			args:   args("whoami --endpoint=https://staging.fastly.com -v"),
			client: testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
			wantOutput: strings.ReplaceAll(basicOutputVerbose,
				"Fastly API endpoint: https://api.fastly.com",
				"Fastly API endpoint (via --endpoint): https://staging.fastly.com",
			),
		},
		{
			name:   "alternative endpoint from environment",
			args:   args("whoami -v"),
			env:    config.Environment{Endpoint: "https://alternative.example.com"},
			client: testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
			wantOutput: strings.ReplaceAll(basicOutputVerbose,
				"Fastly API endpoint: https://api.fastly.com",
				fmt.Sprintf("Fastly API endpoint (via %s): https://alternative.example.com", env.Endpoint),
			),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.MockGlobalData(testcase.args, &stdout)
			opts.Env = testcase.env
			opts.HTTPClient = testcase.client
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err := app.Run(testcase.args, nil)
			opts.Config = config.File{}
			t.Log(stdout.String())
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

type codeClient struct {
	code int
}

func (c codeClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	rec.WriteHeader(c.code)
	return rec.Result(), nil
}

type errorClient struct {
	err error
}

func (c errorClient) Do(*http.Request) (*http.Response, error) {
	return nil, c.err
}

var basicOutput = "Alice Programmer <alice@example.com>\n"

var basicOutputVerbose = strings.TrimSpace(`
Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Customer ID: abc
Customer name: Computer Company
User ID: 123
User name: Alice Programmer
User login: alice@example.com
Token ID: abcdefg
Token name: Token name
Token created at: 2019-01-01T12:00:00Z
Token scope: global
Service count: 2
	First service (1xxaa)
	Second service (2baba)
`) + "\n"
