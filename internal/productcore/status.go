package productcore

import (
	"io"
	"errors"
	"github.com/fastly/go-fastly/v9/fastly"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// Status is a base type for all 'status' commands.
type Status[O any] struct {
	Base
	hooks *EnablementHookFuncs[O]
}

// Init prepares the structure for use by the CLI core.
func (cmd *Status[O]) Init(parent argparser.Registerer, g *global.Data, productName string, hooks *EnablementHookFuncs[O]) {
	cmd.CmdClause = parent.Command("status", "Get the enablement status of the "+productName+" product")
	cmd.hooks = hooks

	cmd.Base.Init(parent, g, productName)
}

// Exec executes the status operation.
func (cmd *Status[O]) Exec(out io.Writer) error {
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

	s := EnablementStatus{}
	state := "disabled"

	_, err = cmd.hooks.GetFunc(cmd.Globals.APIClient, serviceID)
	if err != nil {
		var herr *fastly.HTTPError

		// The API returns a 'Bad Request' error when the
		// product has not been enabled on the service; any
		// other error should be reported
		if !errors.As(err, &herr) || !herr.IsBadRequest() {
			cmd.Globals.ErrLog.Add(err)
			return err
		}
	} else {
		s.Enabled = true
		state = "enabled"
	}

	if ok, err := cmd.WriteJSON(out, s); ok {
		return err
	}

	text.Info(out,
		"%s is %s on service %s", cmd.ProductName, state, serviceID)

	return nil
}
