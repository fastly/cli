package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"

	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

var (
	testTokenScope    = fastly.GlobalScope
	testTokenExpiry   = time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	testTokenSelfFull = func(_ context.Context) (*fastly.Token, error) {
		return &fastly.Token{
			TokenID: fastly.ToPointer("tok-id-123"),
			Name:    fastly.ToPointer("my-api-token"),
			Scope:   &testTokenScope,
		}, nil
	}
	testTokenSelfWithExpiry = func(_ context.Context) (*fastly.Token, error) {
		return &fastly.Token{
			TokenID:   fastly.ToPointer("tok-id-expiry"),
			Name:      fastly.ToPointer("expiring-token"),
			Scope:     &testTokenScope,
			ExpiresAt: &testTokenExpiry,
		}, nil
	}
	testTokenSelfNoName = func(_ context.Context) (*fastly.Token, error) {
		return &fastly.Token{
			TokenID: fastly.ToPointer("tok-id-noname"),
			Scope:   &testTokenScope,
		}, nil
	}
	testGetCurrentUser = func(_ context.Context) (*fastly.User, error) {
		return &fastly.User{
			Login:      fastly.ToPointer("alice@example.com"),
			CustomerID: fastly.ToPointer("cust-abc123"),
		}, nil
	}
)

func TestAuthAdd(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "add with explicit name stores metadata",
			Args: "add mytoken --api-token test-token-value",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfFull,
			},
			WantOutputs: []string{`Token "mytoken" added`, "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("mytoken")
				if at == nil {
					t.Fatal("expected auth token 'mytoken' to exist")
				}
				if at.Email != "alice@example.com" {
					t.Errorf("want email alice@example.com, got %s", at.Email)
				}
				if at.AccountID != "cust-abc123" {
					t.Errorf("want account ID cust-abc123, got %s", at.AccountID)
				}
				if at.APITokenName != "my-api-token" {
					t.Errorf("want APITokenName my-api-token, got %s", at.APITokenName)
				}
				if at.APITokenScope != "global" {
					t.Errorf("want APITokenScope global, got %s", at.APITokenScope)
				}
				if at.APITokenID != "tok-id-123" {
					t.Errorf("want APITokenID tok-id-123, got %s", at.APITokenID)
				}
			},
		},
		{
			Name: "add without name derives from API token name",
			Args: "add --api-token test-token-value",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfFull,
			},
			WantOutputs: []string{`Token "my-api-token" added`, "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("my-api-token")
				if at == nil {
					t.Fatal("expected auth token 'my-api-token' to exist")
				}
				if at.APITokenName != "my-api-token" {
					t.Errorf("want APITokenName my-api-token, got %s", at.APITokenName)
				}
			},
		},
		{
			Name: "add without name fails when API token has no name",
			Args: "add --api-token test-token-value",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfNoName,
			},
			WantError: "could not determine a name for this token",
		},
		{
			Name: "add stores expiry when present",
			Args: "add expiring --api-token test-token-value",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfWithExpiry,
			},
			WantOutputs: []string{`Token "expiring" added`, "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("expiring")
				if at == nil {
					t.Fatal("expected auth token to exist")
				}
				if at.APITokenExpiresAt == "" {
					t.Error("expected APITokenExpiresAt to be set")
				}
			},
		},
		{
			Name: "add sets default when no default exists",
			Args: "add first-token --api-token test-token-value",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfFull,
			},
			ConfigFile: &config.File{
				Auth: config.Auth{},
			},
			WantOutputs: []string{`Token "first-token" added`, "set as default"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				if opts.Config.Auth.Default != "first-token" {
					t.Errorf("want Auth.Default first-token, got %s", opts.Config.Auth.Default)
				}
			},
		},
		{
			Name: "add rejects duplicate name",
			Args: "add existing --api-token test-token-value",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfFull,
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "existing",
					Tokens: config.AuthTokens{
						"existing": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "old-token",
						},
					},
				},
			},
			WantError: `token "existing" already exists`,
		},
	}

	testutil.RunCLIScenarios(t, []string{"auth"}, scenarios)
}

