package productcore

import (
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
)

// Base is a base type for all product commands.
type Base struct {
	argparser.Base
	Manifest manifest.Data

	ServiceName argparser.OptionalServiceNameID
	ProductName string
}

// Init prepares the structure for use by the CLI core.
func (cmd *Base) Init(parent argparser.Registerer, g *global.Data, productName string) {
	cmd.Globals = g
	cmd.ProductName = productName

	// Optional flags.
	cmd.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	cmd.RegisterFlag(argparser.StringFlagOpts{
		Action:      cmd.ServiceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &cmd.ServiceName.Value,
	})
}

type EnablementHookFns[O any] struct {
	DisableFn func(api.Interface, string) error
	EnableFn func(api.Interface, string) (O, error)
	GetFn func(api.Interface, string) (O, error)
}

type ConfigurationHookFns[O, I any] struct {
	GetConfigurationFn func(api.Interface, string) (O, error)
	UpdateConfigurationFn func(api.Interface, string, I) (O, error)
}
