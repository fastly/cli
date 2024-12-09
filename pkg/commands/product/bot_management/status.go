package bot_management

import (
	"errors"
	"io"

	fsterr "github.com/fastly/cli/pkg/errors"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/bot_management"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// GetFn is a dependency-injection point for unit tests to provide
// a mock implementation of the API operation.
var GetFn = func(client api.Interface, serviceID string) (*bot_management.EnableOutput, error) {
	return bot_management.Get(client.(*fastly.Client), serviceID)
}

// StatusCommand calls the Fastly API to get the enablement status of the product.
type StatusCommand struct {
	argparser.Base
	argparser.JSONOutput
	Manifest manifest.Data

	serviceName argparser.OptionalServiceNameID
}

// NewStatusCommand returns a usable command registered under the parent.
func NewStatusCommand(parent argparser.Registerer, g *global.Data) *StatusCommand {
	c := StatusCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("status", "Get the enablement status of the "+bot_management.ProductName+" product")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
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

type status struct {
	Enabled bool `json:"enabled"`
}

// Exec invokes the application logic for the command.
func (c *StatusCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	var s status

	_, err = GetFn(c.Globals.APIClient, serviceID)
	if err != nil {
		var herr *fastly.HTTPError

		// The API returns a 'Bad Request' error when the
		// product has not been enabled on the service; any
		// other error should be reported
		if !errors.As(err, &herr) || !herr.IsBadRequest() {
			c.Globals.ErrLog.Add(err)
			return err
		}
	} else {
		s.Enabled = true
	}

	if ok, err := c.WriteJSON(out, s); ok {
		return err
	}

	var msg string
	if s.Enabled {
		msg = "enabled"
	} else {
		msg = "disabled"
	}

	text.Info(out,
		bot_management.ProductName+" is %s on service %s", msg, serviceID)

	return nil
}
