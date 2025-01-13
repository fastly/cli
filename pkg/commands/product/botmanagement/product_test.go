package botmanagement_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/botmanagement"

	"github.com/fastly/cli/pkg/api"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/botmanagement"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestProductEnablement(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "validate missing Service ID: enable",
			Args:      "enable",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing Service ID: disable",
			Args:      "enable",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate missing Service ID: status",
			Args:      "enable",
			WantError: "error reading service: no service ID found",
		},
		{
			Name:      "validate invalid json/verbose flag combo: status",
			Args:      "status --service-id 123 --json --verbose",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate success for enabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.EnableFunc = func(_ api.Interface, _ string) (*botmanagement.EnableOutput, error) {
					return nil, nil
				}
			},
			Args:       "enable --service-id 123",
			WantOutput: "SUCCESS: Enabled " + botmanagement.ProductName + " on service 123",
		},
		{
			Name: "validate failure for enabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.EnableFunc = func(_ api.Interface, _ string) (*botmanagement.EnableOutput, error) {
					return nil, testutil.Err
				}
			},
			Args:      "enable --service-id 123",
			WantError: "test error",
		},
		{
			Name: "validate success for disabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.DisableFunc = func(_ api.Interface, _ string) error {
					return nil
				}
			},
			Args:       "disable --service-id 123",
			WantOutput: "SUCCESS: Disabled " + botmanagement.ProductName + " on service 123",
		},
		{
			Name: "validate failure for disabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.DisableFunc = func(_ api.Interface, _ string) error {
					return testutil.Err
				}
			},
			Args:      "disable --service-id 123",
			WantError: "test error",
		},
		{
			Name: "validate regular status output for enabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.GetFunc = func(_ api.Interface, _ string) (*botmanagement.EnableOutput, error) {
					return nil, nil
				}
			},
			Args:       "status --service-id 123",
			WantOutput: "INFO: " + botmanagement.ProductName + " is enabled on service 123",
		},
		{
			Name: "validate JSON status output for enabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.GetFunc = func(_ api.Interface, _ string) (*botmanagement.EnableOutput, error) {
					return nil, nil
				}
			},
			Args:       "status --service-id 123 --json",
			WantOutput: "{\n  \"enabled\": true\n}",
		},
		{
			Name: "validate regular status output for disabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.GetFunc = func(_ api.Interface, _ string) (*botmanagement.EnableOutput, error) {
					// The API returns a 'Bad Request' error when the
					// product has not been enabled on the service
					return nil, &fastly.HTTPError{StatusCode: 400}
				}
			},
			Args:       "status --service-id 123",
			WantOutput: "INFO: " + botmanagement.ProductName + " is disabled on service 123",
		},
		{
			Name: "validate JSON status output for disabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				sub.EnablementHooks.GetFunc = func(_ api.Interface, _ string) (*botmanagement.EnableOutput, error) {
					// The API returns a 'Bad Request' error when the
					// product has not been enabled on the service
					return nil, &fastly.HTTPError{StatusCode: 400}
				}
			},
			Args:       "status --service-id 123 --json",
			WantOutput: "{\n  \"enabled\": false\n}",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName}, scenarios)
}
