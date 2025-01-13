package domaininspector_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/domaininspector"
	"github.com/fastly/go-fastly/v9/fastly/products/domaininspector"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*domaininspector.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   domaininspector.ProductID,
		ProductName: domaininspector.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
