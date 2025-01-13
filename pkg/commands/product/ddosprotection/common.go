package ddosprotection

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/ddosprotection"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*ddosprotection.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return ddosprotection.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*ddosprotection.EnableOutput, error) {
		return ddosprotection.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*ddosprotection.EnableOutput, error) {
		return ddosprotection.Get(client.(*fastly.Client), serviceID)
	},
}
