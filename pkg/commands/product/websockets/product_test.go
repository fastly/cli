package websockets_test

import (
	"testing"

	"github.com/fastly/cli/internal/productcore_test"
	root "github.com/fastly/cli/pkg/commands/product"
	sub "github.com/fastly/cli/pkg/commands/product/websockets"
	"github.com/fastly/go-fastly/v9/fastly/products/websockets"
)

func TestWebSocketsEnablement(t *testing.T) {
	productcore_test.TestEnablement(productcore_test.TestEnablementInput[*websockets.EnableOutput]{
		T:           t,
		Commands:    []string{root.CommandName, sub.CommandName},
		ProductID:   websockets.ProductID,
		ProductName: websockets.ProductName,
		Hooks:       &sub.EnablementHooks,
	})
}
