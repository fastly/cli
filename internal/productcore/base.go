package productcore

import (
	"fmt"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v9/fastly/products"
)

// Base is a base type for all product commands.
type Base struct {
	argparser.Base
	argparser.JSONOutput
	Manifest manifest.Data

	ServiceName argparser.OptionalServiceNameID
}

// Init prepares the structure for use by the CLI core.
func (cmd *Base) Init(parent argparser.Registerer, g *global.Data) {
	cmd.Globals = g

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

// EnablementStatus is a structure used to generate output from
// the enablement-related commands
type EnablementStatus[_ products.ProductOutput] struct {
	ProductName string `json:"-"`
	ProductID   string `json:"product_id"`
	ServiceID   string `json:"service_id"`
	Enabled     bool   `json:"enabled"`
}

type StatusManager[O products.ProductOutput] interface {
	SetEnabled(bool)
	GetEnabled() string
	GetProductName() string
	SetProductID(string)
	SetServiceID(string)
	TransformOutput(O)
	GetTextResult() string
}

func (s *EnablementStatus[_]) SetEnabled(e bool) {
	s.Enabled = e
}

func (s *EnablementStatus[_]) GetEnabled() string {
	if s.Enabled {
		return "enabled"
	}
	return "disabled"
}

func (s *EnablementStatus[_]) GetProductName() string {
	return s.ProductName
}

func (s *EnablementStatus[_]) SetProductID(id string) {
	s.ProductID = id
}

func (s *EnablementStatus[_]) SetServiceID(id string) {
	s.ServiceID = id
}

func (s *EnablementStatus[O]) TransformOutput(o O) {
	s.ProductID = o.ProductID()
	s.ServiceID = o.ServiceID()
}

func (s *EnablementStatus[O]) GetTextResult() string {
	return fmt.Sprintf("%s is %s on service %s", s.ProductName, s.GetEnabled(), s.ServiceID)
}

// EnablementHookFuncs is a structure of dependency-injection points
// used by unit tests to provide mock behaviors
type EnablementHookFuncs[O products.ProductOutput] struct {
	DisableFunc func(api.Interface, string) error
	EnableFunc  func(api.Interface, string) (O, error)
	GetFunc     func(api.Interface, string) (O, error)
}

// ConfigurationHookFuncs is a structure of dependency-injection
// points used by unit tests to provide mock behaviors
type ConfigurationHookFuncs[O, I any] struct {
	GetConfigurationFunc    func(api.Interface, string) (O, error)
	UpdateConfigurationFunc func(api.Interface, string, I) (O, error)
}
