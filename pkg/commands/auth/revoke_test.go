package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

func TestAuthRevoke(t *testing.T) {
	deleteTokenSelfOK := func(_ context.Context) error { return nil }
	deleteTokenOK := func(_ context.Context, _ *fastly.DeleteTokenInput) error { return nil }
	batchDeleteTokensOK := func(_ context.Context, _ *fastly.BatchDeleteTokensInput) error { return nil }

	deleteTokenSelf401 := func(_ context.Context) error {
		return &fastly.HTTPError{StatusCode: http.StatusUnauthorized}
	}
	deleteTokenSelf500 := func(_ context.Context) error {
		return &fastly.HTTPError{StatusCode: http.StatusInternalServerError}
	}

	twoTokenConfig := &config.File{
		Auth: config.Auth{
			Default: "primary",
			Tokens: config.AuthTokens{
				"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary", APITokenID: "id-primary"},
				"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary", APITokenID: "id-secondary"},
			},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "no flags provided",
			Args:      "revoke",
			WantError: "must provide one of",
		},
		{
			Name:      "multiple flags provided",
			Args:      "revoke --current --name foo",
			WantError: "only one of",
		},

		// --current
		{
			Name: "revoke current token",
			Args: "revoke --current --token tok-stored",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "mytoken",
					Tokens: config.AuthTokens{
						"mytoken": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-stored"},
					},
				},
			},
			Stdin:       []string{"y"},
			WantOutputs: []string{"Revoked current token", `Removed local token entry "mytoken"`},
		},
		{
			Name: "revoke current default declined is clean exit",
			Args: "revoke --current --token tok-stored",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "mytoken",
					Tokens: config.AuthTokens{
						"mytoken": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-stored"},
					},
				},
			},
			Stdin:          []string{"n"},
			WantOutput:     "current default token",
			DontWantOutput: "Revoked",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("mytoken") == nil {
					t.Error("expected token to still exist after decline")
				}
			},
		},
		{
			Name:           "revoke current skips prompt with --auto-yes",
			Args:           "revoke --current --token 123 -y",
			API:            &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			DontWantOutput: "Are you sure",
			WantOutput:     "Revoked current token",
		},
		{
			Name:           "revoke current with --token flag (unstored token)",
			Args:           "revoke --current --token raw-ephemeral-token",
			API:            &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			WantOutputs:    []string{"Revoked current token"},
			DontWantOutput: "Removed local",
		},

		// --name
		{
			Name: "revoke by name success",
			Args: "revoke --name secondary",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "primary",
					Tokens: config.AuthTokens{
						"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary"},
						"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary"},
					},
				},
			},
			WantOutputs: []string{`Revoked token "secondary"`, `Removed local token entry "secondary"`},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("secondary") != nil {
					t.Error("expected secondary token to be removed")
				}
				if opts.Config.GetAuthToken("primary") == nil {
					t.Error("expected primary token to still exist")
				}
			},
		},
		{
			Name:      "revoke by name not found",
			Args:      "revoke --name ghost",
			WantError: `token "ghost" not found`,
		},
		{
			Name: "revoke by name remote 401 still cleans up locally",
			Args: "revoke --name secondary",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelf401},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "primary",
					Tokens: config.AuthTokens{
						"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary"},
						"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary"},
					},
				},
			},
			WantOutputs: []string{"already revoked", `Removed local token entry "secondary"`},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("secondary") != nil {
					t.Error("expected secondary token to be removed after 401")
				}
			},
		},
		{
			Name: "revoke by name remote 5xx does not clean up locally",
			Args: "revoke --name secondary",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelf500},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "primary",
					Tokens: config.AuthTokens{
						"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary"},
						"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary"},
					},
				},
			},
			WantError:      "500",
			DontWantOutput: "Removed",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("secondary") == nil {
					t.Error("expected secondary token to still exist after 5xx")
				}
			},
		},
		{
			Name: "revoke by name default token reassigns",
			Args: "revoke --name primary -y",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "primary",
					Tokens: config.AuthTokens{
						"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary"},
						"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary"},
					},
				},
			},
			WantOutputs: []string{`Removed local token entry "primary"`, "Default token reassigned"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.Auth.Default == "primary" {
					t.Error("expected default to no longer be primary")
				}
				if opts.Config.Auth.Default == "" {
					t.Error("expected default to be reassigned")
				}
			},
		},

		// --token-value
		{
			Name: "revoke by token value success with local match",
			Args: "revoke --token-value tok-secondary",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "primary",
					Tokens: config.AuthTokens{
						"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary"},
						"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary"},
					},
				},
			},
			WantOutputs: []string{"Revoked token", `Removed local token entry "secondary"`},
		},
		{
			Name:        "revoke by token value no local match",
			Args:        "revoke --token-value tok-unknown",
			API:         &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			WantOutputs: []string{"Revoked token", "No matching local token entry found"},
		},
		{
			Name: "revoke by token value from stdin",
			Args: "revoke --token-value=-",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "other",
					Tokens: config.AuthTokens{
						"other":  &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "other-tok"},
						"target": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-from-stdin"},
					},
				},
			},
			Stdin:       []string{"tok-from-stdin"},
			WantOutputs: []string{"Revoked token", `Removed local token entry "target"`},
		},
		{
			Name:            "revoke by token value rejects oversized stdin",
			Args:            "revoke --token-value=-",
			Stdin:           []string{strings.Repeat("x", 5000)},
			WantError:       "exceeds 4096 bytes",
			WantRemediation: "single token value",
		},
		{
			Name: "revoke by token value removes duplicate local entries",
			Args: "revoke --token-value shared-tok",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "alias1",
					Tokens: config.AuthTokens{
						"alias1": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "shared-tok"},
						"alias2": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "shared-tok"},
						"other":  &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "other-tok"},
					},
				},
			},
			Stdin: []string{"y"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("alias1") != nil {
					t.Error("expected alias1 to be removed")
				}
				if opts.Config.GetAuthToken("alias2") != nil {
					t.Error("expected alias2 to be removed")
				}
				if opts.Config.GetAuthToken("other") == nil {
					t.Error("expected other token to still exist")
				}
			},
		},

		{
			Name: "revoke by token value confirms when revoking default",
			Args: "revoke --token-value tok-default",
			API:  &mock.API{DeleteTokenSelfFn: deleteTokenSelfOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "mydefault",
					Tokens: config.AuthTokens{
						"mydefault": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-default"},
					},
				},
			},
			Stdin:          []string{"n"},
			WantOutput:     "current default token",
			DontWantOutput: "Revoked",
		},

		// --id
		{
			Name: "revoke by ID with local match",
			Args: "revoke --id id-secondary --token 123",
			API:  &mock.API{DeleteTokenFn: deleteTokenOK},
			ConfigFile: func() *config.File {
				c := *twoTokenConfig
				c.Auth.Tokens = make(config.AuthTokens)
				for k, v := range twoTokenConfig.Auth.Tokens {
					cp := *v
					c.Auth.Tokens[k] = &cp
				}
				return &c
			}(),
			WantOutputs: []string{"Revoked token 'id-secondary'", `Removed local token entry "secondary"`},
		},
		{
			Name: "revoke by ID no local match warns",
			Args: "revoke --id id-unknown --token 123",
			API:  &mock.API{DeleteTokenFn: deleteTokenOK},
			ConfigFile: func() *config.File {
				c := *twoTokenConfig
				c.Auth.Tokens = make(config.AuthTokens)
				for k, v := range twoTokenConfig.Auth.Tokens {
					cp := *v
					c.Auth.Tokens[k] = &cp
				}
				return &c
			}(),
			WantOutputs: []string{"Revoked token 'id-unknown'", "local cleanup skipped"},
		},
		{
			Name: "revoke by ID API 401 returns error without local cleanup",
			Args: "revoke --id some-id --token 123",
			API: &mock.API{
				DeleteTokenFn: func(_ context.Context, _ *fastly.DeleteTokenInput) error {
					return &fastly.HTTPError{StatusCode: http.StatusUnauthorized}
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "stored",
					Tokens: config.AuthTokens{
						"stored": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok", APITokenID: "some-id"},
					},
				},
			},
			WantError:      "401",
			DontWantOutput: "Removed",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("stored") == nil {
					t.Error("expected token to still exist after 401 on --id path")
				}
			},
		},
		{
			Name: "revoke by ID legacy token without APITokenID",
			Args: "revoke --id id-legacy --token 123",
			API:  &mock.API{DeleteTokenFn: deleteTokenOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "legacy",
					Tokens: config.AuthTokens{
						"legacy": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-legacy"},
					},
				},
			},
			WantOutputs: []string{"Revoked token 'id-legacy'", "local cleanup skipped"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("legacy") == nil {
					t.Error("expected legacy token to still exist (no APITokenID)")
				}
			},
		},

		// --file
		{
			Name: "revoke by file success",
			Args: fmt.Sprintf("revoke --file %s --token 123", writeTokenIDFile(t, "id-1\nid-2\n")),
			API:  &mock.API{BatchDeleteTokensFn: batchDeleteTokensOK},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "tok1",
					Tokens: config.AuthTokens{
						"tok1":  &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "t1", APITokenID: "id-1"},
						"tok2":  &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "t2", APITokenID: "id-2"},
						"other": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "t3", APITokenID: "id-other"},
					},
				},
			},
			WantOutputs: []string{"Revoked 2 token(s)", "Removed local token entry"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.GetAuthToken("tok1") != nil {
					t.Error("expected tok1 to be removed")
				}
				if opts.Config.GetAuthToken("tok2") != nil {
					t.Error("expected tok2 to be removed")
				}
				if opts.Config.GetAuthToken("other") == nil {
					t.Error("expected other to still exist")
				}
			},
		},
		{
			Name:            "revoke by file unreadable",
			Args:            "revoke --file /nonexistent/path/tokens.txt --token 123",
			WantError:       "failed to open",
			WantRemediation: "file path and permissions",
		},
		{
			Name:            "revoke by file empty",
			Args:            fmt.Sprintf("revoke --file %s --token 123", writeTokenIDFile(t, "\n\n")),
			WantError:       "contains no token IDs",
			WantRemediation: "one token ID per line",
		},

		// API client factory failure
		{
			Name: "API client factory failure on --name",
			Args: "revoke --name secondary",
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.APIClientFactory = func(_, _ string, _ bool) (api.Interface, error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "primary",
					Tokens: config.AuthTokens{
						"primary":   &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-primary"},
						"secondary": &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "tok-secondary"},
					},
				},
			},
			WantError:       "connection refused",
			WantRemediation: "network connection",
		},
	}

	testutil.RunCLIScenarios(t, []string{"auth"}, scenarios)
}

func writeTokenIDFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "token-ids.txt")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}
