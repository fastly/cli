package sso_test

import (
	"errors"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v10/fastly"

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
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				opts.Opener = func(input string) error {
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
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
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
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
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
		// 4. Success processing OAuth flow
		{
			Args: "sso",
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
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
				"We're going to authenticate the 'user' profile",
				"We need to open your browser to authenticate you.",
				"Session token (persisted to your local configuration): 123",
			},
			Validator: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data, stdout *threadsafe.Buffer) {
				const expectedToken = "123"
				userProfile := opts.Config.Profiles["user"]
				if userProfile.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, userProfile.Token)
				}
			},
		},
		// 5. Success processing OAuth flow while setting specific profile (test_user)
		{
			Args: "sso test_user",
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"test_user": &config.Profile{
						Default: true,
						Email:   "test@example.com",
						Token:   "mock-token",
					},
				},
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
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
				"We're going to authenticate the 'test_user' profile",
				"We need to open your browser to authenticate you.",
				"Session token (persisted to your local configuration): 123",
			},
			Validator: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data, stdout *threadsafe.Buffer) {
				const expectedToken = "123"
				userProfile := opts.Config.Profiles["test_user"]
				if userProfile.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, userProfile.Token)
				}
			},
		},
		// NOTE: The following tests indirectly validate our `app.Run()` logic.
		// Specifically the processing of the token before invoking the subcommand.
		// It allows us to check that the `sso` command is invoked when expected.
		//
		// 6. Success processing `whoami` command.
		// We configure a non-SSO token so we can validate the INFO message.
		// Otherwise no OAuth flow is happening here.
		{
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func() ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:   fastly.ToPointer(float64(1)),
								Longtitude: fastly.ToPointer(float64(2)),
								X:          fastly.ToPointer(float64(3)),
								Y:          fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"user": &config.Profile{
						Default: true,
						Email:   "test@example.com",
						Token:   "mock-token",
					},
				},
			},
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutputs: []string{
				// FIXME: Put back messaging once SSO is GA.
				// "is not a Fastly SSO (Single Sign-On) generated token",
				"{Latitude:1 Longtitude:2 X:3 Y:4}",
			},
			Validator: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data, stdout *threadsafe.Buffer) {
				const expectedToken = "mock-token"
				userProfile := opts.Config.Profiles["user"]
				if userProfile.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, userProfile.Token)
				}
			},
		},
		// 7. Success processing `whoami` command.
		// We set an SSO token that has expired.
		// This allows us to validate the output message about expiration.
		// We don't respond "Y" to the prompt for reauthentication.
		// But we've mocked the request to succeed still so it doesn't matter.
		{
			Args: "whoami",
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"user": &config.Profile{
						AccessTokenCreated: time.Now().Add(-(time.Duration(600) * time.Second)).Unix(), // 10 mins ago
						Default:            true,
						Email:              "test@example.com",
						Token:              "mock-token",
					},
				},
			},
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				opts.HTTPClient = testutil.CurrentCustomerClient(testutil.CurrentCustomerResponse)
			},
			WantOutput:     "Your access token has expired and so has your refresh token.",
			DontWantOutput: "{Latitude:1 Longtitude:2 X:3 Y:4}",
			Validator: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data, stdout *threadsafe.Buffer) {
				const expectedToken = "mock-token"
				userProfile := opts.Config.Profiles["user"]
				if userProfile.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, userProfile.Token)
				}
			},
		},
		// 8. Success processing OAuth flow via `whoami` command
		// We set an SSO token that has expired.
		// This allows us to validate the output messages.
		{
			Args: "pops",
			API: mock.API{
				AllDatacentersFn: func() ([]fastly.Datacenter, error) {
					return []fastly.Datacenter{
						{
							Name:   fastly.ToPointer("Foobar"),
							Code:   fastly.ToPointer("FBR"),
							Group:  fastly.ToPointer("Bar"),
							Shield: fastly.ToPointer("Baz"),
							Coordinates: &fastly.Coordinates{
								Latitude:   fastly.ToPointer(float64(1)),
								Longtitude: fastly.ToPointer(float64(2)),
								X:          fastly.ToPointer(float64(3)),
								Y:          fastly.ToPointer(float64(4)),
							},
						},
					}, nil
				},
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"user": &config.Profile{
						AccessTokenCreated: time.Now().Add(-(time.Duration(300) * time.Second)).Unix(), // 5 mins ago
						Default:            true,
						Email:              "test@example.com",
						Token:              "mock-token",
					},
				},
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
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
				"Your access token has expired and so has your refresh token.",
				"Starting a local server to handle the authentication flow.",
				"Session token (persisted to your local configuration): 123",
				"{Latitude:1 Longtitude:2 X:3 Y:4}",
			},
			Validator: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data, stdout *threadsafe.Buffer) {
				const expectedToken = "123"
				userProfile := opts.Config.Profiles["user"]
				if userProfile.Token != expectedToken {
					t.Errorf("want token: %s, got token: %s", expectedToken, userProfile.Token)
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
