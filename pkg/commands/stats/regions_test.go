package stats_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fastly/go-fastly/v12/fastly"

	root "github.com/fastly/cli/pkg/commands/stats"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestRegions(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:       "",
			API:        mock.API{GetRegionsFn: getRegionsOK},
			WantOutput: "foo\nbar\nbaz\n",
		},
		{
			Args:      "",
			API:       mock.API{GetRegionsFn: getRegionsError},
			WantError: errTest.Error(),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, "regions"}, scenarios)
}

func getRegionsOK(_ context.Context) (*fastly.RegionsResponse, error) {
	return &fastly.RegionsResponse{
		Data: []string{"foo", "bar", "baz"},
	}, nil
}

var errTest = errors.New("fixture error")

func getRegionsError(_ context.Context) (*fastly.RegionsResponse, error) {
	return nil, errTest
}
