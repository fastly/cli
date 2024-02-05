package service

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to create services.
type UpdateCommand struct {
	argparser.Base

	comment     argparser.OptionalString
	input       fastly.UpdateServiceInput
	name        argparser.OptionalString
	serviceName argparser.OptionalServiceNameID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Fastly service")

	// Optional.
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("name", "Service name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.ServiceID = serviceID

	if !c.name.WasSet && !c.comment.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --name or --comment to update service")
	}

	if c.name.WasSet {
		c.input.Name = &c.name.Value
	}
	if c.comment.WasSet {
		c.input.Comment = &c.comment.Value
	}

	s, err := c.Globals.APIClient.UpdateService(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   serviceID,
			"Service Name": c.name.Value,
			"Comment":      c.comment.Value,
		})
		return err
	}

	text.Success(out, "Updated service %s", fastly.ToValue(s.ServiceID))
	return nil
}
