package fanout_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/fanout"
	"github.com/fastly/go-fastly/v9/fastly/products/fanout"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*fanout.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   fanout.ProductID,
		ProductName: fanout.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
