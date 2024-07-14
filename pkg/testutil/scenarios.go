package testutil

import (
	"bytes"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"io"
	"slices"
	"testing"
)

// TestScenario represents a standard test case to be validated.
type TestScenario struct {
	API mock.API
	Arg string
	// will be removed once all users are using RunScenarios()
	Args            []string
	DontWantOutput  string
	DontWantOutputs []string
	Name            string
	WantError       string
	WantOutput      string
	WantOutputs     []string
}

// RunScenario executes a TestScenario struct.
// The Args field of the scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunScenario(t *testing.T, command []string, scenario TestScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		var stdout bytes.Buffer
		var fullargs []string
		if len(scenario.Arg) > 0 {
			fullargs = slices.Concat(command, Args(scenario.Arg))
		} else {
			fullargs = command
		}
		app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
			opts := MockGlobalData(fullargs, &stdout)
			opts.APIClientFactory = mock.APIClient(scenario.API)
			return opts, nil
		}
		err := app.Run(fullargs, nil)
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
