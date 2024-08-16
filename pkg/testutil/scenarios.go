package testutil

import (
	"fmt"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/threadsafe"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

// CLIScenario represents a CLI test case to be validated.
//
// Most of the fields in this struct are optional; if they are not
// provided RunCLIScenario will not apply the behavior indicated for
// those fields.
type CLIScenario struct {
	// API is a mock API implementation which can be used by the
	// command under test
	API mock.API
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
	Env             *EnvConfig
	// EnvVars contains environment variables which will be set
	// during the execution of the scenario
	EnvVars map[string]string
	// Name appears in output when tests are executed
	Name            string
	PathContentFlag *PathContentFlag
	// Setup function can perform additional setup before the scenario is run
	Setup func(t *testing.T, scenario *CLIScenario, opts *global.Data)
	// Stdin contains input to be read by the application
	Stdin []string
	// Validator function can perform additional validation on the results
	// of the scenario
	Validator func(t *testing.T, scenario *CLIScenario, opts *global.Data, stdout *threadsafe.Buffer)
	// WantError will cause the scenario to fail if this string
	// does not appear in an Error
	WantError string
	// WantOutput will cause the scenario to fail if this string
	// does not appear in stdout
	WantOutput string
	// WantOutputs will cause the scenario to fail if any of the
	// strings do not appear in stdout
	WantOutputs []string
}

// PathContentFlag provides the details required to validate that a
// flag value has been parsed correctly by the argument parser.
type PathContentFlag struct {
	Flag    string
	Fixture string
	Content func() string
}

// EnvConfig provides the details required to setup a temporary test
// environment, and optionally a function to run which accepts the
// environment directory and can modify fields in the CLIScenario
type EnvConfig struct {
	Opts *EnvOpts
	// EditScenario holds a function which will be called after
	// the temporary environment has been created but before the
	// scenario setup (and execution) begin; it can make any
	// modifications to the CLIScenario that are needed
	EditScenario func(*CLIScenario, string)
}

// RunCLIScenario executes a CLIScenario struct.
// The Arg field of the scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunCLIScenario(t *testing.T, command []string, scenario CLIScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		var (
			err      error
			fullargs []string
			rootdir  string
			stdout   threadsafe.Buffer
		)

		if len(scenario.Args) > 0 {
			fullargs = slices.Concat(command, SplitArgs(scenario.Args))
		} else {
			fullargs = command
		}

		opts := MockGlobalData(fullargs, &stdout)
		opts.APIClientFactory = mock.APIClient(scenario.API)

		if scenario.Env != nil {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create test environment
			scenario.Env.Opts.T = t
			rootdir = NewEnv(*scenario.Env.Opts)
			defer os.RemoveAll(rootdir)

			// Before running the test, chdir into the build environment.
			// When we're done, chdir back to our original location.
			// This is so we can reliably copy the testdata/ fixtures.
			if err := os.Chdir(rootdir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Chdir(pwd)
			}()

			if scenario.Env.EditScenario != nil {
				scenario.Env.EditScenario(&scenario, rootdir)
			}
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
	})
}

// RunCLIScenarios executes the CLIScenario structs from the slice passed in.
func RunCLIScenarios(t *testing.T, command []string, scenarios []CLIScenario) {
	for _, scenario := range scenarios {
		RunCLIScenario(t, command, scenario)
	}
}
