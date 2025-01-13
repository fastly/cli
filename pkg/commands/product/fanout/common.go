package fanout

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/fanout"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*fanout.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return fanout.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*fanout.EnableOutput, error) {
		return fanout.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*fanout.EnableOutput, error) {
		return fanout.Get(client.(*fastly.Client), serviceID)
	},
}
