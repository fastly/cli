package testutil

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/threadsafe"
)

// APIHookCLIScenario represents a CLI test case to be validated,
// where the underlying go-fastly API function will be replaced with a
// mock (via a hook declared in the implementation of the command).
//
// Most of the fields in this struct are optional; if they are not
// provided RunAPIHookCLIScenario will not apply the behavior indicated for
// those fields.
type APIHookCLIScenario[T any] struct {
	// Args is the input arguments for the command to execute (not
	// including the command names themselves).
	Args string
	// ConfigPath will be copied into global.Data.ConfigPath
	ConfigPath string
	// ConfigFile will be copied into global.Data.ConfigFile
	ConfigFile *config.File
	// DontWantOutput will cause the scenario to fail if the
	// string appears in stdout
	DontWantOutput string
	// DontWantOutputs will cause the scenario to fail if any of
	// the strings appear in stdout
	DontWantOutputs []string
	// EnvVars contains environment variables which will be set
	// during the execution of the scenario
	EnvVars map[string]string
	// MockFactory is a function which will return the 'mock' to
	// replace the go-fastly API function; it is structured as a
	// factory so that it can accept (and capture) the 'testing.T'
	// for use while the mock function runs
	MockFactory func(*testing.T) T
	// Name appears in output when tests are executed
	Name            string
	PathContentFlag *PathContentFlag
	// Setup function can perform additional setup before the scenario is run
	Setup func(t *testing.T, scenario *APIHookCLIScenario[T], opts *global.Data)
	// Stdin contains input to be read by the application
	Stdin []string
	// Validator function can perform additional validation on the results
	// of the scenario
	Validator func(t *testing.T, scenario *APIHookCLIScenario[T], opts *global.Data, stdout *threadsafe.Buffer)
	// WantError will cause the scenario to fail if this string
	// does not appear in an Error
	WantError string
	// WantRemediation will cause the scenario to fail if the
	// error's RemediationError.Remediation doesn't contain this string
	WantRemediation string
	// WantOutput will cause the scenario to fail if this string
	// does not appear in stdout
	WantOutput string
	// WantOutputs will cause the scenario to fail if any of the
	// strings do not appear in stdout
	WantOutputs []string
}

// RunAPIHookCLIScenario executes a APIHookCLIScenario struct.
// The Arg field of the scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunAPIHookCLIScenario[T any](t *testing.T, command []string, scenario APIHookCLIScenario[T], hook *T) {
	t.Run(scenario.Name, func(t *testing.T) {
		var (
			err          error
			fullargs     []string
			originalFunc T
			stdout       threadsafe.Buffer
		)

		if len(scenario.Args) > 0 {
			fullargs = slices.Concat(command, SplitArgs(scenario.Args))
		} else {
			fullargs = command
		}

		opts := MockGlobalData(fullargs, &stdout)

		// scenarios of this type should never actually invoke
		// an APIClient, but the application's startup code
		// assumes that they need one and requires a factory
		// to construct one
		opts.APIClientFactory = func(_, _ string, _ bool) (api.Interface, error) {
			fc, err := fastly.NewClientForEndpoint("no-key", "api.example.com")
			if err != nil {
				return nil, fmt.Errorf("failed to mock fastly.Client: %w", err)
			}
			return fc, nil
		}

		if len(scenario.ConfigPath) > 0 {
			opts.ConfigPath = scenario.ConfigPath
		}

		if scenario.ConfigFile != nil {
			opts.Config = *scenario.ConfigFile
		}

		if scenario.EnvVars != nil {
			for key, value := range scenario.EnvVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := os.Unsetenv(key); err != nil {
						t.Fatal(err)
					}
				}()
			}
		}

		if scenario.Setup != nil {
			scenario.Setup(t, &scenario, opts)
		}

		// If a MockFactory function has been provided, then
		// save the original go-fastly API function so it can
		// be restored later, and replace it with the
		// MockFactory's generated mock function
		if scenario.MockFactory != nil {
			originalFunc = *hook
			*hook = scenario.MockFactory(t)
		}

		if len(scenario.Stdin) > 1 {
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

			// We need a channel so we wait for `run()` to complete
			done := make(chan bool)

			// Call `app.Run()` and wait for response
			go func() {
				app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
					return opts, nil
				}
				err = app.Run(fullargs, nil)
				done <- true
			}()

			// User provides input
			//
			// NOTE: Must provide as much input as is expected to be waited on by `run()`.
			//       For example, if `run()` calls `input()` twice, then provide two messages.
			//       Otherwise the select statement will trigger the timeout error.
			for _, input := range scenario.Stdin {
				inputc <- input
			}

			select {
			case <-done:
				// Wait for app.Run() to finish
			case <-time.After(time.Second):
				t.Fatalf("unexpected timeout waiting for mocked prompt inputs to be processed")
			}
		} else {
			stdin := ""
			if len(scenario.Stdin) > 0 {
				stdin = scenario.Stdin[0]
			}
			opts.Input = strings.NewReader(stdin)
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				return opts, nil
			}
			err = app.Run(fullargs, nil)
		}

		AssertErrorContains(t, err, scenario.WantError)
		if scenario.WantRemediation != "" {
			AssertRemediationErrorContains(t, err, scenario.WantRemediation)
		}
		AssertStringContains(t, stdout.String(), scenario.WantOutput)

		for _, want := range scenario.WantOutputs {
			AssertStringContains(t, stdout.String(), want)
		}

		if len(scenario.DontWantOutput) > 0 {
			AssertStringDoesntContain(t, stdout.String(), scenario.DontWantOutput)
		}
		for _, want := range scenario.DontWantOutputs {
			AssertStringDoesntContain(t, stdout.String(), want)
		}

		if scenario.PathContentFlag != nil {
			pcf := *scenario.PathContentFlag
			AssertPathContentFlag(pcf.Flag, scenario.WantError, fullargs, pcf.Fixture, pcf.Content(), t)
		}

		if scenario.Validator != nil {
			scenario.Validator(t, &scenario, opts, &stdout)
		}

		// If a MockFactory function has been provided, then
		// restore the original go-fastly API function to the
		// hook variable
		if scenario.MockFactory != nil {
			*hook = originalFunc
		}
	})
}

// RunAPIHookCLIScenarios executes the APIHookCLIScenario structs from
// the slice passed in.
func RunAPIHookCLIScenarios[T any](t *testing.T, command []string, scenarios []APIHookCLIScenario[T], hook *T) {
	for _, scenario := range scenarios {
		RunAPIHookCLIScenario(t, command, scenario, hook)
	}
}
