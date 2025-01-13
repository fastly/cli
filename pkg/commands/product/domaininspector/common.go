package domaininspector

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/domaininspector"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*domaininspector.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return domaininspector.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*domaininspector.EnableOutput, error) {
		return domaininspector.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*domaininspector.EnableOutput, error) {
		return domaininspector.Get(client.(*fastly.Client), serviceID)
	},
}
