package brotlicompression

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/brotlicompression"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*brotlicompression.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return brotlicompression.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*brotlicompression.EnableOutput, error) {
		return brotlicompression.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*brotlicompression.EnableOutput, error) {
		return brotlicompression.Get(client.(*fastly.Client), serviceID)
	},
}
