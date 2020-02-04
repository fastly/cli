package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to create services.
type UpdateCommand struct {
	common.Base
	manifest    manifest.Data
	getInput    fastly.GetServiceInput
	updateInput fastly.UpdateServiceInput
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a Fastly service")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("name", "Service name").Short('n').StringVar(&c.updateInput.Name)
	c.CmdClause.Flag("comment", "Human-readable comment").StringVar(&c.updateInput.Comment)
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
	if c.updateInput.Name == "" && c.updateInput.Comment == "" {
		return fmt.Errorf("error parsing arguments: must provide either --name or --comment to update service")
	}

	s, err := c.Globals.Client.GetService(&c.getInput)
	if err != nil {
		return err
	}

	c.updateInput.ID = s.ID

	// Only update name if non-empty.
	if c.updateInput.Name == "" {
		c.updateInput.Name = s.Name
	}

	// Only update comment if non-empty.
	if c.updateInput.Comment == "" {
		c.updateInput.Comment = s.Comment
	}

	s, err = c.Globals.Client.UpdateService(&c.updateInput)
	if err != nil {
		return err
	}

	text.Success(out, "Updated service %s", s.ID)
	return nil
}
