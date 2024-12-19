package bot_management

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/bot_management"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

var EnablementHooks = productcore.EnablementHookFns[*bot_management.EnableOutput]{
	DisableFn: func(client api.Interface, serviceID string) error {
		return bot_management.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFn: func(client api.Interface, serviceID string) (*bot_management.EnableOutput, error) {
		return bot_management.Enable(client.(*fastly.Client), serviceID)
	},
	GetFn: func(client api.Interface, serviceID string) (*bot_management.EnableOutput, error) {
		return bot_management.Get(client.(*fastly.Client), serviceID)
	},
}
