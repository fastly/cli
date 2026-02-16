package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/auth"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/threadsafe"
)

func TestSSO(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		// 0. User cancels authentication prompt
		{
			Args: "auth login --sso",
			Stdin: []string{
				"N", // when prompted to open a web browser to start authentication
			},
			WantError: "will not continue",
		},
		// 1. Error opening web browser
		{
			Args: "auth login --sso",
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Opener = func(_ string) error {
					return errors.New("failed to open web browser")
				}
			},
			WantError: "failed to open web browser",
		},
		// 2. Error processing OAuth flow (error encountered)
		{
			Args: "auth login --sso",
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						Err: errors.New("no authorization code returned"),
					}
				}()
			},
			WantError: "failed to authorize: no authorization code returned",
		},
		// 3. Error processing OAuth flow (empty SessionToken field)
		{
			Args: "auth login --sso",
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "",
					}
				}()
			},
			WantError: "failed to authorize: no session token",
		},
		// 4. Success processing OAuth flow (stores as "sso" by default)
		{
			Args: "auth login --sso",
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "123",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"We're going to authenticate the 'sso' token",
				"We need to open your browser to authenticate you.",
				"has been stored. Use 'fastly auth list' to view tokens.",
				"Token saved to",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				const expectedToken = "123"
				at := opts.Config.GetAuthToken("sso")
				if at == nil {
					t.Fatal("expected auth token 'sso' to exist")
				}
				if at.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, at.Token)
				}
				if opts.Config.Auth.Default != "sso" {
					t.Errorf("want default: sso, got: %s", opts.Config.Auth.Default)
				}
			},
		},
		// 5. Success processing OAuth flow targeting specific auth token via --token flag
		{
			Args: "auth login --sso --token test_user",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "test_user",
					Tokens: config.AuthTokens{
						"test_user": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "old-token",
							Email: "test@example.com",
						},
					},
				},
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "123",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"We're going to authenticate the 'test_user' token",
				"We need to open your browser to authenticate you.",
				"has been stored. Use 'fastly auth list' to view tokens.",
				"Token saved to",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				const expectedToken = "123"
				at := opts.Config.GetAuthToken("test_user")
				if at == nil {
					t.Fatal("expected auth token 'test_user' to exist")
				}
				if at.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, at.Token)
				}
			},
		},
		// 6. Success processing `pops` command with a static (non-SSO) auth token.
		{
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "mock-token",
							Email: "test@example.com",
						},
					},
				},
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"{Latitude:1 Longitude:2 X:3 Y:4}",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				const expectedToken = "mock-token"
				at := opts.Config.GetAuthToken("user")
				if at == nil {
					t.Fatal("expected auth token 'user' to exist")
				}
				if at.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, at.Token)
				}
			},
		},
		// 7. SSO token with both access and refresh expired -> triggers re-auth prompt.
		// The user declines, so the command does not execute.
		{
			Args: "whoami",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:             config.AuthTokenTypeSSO,
							Token:            "mock-token",
							Email:            "test@example.com",
							RefreshToken:     "mock-refresh",
							AccessExpiresAt:  time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
							RefreshExpiresAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
						},
					},
				},
			},
			Stdin: []string{
				"N", // decline re-authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutput:     "Your auth token has expired and needs re-authentication",
			DontWantOutput: "{Latitude:1 Longitude:2 X:3 Y:4}",
		},
		// 8. SSO token with both access and refresh expired -> user accepts re-auth.
		// The SSO flow succeeds and the command executes afterward.
		{
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:             config.AuthTokenTypeSSO,
							Token:            "mock-token",
							Email:            "test@example.com",
							RefreshToken:     "mock-refresh",
							AccessExpiresAt:  time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
							RefreshExpiresAt: time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
						},
					},
				},
			},
			Stdin: []string{
				"Y", // accept re-authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "new-123",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"Your auth token has expired and needs re-authentication",
				"Starting a local server to handle the authentication flow.",
				"has been stored. Use 'fastly auth list' to view tokens.",
				"Token saved to",
				"{Latitude:1 Longitude:2 X:3 Y:4}",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				const expectedToken = "new-123"
				at := opts.Config.GetAuthToken("user")
				if at == nil {
					t.Fatal("expected auth token 'user' to exist")
				}
				if at.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, at.Token)
				}
			},
		},
		// 9. Migration before token resolution: legacy profiles are migrated to
		// [auth] before processToken() runs, so the token resolves correctly.
		{
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"legacy": &config.Profile{
						Default: true,
						Email:   "legacy@example.com",
						Token:   "legacy-token",
					},
				},
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"{Latitude:1 Longitude:2 X:3 Y:4}",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("legacy")
				if at == nil {
					t.Fatal("expected migrated auth token 'legacy' to exist")
				}
				if at.Token != "legacy-token" {
					t.Errorf("want token: legacy-token, got token: %s", at.Token)
				}
				if opts.Config.Auth.Default != "legacy" {
					t.Errorf("want default auth: legacy, got: %s", opts.Config.Auth.Default)
				}
			},
		},
		// 11. auth login --sso --token newname creates a new token that doesn't exist yet.
		{
			Args: "auth login --sso --token brandnew",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "existing",
					Tokens: config.AuthTokens{
						"existing": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "existing-token",
						},
					},
				},
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "brand-new-token",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"We're going to authenticate the 'brandnew' token",
				"Session token 'brandnew' has been stored.",
				"Token saved to",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("brandnew")
				if at == nil {
					t.Fatal("expected auth token 'brandnew' to exist")
				}
				if at.Token != "brand-new-token" {
					t.Errorf("want token: brand-new-token, got token: %s", at.Token)
				}
				if opts.Config.Auth.Default != "brandnew" {
					t.Errorf("want default auth: brandnew, got: %s", opts.Config.Auth.Default)
				}
			},
		},
		// 12. Missing manifest profile emits a single warning.
		{
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "mock-token",
						},
					},
				},
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Manifest.File.Profile = "nonexistent"
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				`profile "nonexistent" not found in auth config, using default token "user"`,
				"{Latitude:1 Longitude:2 X:3 Y:4}",
			},
		},
		// 13. Missing manifest profile warning is suppressed under --quiet.
		{
			Args: "pops --quiet",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "mock-token",
						},
					},
				},
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Manifest.File.Profile = "nonexistent"
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			DontWantOutput: "not found in auth config",
		},
		// 14. Missing manifest profile warning suppressed when --token overrides.
		{
			Args: "pops --token override-token",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "user",
					Tokens: config.AuthTokens{
						"user": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "mock-token",
						},
					},
				},
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Manifest.File.Profile = "nonexistent"
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			DontWantOutput: "not found in auth config",
		},
		// 15. Auto-prompt: directly prompts for a static API token.
		{
			Name: "auto-prompt static token",
			Args: "whoami",
			API: mock.API{
				GetCurrentUserFn: func(_ context.Context) (*fastly.User, error) {
					return &fastly.User{
						Login:      fastly.ToPointer("alice@example.com"),
						CustomerID: fastly.ToPointer("abc123"),
					}, nil
				},
				GetTokenSelfFn: testTokenSelfFull,
			},
			ConfigFile: &config.File{},
			Stdin: []string{
				"my-static-token",
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.HTTPClient = testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse)
			},
			WantOutputs: []string{
				"This command requires authentication to access your Fastly account.",
				"Paste your API token",
				`Authenticated as alice@example.com (token stored as "my-api-token")`,
				"Token saved to",
			},
			DontWantOutput: "Log in with browser",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("my-api-token")
				if at == nil {
					t.Fatal("expected auth token 'my-api-token' to exist")
				}
				if at.Token != "my-static-token" {
					t.Errorf("want token: my-static-token, got token: %s", at.Token)
				}
				if at.Type != config.AuthTokenTypeStatic {
					t.Errorf("want type: static, got type: %s", at.Type)
				}
				if opts.Config.Auth.Default != "my-api-token" {
					t.Errorf("want Auth.Default my-api-token, got %s", opts.Config.Auth.Default)
				}
			},
		},
		// 16. Non-interactive: returns ErrNonInteractiveNoToken without prompting.
		{
			Name:            "non-interactive no prompt",
			Args:            "whoami -i",
			ConfigFile:      &config.File{},
			DontWantOutput:  "requires authentication",
			WantError:       "no token provided",
			WantRemediation: "Interactive authentication is not available",
		},
		// 17. Auto-yes: errors immediately from processToken (no prompt).
		{
			Name:            "auto-prompt auto-yes",
			Args:            "pops -y",
			ConfigFile:      &config.File{},
			DontWantOutput:  "requires authentication",
			WantError:       "no token provided",
			WantRemediation: "Interactive authentication is not available",
		},
		// 18. Accept-defaults: errors immediately from processToken (no prompt).
		{
			Name:            "auto-prompt accept-defaults",
			Args:            "pops -d",
			ConfigFile:      &config.File{},
			DontWantOutput:  "requires authentication",
			WantError:       "no token provided",
			WantRemediation: "Interactive authentication is not available",
		},
		// 19. FASTLY_USE_SSO=1 triggers SSO flow instead of token prompt.
		{
			Name: "auto-prompt use-sso env",
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func(_ context.Context) ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:  fastly.ToPointer(float64(1)),
								Longitude: fastly.ToPointer(float64(2)),
								X:         fastly.ToPointer(float64(3)),
								Y:         fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{},
			Stdin: []string{
				"y",
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Env.UseSSO = "1"
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "sso-env-token",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"Do you want to continue",
				"has been stored",
				"Token saved to",
				"{Latitude:1 Longitude:2 X:3 Y:4}",
			},
			DontWantOutput: "Paste your API token",
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("default")
				if at == nil {
					t.Fatal("expected auth token 'default' to exist")
				}
				if at.Token != "sso-env-token" {
					t.Errorf("want token: sso-env-token, got token: %s", at.Token)
				}
			},
		},
		// 21. FASTLY_USE_SSO=1 ignored under --non-interactive.
		{
			Name:       "use-sso ignored non-interactive",
			Args:       "whoami -i",
			ConfigFile: &config.File{},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Env.UseSSO = "1"
			},
			DontWantOutput:  "requires authentication",
			WantError:       "no token provided",
			WantRemediation: "Interactive authentication is not available",
		},
		// 22. FASTLY_USE_SSO=1 ignored under --auto-yes.
		{
			Name:       "use-sso ignored auto-yes",
			Args:       "pops -y",
			ConfigFile: &config.File{},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Env.UseSSO = "1"
			},
			DontWantOutput:  "Do you want to continue",
			WantError:       "no token provided",
			WantRemediation: "Interactive authentication is not available",
		},
		// 23. FASTLY_USE_SSO=1 ignored under --accept-defaults.
		{
			Name:       "use-sso ignored accept-defaults",
			Args:       "pops -d",
			ConfigFile: &config.File{},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				opts.Env.UseSSO = "1"
			},
			DontWantOutput:  "Do you want to continue",
			WantError:       "no token provided",
			WantRemediation: "Interactive authentication is not available",
		},
		// 24. SSO login sets default to "sso" but does not overwrite an existing static token's value.
		{
			Name: "sso login switches default preserves static token",
			Args: "auth login --sso",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "mytoken",
					Tokens: config.AuthTokens{
						"mytoken": &config.AuthToken{
							Type:  config.AuthTokenTypeStatic,
							Token: "static-secret",
							Email: "static@example.com",
						},
					},
				},
			},
			Stdin: []string{
				"Y",
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "sso-new-token",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				"We're going to authenticate the 'sso' token",
				"Session token 'sso' has been stored.",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				// SSO token was created
				ssoAt := opts.Config.GetAuthToken("sso")
				if ssoAt == nil {
					t.Fatal("expected auth token 'sso' to exist")
				}
				if ssoAt.Token != "sso-new-token" {
					t.Errorf("want sso token: sso-new-token, got: %s", ssoAt.Token)
				}
				// Static token is untouched
				staticAt := opts.Config.GetAuthToken("mytoken")
				if staticAt == nil {
					t.Fatal("expected auth token 'mytoken' to still exist")
				}
				if staticAt.Token != "static-secret" {
					t.Errorf("want static token: static-secret, got: %s", staticAt.Token)
				}
				// Default switched to the SSO token (login always sets default)
				if opts.Config.Auth.Default != "sso" {
					t.Errorf("want default: sso, got: %s", opts.Config.Auth.Default)
				}
			},
		},
		// 25. SSO login stores API token metadata via EnrichWithTokenSelf.
		{
			Name: "sso stores api token metadata",
			Args: "auth login --sso",
			API: mock.API{
				GetTokenSelfFn: func(_ context.Context) (*fastly.Token, error) {
					scope := fastly.GlobalScope
					return &fastly.Token{
						TokenID: fastly.ToPointer("sso-tok-id"),
						Name:    fastly.ToPointer("sso-api-token"),
						Scope:   &scope,
					}, nil
				},
			},
			Stdin: []string{
				"Y",
			},
			Setup: func(_ *testing.T, _ *testutil.CLIScenario, opts *global.Data) {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- auth.AuthorizationResult{
						SessionToken: "sso-session-token",
						Email:        "sso@example.com",
					}
				}()
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{"has been stored", "Token saved to"},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				t.Helper()
				at := opts.Config.GetAuthToken("sso")
				if at == nil {
					t.Fatal("expected auth token 'sso' to exist")
				}
				if at.APITokenName != "sso-api-token" {
					t.Errorf("want APITokenName sso-api-token, got %s", at.APITokenName)
				}
				if at.APITokenScope != "global" {
					t.Errorf("want APITokenScope global, got %s", at.APITokenScope)
				}
				if at.APITokenID != "sso-tok-id" {
					t.Errorf("want APITokenID sso-tok-id, got %s", at.APITokenID)
				}
			},
		},
	}

	testutil.RunCLIScenarios(t, []string{}, scenarios)
}

// TestSSOSuccessMessageDisableAuthCommand verifies RunSSO omits the
// "fastly auth list" hint when FASTLY_DISABLE_AUTH_COMMAND is set.
func TestSSOSuccessMessageDisableAuthCommand(t *testing.T) {
	t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")

	var stdout threadsafe.Buffer
	data := testutil.MockGlobalData([]string{"fastly"}, &stdout)

	result := make(chan auth.AuthorizationResult)
	data.AuthServer = testutil.MockAuthServer{Result: result}
	go func() {
		result <- auth.AuthorizationResult{SessionToken: "test-token"}
	}()
	data.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
	data.Flags.AutoYes = true // skip interactive prompt

	err := authcmd.RunSSO(nil, &stdout, data, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "has been stored.") {
		t.Errorf("expected success message, got: %s", out)
	}
	if strings.Contains(out, "fastly auth list") {
		t.Errorf("expected no 'fastly auth list' hint when env var set, got: %s", out)
	}
	if !strings.Contains(out, "Token saved to") {
		t.Errorf("expected 'Token saved to' message, got: %s", out)
	}
}
