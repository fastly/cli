package websockets

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/websockets"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*websockets.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return websockets.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*websockets.EnableOutput, error) {
		return websockets.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*websockets.EnableOutput, error) {
		return websockets.Get(client.(*fastly.Client), serviceID)
	},
}
