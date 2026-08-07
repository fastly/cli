package botmanagement_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/botmanagement"
	product "github.com/fastly/go-fastly/v11/fastly/products/botmanagement"
	"github.com/fastly/cli/pkg/testutil"
)

var CRA = []productcore_test.CommandRequiredArguments{
	{
		Command: "enable",
		Arguments: []string{"--service-id " + productcore_test.TestServiceID},
	},
	{
		Command: "disable",
		Arguments: []string{"--service-id " + productcore_test.TestServiceID},
	},
	{
		Command: "status",
		Arguments: []string{"--service-id " + productcore_test.TestServiceID},
	},
}

func MockEnableOutput(serviceID string) product.EnableOutput {
	return product.NewEnableOutput(serviceID)
}

func TestBotManagementEnablement(t *testing.T) {
	scenarios := productcore_test.MissingServiceIDScenarios(CRA)
	scenarios = append(scenarios, productcore_test.InvalidJSONVerboseScenarios(CRA)...)
	scenarios = append(scenarios, productcore_test.EnableScenarios(CRA, product.ProductID, product.ProductName, &sub.EnablementHooks, MockEnableOutput)...)
	scenarios = append(scenarios, productcore_test.DisableScenarios(CRA, product.ProductID, product.ProductName, &sub.EnablementHooks)...)
	scenarios = append(scenarios, productcore_test.StatusScenarios(CRA, product.ProductID, product.ProductName, &sub.EnablementHooks, MockEnableOutput)...)

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName}, scenarios)
}
