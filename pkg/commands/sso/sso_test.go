package sso_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v14/fastly"

	"github.com/fastly/cli/pkg/auth"
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
			Args: "sso",
			Stdin: []string{
				"N", // when prompted to open a web browser to start authentication
			},
			WantError: "will not continue",
		},
		// 1. Error opening web browser
		{
			Args: "sso",
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
			Args: "sso",
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
			Args: "sso",
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
		// 4. Success processing OAuth flow (uses default auth token name "user" from MockGlobalData)
		{
			Args: "sso",
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
				"We're going to authenticate the 'user' token",
				"We need to open your browser to authenticate you.",
				"Session token 'user' has been stored.",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("user")
				if at == nil {
					t.Fatal("expected 'user' auth token to exist")
				}
				if at.Token != "123" {
					t.Errorf("want token: 123, got token: %s", at.Token)
				}
			},
		},
		// 5. Success processing OAuth flow while setting specific profile (test_user)
		{
			Args: "sso test_user",
			ConfigFile: &config.File{
				Auth: config.Auth{
					Default: "test_user",
					Tokens: config.AuthTokens{
						"test_user": &config.AuthToken{
							Type:  config.AuthTokenTypeSSO,
							Token: "mock-token",
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
				"Session token 'test_user' has been stored.",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("test_user")
				if at == nil {
					t.Fatal("expected 'test_user' auth token to exist")
				}
				if at.Token != "123" {
					t.Errorf("want token: 123, got token: %s", at.Token)
				}
			},
		},
		// NOTE: The following tests indirectly validate our `app.Run()` logic.
		// Specifically the processing of the token before invoking the subcommand.
		// It allows us to check that the `sso` command is invoked when expected.
		//
		// 6. Success processing `pops` command.
		// We configure a non-SSO token so we can validate the INFO message.
		// Otherwise no OAuth flow is happening here.
		{
			Args: "pops",
			API: &mock.API{
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
				at := opts.Config.GetAuthToken("user")
				if at == nil {
					t.Fatal("expected 'user' auth token to exist")
				}
				if at.Token != "mock-token" {
					t.Errorf("want token: mock-token, got token: %s", at.Token)
				}
			},
		},
		// 7. SSO token with both access and refresh expired.
		// The `whoami` command triggers re-auth via the processToken flow.
		// The user declines re-authentication.
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
							AccessExpiresAt:  time.Now().Add(-600 * time.Second).Format(time.RFC3339),
							RefreshExpiresAt: time.Now().Add(-600 * time.Second).Format(time.RFC3339),
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
		// 8. SSO token with both access and refresh expired.
		// The user accepts re-auth, and the `pops` command executes after.
		{
			Args: "pops",
			API: &mock.API{
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
							AccessExpiresAt:  time.Now().Add(-300 * time.Second).Format(time.RFC3339),
							RefreshExpiresAt: time.Now().Add(-300 * time.Second).Format(time.RFC3339),
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
				"Your auth token has expired and needs re-authentication",
				"Starting a local server to handle the authentication flow.",
				"Session token 'user' has been stored.",
				"{Latitude:1 Longitude:2 X:3 Y:4}",
			},
			Validator: func(t *testing.T, _ *testutil.CLIScenario, opts *global.Data, _ *threadsafe.Buffer) {
				at := opts.Config.GetAuthToken("user")
				if at == nil {
					t.Fatal("expected 'user' auth token to exist")
				}
				if at.Token != "123" {
					t.Errorf("want token: 123, got token: %s", at.Token)
				}
			},
		},
	}

	// unlike the usual usage of this function, the "command name"
	// slice is empty here because the commands to be run are
	// embedded in the scenarios (some scenarios run different
	// commands)
	testutil.RunCLIScenarios(t, []string{}, scenarios)
}
