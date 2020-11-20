package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// UpdateCommand calls the Fastly API to create services.
type UpdateCommand struct {
	common.Base
	manifest    manifest.Data
	getInput    fastly.GetServiceInput
	updateInput fastly.UpdateServiceInput

	// TODO(integralist):
	// ensure consistency in capitalization
	// should be lowercase to avoid ambiguity in common.Command interface
	//
	name    common.OptionalString
	comment common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a Fastly service")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("name", "Service name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.getInput.ID = serviceID

	// If neither arguments are provided, error with useful message.
	if !c.name.Valid && !c.comment.Valid {
		return fmt.Errorf("error parsing arguments: must provide either --name or --comment to update service")
	}

	s, err := c.Globals.Client.GetService(&c.getInput)
	if err != nil {
		return err
	}

	// set original value, and then afterwards check if we can use the flag value
	c.updateInput.ID = s.ID
	c.updateInput.Name = fastly.String(s.Name)
	c.updateInput.Comment = fastly.String(s.Comment)

	// update field value as required
	if c.name.Valid {
		c.updateInput.Name = fastly.String(c.name.Value)
	}
	if c.comment.Valid {
		c.updateInput.Comment = fastly.String(c.comment.Value)
	}

	s, err = c.Globals.Client.UpdateService(&c.updateInput)
	if err != nil {
		return err
	}

	text.Success(out, "Updated service %s", s.ID)
	return nil
}
