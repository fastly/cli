package vcl_test

import (
	"context"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"

	root "github.com/fastly/cli/pkg/commands/service"
	sub "github.com/fastly/cli/pkg/commands/service/vcl"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVCLDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing --version flag",
			WantError: "error parsing arguments: required flag --version not provided",
		},
		{
			Name:      "validate missing --service-id flag",
			Args:      "--version 3",
			WantError: "error reading service: no service ID found",
		},
		{
			Name: "validate DescribeVCL API error",
			API: mock.API{
				ListVersionsFn: testutil.ListVersions,
				GetGeneratedVCLFn: func(_ context.Context, _ *fastly.GetGeneratedVCLInput) (*fastly.VCL, error) {
					return nil, testutil.Err
				},
			},
			Args:      "--service-id 123 --version 3",
			WantError: testutil.Err.Error(),
		},
		{
			Name: "validate DescribeVCL API success",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetGeneratedVCLFn: getVCL,
			},
			Args:       "--service-id 123 --version 3",
			WantOutput: "# some vcl content\n",
		},
		{
			Name: "validate missing --verbose flag",
			API: mock.API{
				ListVersionsFn:    testutil.ListVersions,
				GetGeneratedVCLFn: getVCL,
			},
			Args:       "--service-id 123 --verbose --version 1",
			WantOutput: "Fastly API endpoint: https://api.fastly.com\nFastly API token provided via config file (profile: user)\n\nService ID (via --service-id): 123\n\nService Version: 1\n\nName: foo\nMain: false\nCreated at: 2021-06-15 23:00:00 +0000 UTC\nUpdated at: 2021-06-15 23:00:00 +0000 UTC\nDeleted at: 2021-06-15 23:00:00 +0000 UTC\nContent: \n# some vcl content\n",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func getVCL(_ context.Context, i *fastly.GetGeneratedVCLInput) (*fastly.VCL, error) {
	t := testutil.Date

	return &fastly.VCL{
		Content:        fastly.ToPointer("# some vcl content"),
		Main:           fastly.ToPointer(false),
		Name:           fastly.ToPointer("foo"),
		ServiceID:      fastly.ToPointer(i.ServiceID),
		ServiceVersion: fastly.ToPointer(i.ServiceVersion),
		CreatedAt:      &t,
		DeletedAt:      &t,
		UpdatedAt:      &t,
	}, nil
}