func TestAuthLogin(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "login stores metadata under API token name",
			Args: "login",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfFull,
			},
			Stdin: []string{
				"my-login-token",
			},
			WantOutputs: []string{`Authenticated as alice@example.com (token stored as "my-api-token")`, "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("my-api-token")
				if at == nil {
					t.Fatal("expected auth token 'my-api-token' to exist")
				}
				if at.APITokenName != "my-api-token" {
					t.Errorf("want APITokenName my-api-token, got %s", at.APITokenName)
				}
				if at.APITokenScope != "global" {
					t.Errorf("want APITokenScope global, got %s", at.APITokenScope)
				}
				if at.APITokenID != "tok-id-123" {
					t.Errorf("want APITokenID tok-id-123, got %s", at.APITokenID)
				}
				if at.Email != "alice@example.com" {
					t.Errorf("want email alice@example.com, got %s", at.Email)
				}
				if opts.Config.Auth.Default != "my-api-token" {
					t.Errorf("want Auth.Default my-api-token, got %s", opts.Config.Auth.Default)
				}
			},
		},
		{
			Name: "login falls back to default when API token has no name",
			Args: "login",
			API: mock.API{
				GetCurrentUserFn: testGetCurrentUser,
				GetTokenSelfFn:   testTokenSelfNoName,
			},
			Stdin: []string{
				"my-login-token",
			},
			WantOutputs: []string{`Authenticated as alice@example.com (token stored as "default")`, "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("default")
				if at == nil {
					t.Fatal("expected auth token 'default' to exist")
				}
				if opts.Config.Auth.Default != "default" {
					t.Errorf("want Auth.Default default, got %s", opts.Config.Auth.Default)
				}
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{"auth"}, scenarios)
}

func TestAuthShow(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "show displays metadata when present",
			Args: "show mytoken",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "mytoken",
					Tokens: config.AuthTokens{
						"mytoken": &config.AuthToken{
							Type:              config.AuthTokenTypeStatic,
							Token:             "test-token-value",
							Email:             "alice@example.com",
							AccountID:         "cust-abc123",
							APITokenName:      "my-api-token",
							APITokenScope:     "global",
							APITokenExpiresAt: "2025-12-31T23:59:59Z",
							APITokenID:        "tok-id-123",
						},
					},
				},
			},
			WantOutputs: []string{
				"API token name: my-api-token",
				"API token scope: global",
				"API token expires at: 2025-12-31T23:59:59Z",
				"API token ID: tok-id-123",
			},
		},
		{
			Name: "show omits metadata when absent",
			Args: "show basic",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "basic",
					Tokens: config.AuthTokens{
						"basic": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "test-token-value",
						},
					},
				},
			},
			DontWantOutputs: []string{
				"API token name:",
				"API token scope:",
				"API token expires at:",
				"API token ID:",
			},
		},
		{
			Name: "show without name uses default token",
			Args: "show",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "mytoken",
					Tokens: config.AuthTokens{
						"mytoken": &config.AuthToken{
							Type:         config.AuthTokenTypeStatic,
							Token:        "test-token-value",
							Email:        "alice@example.com",
							APITokenName: "my-api-token",
						},
					},
				},
			},
			WantOutputs: []string{
				"Name: mytoken (default)",
				"Email: alice@example.com",
				"API token name: my-api-token",
			},
		},
		{
			Name: "show without name errors when env token set",
			Args: "show",
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Env.APIToken = "env-token-value"
			},
			ConfigFile: &config.File{},
			WantError:  "current token is not stored",
		},
		{
			Name: "show without name resolves manifest profile",
			Args: "show",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "other",
					Tokens: config.AuthTokens{
						"other": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "other-token",
						},
						"from-manifest": &config.AuthToken{
							Type:         config.AuthTokenTypeStatic,
							Token:        "manifest-token",
							Email:        "manifest@example.com",
							APITokenName: "manifest-api-token",
						},
					},
				},
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Manifest.File.Profile = "from-manifest"
			},
			WantOutputs: []string{
				"Name: from-manifest",
				"Email: manifest@example.com",
				"API token name: manifest-api-token",
			},
			DontWantOutputs: []string{
				"(default)",
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{"auth"}, scenarios)
}

