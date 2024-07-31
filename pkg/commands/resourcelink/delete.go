package resourcelink

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete service resource links.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	autoClone      argparser.OptionalAutoClone
	input          fastly.DeleteResourceInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a resource link for a Fastly service version").Alias("remove")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "id",
		Description: flagIDDescription,
		Dst:         &c.input.ResourceID,
		Required:    true,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// At least one of the following is required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Short:       's',
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceName,
		Action:      c.serviceName.Set,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AutoCloneFlag:      c.autoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      c.Globals.Manifest.Flag.ServiceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	err = c.Globals.APIClient.DeleteResource(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"ID":              c.input.ResourceID,
			"Service ID":      c.input.ServiceID,
			"Service Version": c.input.ServiceVersion,
		})
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID             string `json:"id"`
			ServiceID      string `json:"service_id"`
			ServiceVersion int    `json:"service_version"`
			Deleted        bool   `json:"deleted"`
		}{
			c.input.ResourceID,
			c.input.ServiceID,
			c.input.ServiceVersion,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted service resource link %s from service %s version %d", c.input.ResourceID, c.input.ServiceID, c.input.ServiceVersion)
	return nil
}
