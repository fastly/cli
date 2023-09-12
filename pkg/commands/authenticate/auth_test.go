package authenticate_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestAuth(t *testing.T) {
	args := testutil.Args
	type ts struct {
		testutil.TestScenario

		AuthResult    *auth.AuthorizationResult
		ConfigProfile *config.Profile
		Opener        func(input string) error
		Stdin         []string
	}
	scenarios := []ts{
		// User cancels authentication prompt
		{
			TestScenario: testutil.TestScenario{
				Args:      args("authenticate"),
				WantError: "user cancelled execution",
			},
			Stdin: []string{
				"N", // when prompted to open a web browser to start authentication
			},
		},
		// Error opening web browser
		{
			TestScenario: testutil.TestScenario{
				Args:      args("authenticate"),
				WantError: "failed to open web browser",
			},
			Opener: func(input string) error {
				return errors.New("failed to open web browser")
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// Error processing OAuth flow (error encountered)
		{
			TestScenario: testutil.TestScenario{
				Args:      args("authenticate"),
				WantError: "failed to authorize: no authorization code returned",
			},
			AuthResult: &auth.AuthorizationResult{
				Err: errors.New("no authorization code returned"),
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// Error processing OAuth flow (empty SessionToken field)
		{
			TestScenario: testutil.TestScenario{
				Args:      args("authenticate"),
				WantError: "failed to authorize: no session token",
			},
			AuthResult: &auth.AuthorizationResult{
				SessionToken: "",
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
		// Success processing OAuth flow
		{
			TestScenario: testutil.TestScenario{
				Args:       args("authenticate"),
				WantOutput: "Session token (persisted to your local configuration): 123",
			},
			AuthResult: &auth.AuthorizationResult{
				SessionToken: "123",
			},
			ConfigProfile: &config.Profile{
				Token: "123",
			},
			Stdin: []string{
				"Y", // when prompted to open a web browser to start authentication
			},
		},
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)

			if testcase.AuthResult != nil {
				result := make(chan auth.AuthorizationResult)
				opts.AuthServer = testutil.MockAuthServer{
					Result: result,
				}
				go func() {
					result <- *testcase.AuthResult
				}()
			}
			if testcase.Opener != nil {
				opts.Opener = testcase.Opener
			}

			var err error

			if len(testcase.Stdin) > 1 {
				// To handle multiple prompt input from the user we need to do some
				// coordination around io pipes to mimic the required user behaviour.
				stdin, prompt := io.Pipe()
				opts.Stdin = stdin

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
					err = app.Run(opts)
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
				opts.Stdin = strings.NewReader(stdin)
				err = app.Run(opts)
			}

			if testcase.ConfigProfile != nil {
				userProfile := opts.ConfigFile.Profiles["user"]
				if userProfile.Token != testcase.ConfigProfile.Token {
					t.Errorf("want token: %s, got token: %s", testcase.ConfigProfile.Token, userProfile.Token)
				}
			}

			t.Log(stdout.String())

			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
