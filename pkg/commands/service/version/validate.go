package version

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// ValidateCommand calls the Fastly API to validate a service version.
type ValidateCommand struct {
	argparser.Base
	argparser.JSONOutput

	input          fastly.ValidateVersionInput
	serviceVersion argparser.OptionalServiceVersion
}

// NewValidateCommand returns a usable command registered under the parent.
func NewValidateCommand(parent argparser.Registerer, g *global.Data) *ValidateCommand {
	c := ValidateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.CmdClause = parent.Command("validate", "Validate a service version")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ValidateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}
	/*if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}*/

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	valid, msg, err := c.Globals.APIClient.ValidateVersion(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, map[string]any{
		"valid":   valid,
		"message": msg,
	}); ok {
		return err
	}

	fmt.Fprintf(out, "\nService ID: %s\n", serviceID)
	fmt.Fprintf(out, "Service Version: %d\n", c.input.ServiceVersion)
	fmt.Fprintf(out, "Valid: %t\n", valid)
	if msg != "" {
		fmt.Fprintf(out, "Message: %s\n", msg)
	}
	fmt.Fprintln(out)

	return nil
}
