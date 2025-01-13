package botmanagement

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/botmanagement"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*botmanagement.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return botmanagement.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*botmanagement.EnableOutput, error) {
		return botmanagement.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*botmanagement.EnableOutput, error) {
		return botmanagement.Get(client.(*fastly.Client), serviceID)
	},
}
