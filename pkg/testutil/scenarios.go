package testutil

import "github.com/fastly/cli/pkg/mock"

// TestScenario represents a standard test case to be validated.
type TestScenario struct {
	API        mock.API
	Args       []string
	Name       string
	WantError  string
	WantOutput string
}
