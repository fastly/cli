package productcore

import (
	"io"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// Enable is a base type for all 'enable' commands.
type Enable[O any] struct {
	Base
	hooks *HookFns[O]
}

// Init prepares the structure for use by the CLI core.
func (cmd *Enable[O]) Init(parent argparser.Registerer, g *global.Data, productName string, hooks *HookFns[O]) {
	cmd.CmdClause = parent.Command("enable", "Enable the "+productName+" product")
	cmd.hooks = hooks

	cmd.Base.Init(parent, g, productName)
}

// Exec executes the disablement operation.
func (cmd *Enable[O]) Exec(out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(cmd.ServiceName, *cmd.Globals.Manifest, cmd.Globals.APIClient, cmd.Globals.ErrLog)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if cmd.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	_, err = cmd.hooks.EnableFn(cmd.Globals.APIClient, serviceID)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Enabled "+cmd.ProductName+" on service %s", serviceID)

	return nil
}
