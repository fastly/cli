package whoami_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/whoami"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
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
			name:      "no token",
			args:      args("whoami"),
			client:    verifyClient(basicResponse),
			wantError: "no token provided",
		},
		{
			name:       "basic response",
			args:       args("--token=x whoami"),
			client:     verifyClient(basicResponse),
			wantOutput: basicOutput,
		},
		{
			name:       "basic response verbose",
			args:       args("--token=x whoami -v"),
			client:     verifyClient(basicResponse),
			wantOutput: basicOutputVerbose,
		},
		{
			name:      "500 from API",
			args:      args("--token=x whoami"),
			client:    codeClient{code: http.StatusInternalServerError},
			wantError: "error from API: 500 Internal Server Error",
		},
		{
			name:      "local error",
			args:      args("--token=x whoami"),
			client:    errorClient{err: errors.New("some network failure")},
			wantError: "error executing API request: some network failure",
		},
		{
			name:   "alternative endpoint from flag",
			args:   args("--token=x whoami --endpoint=https://staging.fastly.com -v"),
			client: verifyClient(basicResponse),
			wantOutput: strings.ReplaceAll(basicOutputVerbose,
				"Fastly API endpoint: https://api.fastly.com",
				"Fastly API endpoint: https://staging.fastly.com",
			),
		},
		{
			name:   "alternative endpoint from environment",
			args:   args("--token=x whoami -v"),
			env:    config.Environment{Endpoint: "https://alternative.example.com"},
			client: verifyClient(basicResponse),
			wantOutput: strings.ReplaceAll(basicOutputVerbose,
				"Fastly API endpoint: https://api.fastly.com",
				fmt.Sprintf("Fastly API endpoint (via %s): https://alternative.example.com", env.Endpoint),
			),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.args, &stdout)
			opts.Env = testcase.env
			opts.HTTPClient = testcase.client
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.wantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.wantOutput)
		})
	}
}

type verifyClient whoami.VerifyResponse

func (c verifyClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	err := json.NewEncoder(rec).Encode(whoami.VerifyResponse(c))
	if err != nil {
		return nil, fmt.Errorf("failed to encode response into json: %w", err)
	}
	return rec.Result(), nil
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

var basicResponse = whoami.VerifyResponse{
	Customer: whoami.Customer{
		ID:   "abc",
		Name: "Computer Company",
	},
	User: whoami.User{
		ID:    "123",
		Name:  "Alice Programmer",
		Login: "alice@example.com",
	},
	Services: map[string]string{
		"1xxaa": "First service",
		"2baba": "Second service",
	},
	Token: whoami.Token{
		ID:        "abcdefg",
		Name:      "Token name",
		CreatedAt: "2019-01-01T12:00:00Z",
		// no ExpiresAt
		Scope: "global",
	},
}

var basicOutput = "Alice Programmer <alice@example.com>\n"

var basicOutputVerbose = strings.TrimSpace(`
Fastly API token provided via --token
Fastly API endpoint: https://api.fastly.com

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
