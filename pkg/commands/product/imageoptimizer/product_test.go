package imageoptimizer_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/imageoptimizer"
	"github.com/fastly/go-fastly/v9/fastly/products/imageoptimizer"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*imageoptimizer.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   imageoptimizer.ProductID,
		ProductName: imageoptimizer.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
