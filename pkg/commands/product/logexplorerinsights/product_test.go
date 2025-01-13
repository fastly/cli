package logexplorerinsights_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/logexplorerinsights"
	"github.com/fastly/go-fastly/v9/fastly/products/logexplorerinsights"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*logexplorerinsights.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   logexplorerinsights.ProductID,
		ProductName: logexplorerinsights.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
