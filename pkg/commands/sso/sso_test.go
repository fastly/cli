package sso_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestSSO(t *testing.T) {
	args := testutil.Args
	type ts struct {
		testutil.TestScenario

		AuthResult            *auth.AuthorizationResult
		ConfigFile            *config.File
		ExpectedConfigProfile *config.Profile
		HTTPClient            api.HTTPClient
		Opener                func(input string) error
		Stdin                 []string
	}
	scenarios := []ts{
		// 0. User cancels authentication prompt
		{
			TestScenario: testutil.TestScenario{
				Args:      args("sso"),
				WantError: "user cancelled execution",
			},
			Stdin: []string{
				"N", // when prompted to open a web browser to start authentication
			},
		},
		// 1. Error opening web browser
		{
			TestScenario: testutil.TestScenario{
				Args:      args("sso"),
				WantError: "failed to open web browser",
			},
			Opener: func(input string) error {
				return errors.New("failed to open web browser")
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// 2. Error processing OAuth flow (error encountered)
		{
			TestScenario: testutil.TestScenario{
				Args:      args("sso"),
				WantError: "failed to authorize: no authorization code returned",
			},
			AuthResult: &auth.AuthorizationResult{
				Err: errors.New("no authorization code returned"),
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// 3. Error processing OAuth flow (empty SessionToken field)
		{
			TestScenario: testutil.TestScenario{
				Args:      args("sso"),
				WantError: "failed to authorize: no session token",
			},
			AuthResult: &auth.AuthorizationResult{
				SessionToken: "",
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// 4. Success processing OAuth flow
		{
			TestScenario: testutil.TestScenario{
				Args: args("sso"),
				WantOutputs: []string{
					"We're going to authenticate the 'user' profile.",
					"We need to open your browser to authenticate you.",
					"Session token (persisted to your local configuration): 123",
				},
			},
			AuthResult: &auth.AuthorizationResult{
				SessionToken: "123",
			},
			ExpectedConfigProfile: &config.Profile{
				Token: "123",
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// 5. Success processing OAuth flow while setting specific profile (test_user)
		{
			TestScenario: testutil.TestScenario{
				Args: args("sso test_user"),
				WantOutputs: []string{
					"We're going to authenticate the 'test_user' profile.",
					"We need to open your browser to authenticate you.",
					"Session token (persisted to your local configuration): 123",
				},
			},
			AuthResult: &auth.AuthorizationResult{
				SessionToken: "123",
			},
			ConfigFile: &config.File{
				Profiles: config.Profiles{
					"test_user": &config.Profile{
						Default: true,
						Email:   "test@example.com",
						Token:   "mock-token",
					},
				},
			},
			ExpectedConfigProfile: &config.Profile{
				Token: "123",
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
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
			TestScenario: testutil.TestScenario{
				Args: args("whoami"),
				WantOutputs: []string{
					// FIXME: Put back messaging once SSO is GA.
					// "is not a Fastly SSO (Single Sign-On) generated token",
					"Alice Programmer <alice@example.com>",
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
			ExpectedConfigProfile: &config.Profile{
				Token: "mock-token",
			},
			HTTPClient: testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
		},
		// 7. Success processing `whoami` command.
		// We set an SSO token that has expired.
		// This allows us to validate the output message about expiration.
		// We don't respond "Y" to prompt to reauthenticate.
		// But we've mocked the request to succeeds still so it doesn't matter.
		{
			TestScenario: testutil.TestScenario{
				Args: args("whoami"),
				WantOutputs: []string{
					"Your access token has expired and so has your refresh token.",
					"Alice Programmer <alice@example.com>",
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
			ExpectedConfigProfile: &config.Profile{
				Token: "mock-token",
			},
			HTTPClient: testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
		},
		// 8. Success processing OAuth flow via `whoami` command
		// We set an SSO token that has expired.
		// This allows us to validate the output messages.
		{
			TestScenario: testutil.TestScenario{
				Args: args("whoami"),
				WantOutputs: []string{
					"Your access token has expired and so has your refresh token.",
					"Starting a local server to handle the authentication flow.",
					"Session token (persisted to your local configuration): 123",
					"Alice Programmer <alice@example.com>",
				},
			},
			AuthResult: &auth.AuthorizationResult{
				SessionToken: "123",
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
			ExpectedConfigProfile: &config.Profile{
				Token: "123",
			},
			HTTPClient: testutil.WhoamiVerifyClient(testutil.WhoamiBasicResponse),
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.MockGlobalData(testcase.Args, &stdout)
			opts.APIClientFactory = mock.APIClient(testcase.API)

			if testcase.HTTPClient != nil {
				opts.HTTPClient = testcase.HTTPClient
			}
			if testcase.ConfigFile != nil {
				opts.Config = *testcase.ConfigFile
			}
			if testcase.Opener != nil {
				opts.Opener = testcase.Opener
			}
			if testcase.AuthResult != nil {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- *testcase.AuthResult
				}()
			}

			var err error

			if len(testcase.Stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Input = stdin

				// Wait for user input and write it to the prompt
				inputc := make(chan string)
				go func() {
					for input := range inputc {
						fmt.Fprintln(prompt, input)
					}
				}()

				// We need a channel so we wait for `Run()` to complete
				done := make(chan bool)

				// Call `app.Run()` and wait for response
				go func() {
					app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
						return opts, nil
					}
					err = app.Run(testcase.Args, nil)
					done <- true
				}()

				// User provides input
				//
				// NOTE: Must provide as much input as is expected to be waited on by `run()`.
				//       For example, if `run()` calls `input()` twice, then provide two messages.
				//       Otherwise the select statement will trigger the timeout error.
				for _, input := range testcase.Stdin {
					inputc <- input
				}

				select {
				case <-done:
					// Wait for app.Run() to finish
				case <-time.After(10 * time.Second):
					t.Fatalf("unexpected timeout waiting for mocked prompt inputs to be processed")
				}
			} else {
				stdin := ""
				if len(testcase.Stdin) > 0 {
					stdin = testcase.Stdin[0]
				}
				opts.Input = strings.NewReader(stdin)
				app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
					return opts, nil
				}
				err = app.Run(testcase.Args, nil)
			}

			if testcase.ExpectedConfigProfile != nil {
				profileName := "user"
				if len(testcase.Args) > 1 {
					profileName = testcase.Args[1] // use the `profile` command argument
				}
				userProfile := opts.Config.Profiles[profileName]
				if userProfile.Token != testcase.ExpectedConfigProfile.Token {
					t.Errorf("want token: %s, got token: %s", testcase.ExpectedConfigProfile.Token, userProfile.Token)
				}
			}

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)

			if testcase.WantOutputs != nil {
				for _, s := range testcase.WantOutputs {
					testutil.AssertStringContains(t, stdout.String(), s)
				}
			} else {
				testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
			}
		})
	}
}
