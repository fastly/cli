package brotlicompression_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/brotlicompression"
	"github.com/fastly/go-fastly/v9/fastly/products/brotlicompression"
)

func TestProductEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*brotlicompression.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   brotlicompression.ProductID,
		ProductName: brotlicompression.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
