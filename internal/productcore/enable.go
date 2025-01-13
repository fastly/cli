package productcore

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// Enable is a base type for all 'enable' commands.
type Enable[O any] struct {
	Base
	// hooks is a pointer to an EnablementHookFuncs structure so
	// that tests can modify the contents of the structure after
	// this structure has been initialized
	hooks *EnablementHookFuncs[O]
}

// Init prepares the structure for use by the CLI core.
func (cmd *Enable[O]) Init(parent argparser.Registerer, g *global.Data, productID, productName string, hooks *EnablementHookFuncs[O]) {
	cmd.CmdClause = parent.Command("enable", "Enable the "+productName+" product")
	cmd.hooks = hooks

	cmd.Base.Init(parent, g, productID, productName)
}

// Exec executes the enablement operation.
func (cmd *Enable[O]) Exec(out io.Writer) error {
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

	_, err = cmd.hooks.EnableFunc(cmd.Globals.APIClient, serviceID)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, EnablementStatus{ProductID: cmd.ProductID, Enabled: true}); ok {
		return err
	}

	text.Success(out,
		"Enabled %s on service %s", cmd.ProductName, serviceID)

	return nil
}
