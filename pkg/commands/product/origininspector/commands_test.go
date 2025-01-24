package origininspector_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/origininspector"
	product "github.com/fastly/go-fastly/v9/fastly/products/origininspector"
)

func TestOrigininspectorEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*product.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   product.ProductID,
		ProductName: product.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
