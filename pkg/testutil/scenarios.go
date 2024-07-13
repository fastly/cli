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

// RunScenarios executes one (or more) TestScenario structs from the slice passed in.
// The Args field of each scenario is prepended with the content of the 'command'
// slice passed in to construct the complete command to be executed.
func RunScenarios(t *testing.T, command []string, scenarios []TestScenario) {
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			var fullargs []string
			if len(testcase.Arg) > 0 {
				fullargs = slices.Concat(command, Args(testcase.Arg))
			} else {
				fullargs = command
			}
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := MockGlobalData(fullargs, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(fullargs, nil)
			AssertErrorContains(t, err, testcase.WantError)
			AssertStringContains(t, stdout.String(), testcase.WantOutput)
			for _, want := range testcase.WantOutputs {
				AssertStringContains(t, stdout.String(), want)
			}
			if len(testcase.DontWantOutput) > 0 {
				AssertStringDoesntContain(t, stdout.String(), testcase.DontWantOutput)
			}
			for _, want := range testcase.DontWantOutputs {
				AssertStringDoesntContain(t, stdout.String(), want)
			}
		})
	}
}
