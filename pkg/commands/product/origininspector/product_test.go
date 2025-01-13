package origininspector_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/origininspector"
	"github.com/fastly/go-fastly/v9/fastly/products/origininspector"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*origininspector.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   origininspector.ProductID,
		ProductName: origininspector.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
