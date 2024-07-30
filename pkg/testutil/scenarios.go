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
	// will be removed once all users are using RunScenarios()
	Args            []string
	ConfigPath      string
	ConfigFile      *config.File
	DontWantOutput  string
	DontWantOutputs []string
	Name            string
	Stdin           []string
	WantError       string
	WantOutput      string
	WantOutputs     []string
	PathContentFlag *PathContentFlag
	SetEnv          *map[string]string
}

// PathContentFlag provides the details required to validate that a
// flag value has been parsed correctly by the argument parser.
type PathContentFlag struct {
	Flag    string
	Fixture string
	Content func() string
}

// RunScenario executes a TestScenario struct.
// The Args field of the scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunScenario(t *testing.T, command []string, scenario TestScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		var (
			err      error
			fullargs []string
			stdout   bytes.Buffer
		)

		if len(scenario.Arg) > 0 {
			fullargs = slices.Concat(command, Args(scenario.Arg))
		} else {
			fullargs = command
		}

		opts := MockGlobalData(fullargs, &stdout)
		opts.APIClientFactory = mock.APIClient(scenario.API)

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
	})
}

// RunScenarios executes one (or more) TestScenario structs from the slice passed in.
// The Args field of each scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunScenarios(t *testing.T, command []string, scenarios []TestScenario) {
	for _, scenario := range scenarios {
		RunScenario(t, command, scenario)
	}
}
