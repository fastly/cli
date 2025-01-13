package logexplorerinsights

import (
	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/logexplorerinsights"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[*logexplorerinsights.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return logexplorerinsights.Disable(client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (*logexplorerinsights.EnableOutput, error) {
		return logexplorerinsights.Enable(client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (*logexplorerinsights.EnableOutput, error) {
		return logexplorerinsights.Get(client.(*fastly.Client), serviceID)
	},
}
