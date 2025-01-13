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
	argparser.JSONOutput
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
	cmd.RegisterFlagBool(cmd.JSONFlag()) // --json
}

// EnablementStatus is a structure used to generate JSON output from
// the enablement-related commands
type EnablementStatus struct {
	Enabled bool `json:"enabled"`
}

// EnablementHookFuncs is a structure of dependency-injection points
// used by unit tests to provide mock behaviors
type EnablementHookFuncs[O any] struct {
	DisableFunc func(api.Interface, string) error
	EnableFunc func(api.Interface, string) (O, error)
	GetFunc func(api.Interface, string) (O, error)
}

// ConfigurationHookFuncs is a structure of dependency-injection
// points used by unit tests to provide mock behaviors
type ConfigurationHookFuncs[O, I any] struct {
	GetConfigurationFunc func(api.Interface, string) (O, error)
	UpdateConfigurationFunc func(api.Interface, string, I) (O, error)
}
