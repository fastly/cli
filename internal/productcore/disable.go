package productcore

import (
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v9/fastly/products"
)

// Disable is a base type for all 'disable' commands.
type Disable[O products.ProductOutput, _ StatusManager[O]] struct {
	Base
	ProductID string
	// hooks is a pointer to an EnablementHookFuncs structure so
	// that tests can modify the contents of the structure after
	// this structure has been initialized
	hooks *EnablementHookFuncs[O]
}

// Init prepares the structure for use by the CLI core.
func (cmd *Disable[O, _]) Init(parent argparser.Registerer, g *global.Data, productID, productName string, hooks *EnablementHookFuncs[O]) {
	cmd.CmdClause = parent.Command("disable", "Disable the "+productName+" product")
	cmd.hooks = hooks

	cmd.Base.Init(parent, g)
	cmd.ProductID = productID
}

// Exec executes the disablement operation.
func (cmd *Disable[O, S]) Exec(out io.Writer, status S) error {
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

	status.SetEnabled(false)
	// The API does not return details of the service and product
	// which were disabled, so they have to be inserted into
	// 'status' directly
	status.SetProductID(cmd.ProductID)
	status.SetServiceID(serviceID)

	if ok, err := cmd.WriteJSON(out, status); ok {
		return err
	}

	text.Success(out, status.GetTextResult())

	return nil
}