func TestAuthAddScopedToken(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name: "add with name succeeds when GetCurrentUser fails but GetTokenSelf succeeds",
			Args: "add purge-token --api-token scoped-token-value",
			API: mock.API{
				GetCurrentUserFn: func(_ context.Context) (*fastly.User, error) {
					return nil, fmt.Errorf("403 Forbidden: Access denied to purge token")
				},
				GetTokenSelfFn: testTokenSelfFull,
			},
			WantOutputs: []string{`Token "purge-token" added`, "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("purge-token")
				if at == nil {
					t.Fatal("expected auth token 'purge-token' to exist")
				}
				if at.Token != "scoped-token-value" {
					t.Errorf("want token scoped-token-value, got %s", at.Token)
				}
				if at.Email != "" {
					t.Errorf("want empty email for scoped token, got %s", at.Email)
				}
				if at.APITokenName != "my-api-token" {
					t.Errorf("want APITokenName my-api-token, got %s", at.APITokenName)
				}
				if at.APITokenScope != "global" {
					t.Errorf("want APITokenScope global, got %s", at.APITokenScope)
				}
			},
		},
		{
			Name: "add with name fails when both API calls fail (invalid token)",
			Args: "add bad-token --api-token invalid-token-value",
			API: mock.API{
				GetCurrentUserFn: func(_ context.Context) (*fastly.User, error) {
					return nil, fmt.Errorf("403 Forbidden")
				},
				GetTokenSelfFn: func(_ context.Context) (*fastly.Token, error) {
					return nil, fmt.Errorf("403 Forbidden")
				},
			},
			WantError: "token validation failed: neither /current_user nor /tokens/self responded successfully",
		},
		{
			Name: "add without name gives friendly error for scoped token",
			Args: "add --api-token scoped-token-value",
			API: mock.API{
				GetCurrentUserFn: func(_ context.Context) (*fastly.User, error) {
					return nil, fmt.Errorf("403 Forbidden: Access denied to purge token")
				},
				GetTokenSelfFn: testTokenSelfNoName,
			},
			WantError: "could not determine a name for this token",
		},
	}

	testutil.RunCLIScenarios(t, []string{"auth"}, scenarios)
}

func TestEnrichWithTokenSelfPreservesOnFailure(t *testing.T) {
	var stdout threadsafe.Buffer
	data := testutil.MockGlobalData([]string{"fastly"}, &stdout)

	data.APIClientFactory = mock.APIClient(mock.API{
		GetTokenSelfFn: func(_ context.Context) (*fastly.Token, error) {
			return nil, fmt.Errorf("403 forbidden")
		},
	})

	at := &config.AuthToken{
		Token:             "existing-token",
		APITokenName:      "original-name",
		APITokenScope:     "global",
		APITokenExpiresAt: "2025-06-01T00:00:00Z",
		APITokenID:        "original-id",
	}

	authcmd.EnrichWithTokenSelf(data, at)

	if at.APITokenName != "original-name" {
		t.Errorf("want APITokenName preserved as original-name, got %s", at.APITokenName)
	}
	if at.APITokenScope != "global" {
		t.Errorf("want APITokenScope preserved as global, got %s", at.APITokenScope)
	}
	if at.APITokenExpiresAt != "2025-06-01T00:00:00Z" {
		t.Errorf("want APITokenExpiresAt preserved, got %s", at.APITokenExpiresAt)
	}
	if at.APITokenID != "original-id" {
		t.Errorf("want APITokenID preserved as original-id, got %s", at.APITokenID)
	}
}
