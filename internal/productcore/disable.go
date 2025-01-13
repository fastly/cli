package productcore

import (
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// Disable is a base type for all 'disable' commands.
type Disable[O any] struct {
	Base
	hooks *EnablementHookFuncs[O]
}

// Init prepares the structure for use by the CLI core.
func (cmd *Disable[O]) Init(parent argparser.Registerer, g *global.Data, productID, productName string, hooks *EnablementHookFuncs[O]) {
	cmd.CmdClause = parent.Command("disable", "Disable the "+productName+" product")
	cmd.hooks = hooks

	cmd.Base.Init(parent, g, productID, productName)
}

// Exec executes the disablement operation.
func (cmd *Disable[O]) Exec(out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(cmd.ServiceName, *cmd.Globals.Manifest, cmd.Globals.APIClient, cmd.Globals.ErrLog)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	err = cmd.hooks.DisableFunc(cmd.Globals.APIClient, serviceID)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, EnablementStatus{ProductID: cmd.ProductID, Enabled: false}); ok {
		return err
	}

	text.Success(out,
		"Disabled %s on service %s", cmd.ProductName, serviceID)

	return nil
}
