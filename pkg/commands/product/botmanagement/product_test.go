package botmanagement_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/botmanagement"
	"github.com/fastly/go-fastly/v9/fastly/products/botmanagement"
)

func TestBotManagementEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*botmanagement.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   botmanagement.ProductID,
		ProductName: botmanagement.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
