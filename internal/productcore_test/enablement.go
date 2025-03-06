package productcore_test

import (
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

const TestServiceID = "123"

type CommandRequiredArguments struct {
	Command string
	Arguments []string
}

func MissingServiceIDScenarios(cra []CommandRequiredArguments) (scenarios []testutil.CLIScenario) {
	for _, v := range cra {
		scenario := testutil.CLIScenario{}

		scenario.Name = "validate missing Service ID: " + v.Command
		scenario.Args = v.Command
		for _, a := range v.Arguments {
			if (!strings.HasPrefix(a, "--service-id")) {
				scenario.Args += " " + a
			}
		}
		scenario.WantError = "error reading service: no service ID found"

		scenarios = append(scenarios, scenario)
	}

	return
}

func InvalidJSONVerboseScenarios(cra []CommandRequiredArguments) (scenarios []testutil.CLIScenario) {
	for _, v := range cra {
		scenario := testutil.CLIScenario{}

		scenario.Name = "validate invalid json/verbose flag combo: " + v.Command
		scenario.Args = v.Command
		for _, a := range v.Arguments {
			scenario.Args += " " + a
		}
		scenario.Args += " --json --verbose"
		scenario.WantError = "invalid flag combination, --verbose and --json"

		scenarios = append(scenarios, scenario)
	}

	return
}

func EnableScenarios[O products.ProductOutput](cra []CommandRequiredArguments, productID, productName string, hooks *productcore.EnablementHookFuncs[O], mocker func(string) O) (scenarios []testutil.CLIScenario) {
	for _, v := range cra {
		if v.Command != "enable" {
			continue
		}

		scenario := testutil.CLIScenario{}

		scenario.Name = "validate text output success for enabling product"
		scenario.Args = v.Command
		for _, a := range v.Arguments {
			scenario.Args += " " + a
		}
		scenario.Setup = func(_ *testing.T, _ *testutil.CLIScenario, _ *global.Data) {
			hooks.EnableFunc = func(_ api.Interface, serviceID string) (O, error) {
				return mocker(serviceID), nil
			}
		}
		scenario.WantOutput = "SUCCESS: " + productName + " is enabled on service " + TestServiceID
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate JSON output success for enabling product"
		scenario.Args += " --json"
		scenario.WantOutput = "{\n  \"product_id\": \"" + productID + "\",\n  \"service_id\": \"" + TestServiceID + "\",\n  \"enabled\": true\n}"
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate failure for enabling product"
		scenario.Setup = func(_ *testing.T, _ *testutil.CLIScenario, _ *global.Data) {
			hooks.EnableFunc = func(_ api.Interface, serviceID string) (O, error) {
				return mocker(serviceID), testutil.Err
			}
		}
		scenario.WantOutput = ""
		scenario.WantError = "test error"
		scenarios = append(scenarios, scenario)
	}

	return
}

func DisableScenarios[O products.ProductOutput](cra []CommandRequiredArguments, productID, productName string, hooks *productcore.EnablementHookFuncs[O]) (scenarios []testutil.CLIScenario) {
	for _, v := range cra {
		if v.Command != "disable" {
			continue
		}

		scenario := testutil.CLIScenario{}

		scenario.Name = "validate text output success for disabling product"
		scenario.Args = v.Command
		for _, a := range v.Arguments {
			scenario.Args += " " + a
		}
		scenario.Setup = func(_ *testing.T, _ *testutil.CLIScenario, _ *global.Data) {
			hooks.DisableFunc = func(_ api.Interface, serviceID string) error {
				return nil
			}
		}
		scenario.WantOutput = "SUCCESS: " + productName + " is disabled on service " + TestServiceID
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate JSON output success for disabling product"
		scenario.Args += " --json"
		scenario.WantOutput = "{\n  \"product_id\": \"" + productID + "\",\n  \"service_id\": \"" + TestServiceID + "\",\n  \"enabled\": false\n}"
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate failure for disabling product"
		scenario.Setup = func(_ *testing.T, _ *testutil.CLIScenario, _ *global.Data) {
			hooks.DisableFunc = func(_ api.Interface, serviceID string) error {
				return testutil.Err
			}
		}
		scenario.WantOutput = ""
		scenario.WantError = "test error"
		scenarios = append(scenarios, scenario)
	}

	return
}

func StatusScenarios[O products.ProductOutput](cra []CommandRequiredArguments, productID, productName string, hooks *productcore.EnablementHookFuncs[O], mocker func(string) O) (scenarios []testutil.CLIScenario) {
	for _, v := range cra {
		if v.Command != "statu" {
			continue
		}

		scenario := testutil.CLIScenario{}

		scenario.Name = "validate text output for enabled product"
		scenario.Args = v.Command
		for _, a := range v.Arguments {
			scenario.Args += " " + a
		}
		scenario.Setup = func(_ *testing.T, _ *testutil.CLIScenario, _ *global.Data) {
			hooks.GetFunc = func(_ api.Interface, serviceID string) (O, error) {
				return mocker(serviceID), nil
			}
		}
		scenario.WantOutput = "INFO: " + productName + " is enabled on service " + TestServiceID
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate JSON output for enabled product"
		scenario.Args += " --json"
		scenario.WantOutput = "{\n  \"product_id\": \"" + productID + "\",\n  \"service_id\": \"" + TestServiceID + "\",\n  \"enabled\": true\n}"
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate text output for disabled product"
		scenario.Args = v.Command
		for _, a := range v.Arguments {
			scenario.Args += " " + a
		}
		scenario.Setup = func(_ *testing.T, _ *testutil.CLIScenario, _ *global.Data) {
			hooks.GetFunc = func(_ api.Interface, serviceID string) (O, error) {
				// The API returns a 'Bad Request' error when the
				// product has not been enabled on the service
				return mocker(serviceID), &fastly.HTTPError{StatusCode: 400}
			}
		}
		scenario.WantOutput = "INFO: " + productName + " is disabled on service " + TestServiceID
		scenarios = append(scenarios, scenario)

		scenario.Name = "validate JSON output for disabled product"
		scenario.Args += " --json"
		scenario.WantOutput = "{\n  \"product_id\": \"" + productID + "\",\n  \"service_id\": \"" + TestServiceID + "\",\n  \"enabled\": false\n}"
		scenarios = append(scenarios, scenario)
	}

	return
}
