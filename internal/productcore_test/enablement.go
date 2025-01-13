package productcore_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

type TestEnablementInput[O any] struct {
	T *testing.T
	Commands []string
	ProductName string
	Hooks *productcore.EnablementHookFuncs[O]
}

func TestEnablement[O any](i TestEnablementInput[O]) {
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
			Name:      "validate invalid json/verbose flag combo: enable",
			Args:      "enable --service-id 123 --json --verbose",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name:      "validate invalid json/verbose flag combo: disable",
			Args:      "disable --service-id 123 --json --verbose",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name:      "validate invalid json/verbose flag combo: status",
			Args:      "status --service-id 123 --json --verbose",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate text output success for enabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.EnableFunc = func(_ api.Interface, _ string) (o O, err error) {
					return
				}
			},
			Args:       "enable --service-id 123",
			WantOutput: "SUCCESS: Enabled " + i.ProductName + " on service 123",
		},
		{
			Name: "validate JSON output success for enabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.EnableFunc = func(_ api.Interface, _ string) (o O, err error) {
					return
				}
			},
			Args:       "enable --service-id 123 --json",
			WantOutput: "{\n  \"enabled\": true\n}",
		},
		{
			Name: "validate failure for enabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.EnableFunc = func(_ api.Interface, _ string) (o O, err error) {
					err = testutil.Err
					return
				}
			},
			Args:      "enable --service-id 123",
			WantError: "test error",
		},
		{
			Name: "validate text output success for disabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.DisableFunc = func(_ api.Interface, _ string) error {
					return nil
				}
			},
			Args:       "disable --service-id 123",
			WantOutput: "SUCCESS: Disabled " + i.ProductName + " on service 123",
		},
		{
			Name: "validate JSON output success for disabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.DisableFunc = func(_ api.Interface, _ string) error {
					return nil
				}
			},
			Args:       "disable --service-id 123 --json",
			WantOutput: "{\n  \"enabled\": false\n}",
		},
		{
			Name: "validate failure for disabling product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.DisableFunc = func(_ api.Interface, _ string) error {
					return testutil.Err
				}
			},
			Args:      "disable --service-id 123",
			WantError: "test error",
		},
		{
			Name: "validate text status output for enabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.GetFunc = func(_ api.Interface, _ string) (o O, err error) {
					return
				}
			},
			Args:       "status --service-id 123",
			WantOutput: "INFO: " + i.ProductName + " is enabled on service 123",
		},
		{
			Name: "validate JSON status output for enabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.GetFunc = func(_ api.Interface, _ string) (o O, err error) {
					return
				}
			},
			Args:       "status --service-id 123 --json",
			WantOutput: "{\n  \"enabled\": true\n}",
		},
		{
			Name: "validate text status output for disabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.GetFunc = func(_ api.Interface, _ string) (o O, err error) {
					// The API returns a 'Bad Request' error when the
					// product has not been enabled on the service
					err =  &fastly.HTTPError{StatusCode: 400}
					return
				}
			},
			Args:       "status --service-id 123",
			WantOutput: "INFO: " + i.ProductName + " is disabled on service 123",
		},
		{
			Name: "validate JSON status output for disabled product",
			Setup: func(t *testing.T, scenario *testutil.CLIScenario, opts *global.Data) {
				i.Hooks.GetFunc = func(_ api.Interface, _ string) (o O, err error) {
					// The API returns a 'Bad Request' error when the
					// product has not been enabled on the service
					err =  &fastly.HTTPError{StatusCode: 400}
					return
				}
			},
			Args:       "status --service-id 123 --json",
			WantOutput: "{\n  \"enabled\": false\n}",
		},
	}

	testutil.RunCLIScenarios(i.T, i.Commands, scenarios)
}
