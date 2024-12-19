package productcore

import (
	"io"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DisableFn is the type of the function that will be used to perform
// the disablement.
type DisableFn func(api.Interface, string) error

// Disable is a base type for all 'disable' commands.
type Disable struct {
	Base
}

// Init prepares the structure for use by the CLI core.
func (cmd *Disable) Init(parent argparser.Registerer, g *global.Data, productName string) {
	cmd.CmdClause = parent.Command("disable", "Disable the "+productName+" product")

	cmd.Base.Init(parent, g, productName)
}

// Exec executes the disablement operation.
func (cmd *Disable) Exec(out io.Writer, op DisableFn) error {
	serviceID, source, flag, err := argparser.ServiceID(cmd.ServiceName, *cmd.Globals.Manifest, cmd.Globals.APIClient, cmd.Globals.ErrLog)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	err = op(cmd.Globals.APIClient, serviceID)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Disabled "+cmd.ProductName+" on service %s", serviceID)

	return nil
}
