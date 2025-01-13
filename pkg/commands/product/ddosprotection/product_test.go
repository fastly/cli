package ddosprotection_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/ddosprotection"
	"github.com/fastly/go-fastly/v9/fastly/products/ddosprotection"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*ddosprotection.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   ddosprotection.ProductID,
		ProductName: ddosprotection.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
