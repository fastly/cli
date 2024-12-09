package bot_management

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/bot_management"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DisableFn is a dependency-injection point for unit tests to provide
// a mock implementation of the API operation.
var DisableFn = func(client api.Interface, serviceID string) error {
	return bot_management.Disable(client.(*fastly.Client), serviceID)
}

// DisableCommand calls the Fastly API to disable the product.
type DisableCommand struct {
	argparser.Base
	Manifest manifest.Data

	serviceName argparser.OptionalServiceNameID
}

// NewDisableCommand returns a usable command registered under the parent.
func NewDisableCommand(parent argparser.Registerer, g *global.Data) *DisableCommand {
	c := DisableCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("disable", "Disable the "+bot_management.ProductName+" product")

	// Optional.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DisableCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	err = DisableFn(c.Globals.APIClient, serviceID)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Disabled "+bot_management.ProductName+" on service %s", serviceID)

	return nil
}
