package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// UpdateCommand calls the Fastly API to create services.
type UpdateCommand struct {
	cmd.Base

	comment     cmd.OptionalString
	input       fastly.UpdateServiceInput
	manifest    manifest.Data
	name        cmd.OptionalString
	serviceName cmd.OptionalServiceNameID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("update", "Update a Fastly service")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceNameFlag(c.serviceName.Set, &c.serviceName.Value)
	c.CmdClause.Flag("name", "Service name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		var err error
		if !c.serviceName.WasSet {
			err = errors.ErrNoServiceID
			c.Globals.ErrLog.Add(err)
			return err
		}
		serviceID, err = c.serviceName.Parse(c.Globals.Client)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	}
	c.input.ServiceID = serviceID

	// TODO(integralist):
	// Validation such as this should become redundant once Go-Fastly is
	// consistently implementing validation (which itself should be redundant
	// once each backend API is 100% confirmed as validating client inputs).
	//
	// As it stands we have multiple clients duplicating logic which should exist
	// (and thus be relied upon) at the API layer.
	//
	// If neither arguments are provided, error with useful message.
	if !c.name.WasSet && !c.comment.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --name or --comment to update service")
	}

	if c.name.WasSet {
		c.input.Name = fastly.String(c.name.Value)
	}
	if c.comment.WasSet {
		c.input.Comment = fastly.String(c.comment.Value)
	}

	s, err := c.Globals.Client.UpdateService(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
			"Name":       c.name.Value,
			"Comment":    c.comment.Value,
		})
		return err
	}

	text.Success(out, "Updated service %s", s.ID)
	return nil
}
