package testutil

import (
	"bytes"
	"fmt"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestScenario represents a standard test case to be validated.
type TestScenario struct {
	API mock.API
	Arg string
	// FIXME: remove once all users are using RunScenario() or RunScenarios()
	Args []string
	// this will be copied into global.Data.ConfigPath
	ConfigPath string
	// this will be copied into global.Data.ConfigFile
	ConfigFile *config.File
	// fail the scenario if this string appears in stdout
	DontWantOutput string
	// fail the scenario if any of these strings appear in stdout
	DontWantOutputs []string
	// scenario name (appears in output when tests are executed)
	Name string
	// if this is supplied, NewEnv will be used to create a temporary environment
	// for the scenario
	NewEnv          *NewEnvConfig
	PathContentFlag *PathContentFlag
	// set one (or more) environment variables during the scenario's execution
	SetEnv *map[string]string
	// input to be read by the application
	Stdin []string
	// this function can perform additional validation on the results of the scenario
	Validator func(t *testing.T, scenario *TestScenario, opts *global.Data, err error, stdout bytes.Buffer)
	// fail the scenario if this string does not appear in an Error
	WantError string
	// fail the scenario if this string does not appear in stdout
	WantOutput string
	// fail the scenario if any of these strings do not appear in stdout
	WantOutputs []string
}

// PathContentFlag provides the details required to validate that a
// flag value has been parsed correctly by the argument parser.
type PathContentFlag struct {
	Flag    string
	Fixture string
	Content func() string
}

// NewEnvConfig provides the details required to setup a temporary test
// environment, and optionally a function to run which accepts the
// environment directory and can modify fields in the TestScenario
type NewEnvConfig struct {
	EnvOpts      *EnvOpts
	EditScenario func(*TestScenario, string)
}

// RunScenario executes a TestScenario struct.
// The Arg field of the scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunScenario(t *testing.T, command []string, scenario TestScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		var (
			err      error
			fullargs []string
			rootdir  string
			stdout   bytes.Buffer
		)

		if len(scenario.Arg) > 0 {
			fullargs = slices.Concat(command, Args(scenario.Arg))
		} else {
			fullargs = command
		}

		opts := MockGlobalData(fullargs, &stdout)
		opts.APIClientFactory = mock.APIClient(scenario.API)

		if scenario.NewEnv != nil {
			// We're going to chdir to a deploy environment,
			// so save the PWD to return to, afterwards.
			pwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			// Create test environment
			scenario.NewEnv.EnvOpts.T = t
			rootdir = NewEnv(*scenario.NewEnv.EnvOpts)
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

			if scenario.NewEnv.EditScenario != nil {
				scenario.NewEnv.EditScenario(&scenario, rootdir)
			}
		}

		if len(scenario.ConfigPath) > 0 {
			opts.ConfigPath = scenario.ConfigPath
		}

		if scenario.ConfigFile != nil {
			opts.Config = *scenario.ConfigFile
		}

		if scenario.SetEnv != nil {
			for key, value := range *scenario.SetEnv {
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
				err = app.Run(scenario.Args, nil)
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
			scenario.Validator(t, &scenario, opts, err, stdout)
		}
	})
}

// RunScenarios executes the TestScenario structs from the slice passed in.
func RunScenarios(t *testing.T, command []string, scenarios []TestScenario) {
	for _, scenario := range scenarios {
		RunScenario(t, command, scenario)
	}
}
