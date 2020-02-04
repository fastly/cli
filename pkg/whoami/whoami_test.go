package whoami_test

import (
	"bytes"
	"encoding/json"
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
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/cli/pkg/whoami"
)

func TestWhoami(t *testing.T) {
	for _, testcase := range []struct {
		name       string
		args       []string
		env        config.Environment
		file       config.File
		client     api.HTTPClient
		wantError  string
		wantOutput string
	}{
		{
			name:      "no token",
			args:      []string{"whoami"},
			client:    verifyClient(basicResponse),
			wantError: "no token provided",
		},
		{
			name:       "basic response",
			args:       []string{"--token=x", "whoami"},
			client:     verifyClient(basicResponse),
			wantOutput: basicOutput,
		},
		{
			name:       "basic response verbose",
			args:       []string{"--token=x", "whoami", "-v"},
			client:     verifyClient(basicResponse),
			wantOutput: basicOutputVerbose,
		},
		{
			name:      "500 from API",
			args:      []string{"--token=x", "whoami"},
			client:    codeClient{code: http.StatusInternalServerError},
			wantError: "error from API: 500 Internal Server Error",
		},
		{
			name:      "local error",
			args:      []string{"--token=x", "whoami"},
			client:    errorClient{err: errors.New("some network failure")},
			wantError: "error executing API request: some network failure",
		},
		{
			name:   "alternative endpoint from flag",
			args:   []string{"--token=x", "whoami", "--endpoint=https://staging.fastly.com", "-v"},
			client: verifyClient(basicResponse),
			wantOutput: strings.ReplaceAll(basicOutputVerbose,
				"Fastly API endpoint: https://api.fastly.com",
				"Fastly API endpoint: https://staging.fastly.com",
			),
		},
		{
			name:   "alternative endpoint from environment",
			args:   []string{"--token=x", "whoami", "-v"},
			env:    config.Environment{Endpoint: "https://alternative.example.com"},
			client: verifyClient(basicResponse),
			wantOutput: strings.ReplaceAll(basicOutputVerbose,
				"Fastly API endpoint: https://api.fastly.com",
				fmt.Sprintf("Fastly API endpoint (via %s): https://alternative.example.com", config.EnvVarEndpoint),
			),
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = testcase.env
				file                            = testcase.file
				configFileName                  = "/dev/null"
				clientFactory                   = mock.APIClient(mock.API{})
				httpClient                      = testcase.client
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

type verifyClient whoami.VerifyResponse

func (c verifyClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	json.NewEncoder(rec).Encode(whoami.VerifyResponse(c))
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
