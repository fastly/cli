package domain

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update domains.
type UpdateCommand struct {
	cmd.Base
	manifest       manifest.Data
	input          fastly.UpdateDomainInput
	serviceVersion cmd.OptionalServiceVersion
	autoClone      cmd.OptionalAutoClone

	NewName cmd.OptionalString
	Comment cmd.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("update", "Update a domain on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.NewServiceVersionFlag(cmd.ServiceVersionFlagOpts{Dst: &c.serviceVersion.Value})
	c.NewAutoCloneFlag(c.autoClone.Set, &c.autoClone.Value)
	c.CmdClause.Flag("name", "Domain name").Short('n').Required().StringVar(&c.input.Name)
	c.CmdClause.Flag("new-name", "New domain name").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.Comment.Set).StringVar(&c.Comment.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.input.ServiceID = serviceID

	v, err := c.serviceVersion.Parse(c.input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	v, err = c.autoClone.Parse(v, c.input.ServiceID, c.Globals.Client)
	if err != nil {
		return err
	}
	c.input.ServiceVersion = v.Number

	// If neither arguments are provided, error with useful message.
	if !c.NewName.WasSet && !c.Comment.WasSet {
		return fmt.Errorf("error parsing arguments: must provide either --new-name or --comment to update domain")
	}

	if c.NewName.WasSet {
		c.input.NewName = fastly.String(c.NewName.Value)
	}
	if c.Comment.WasSet {
		c.input.Comment = fastly.String(c.Comment.Value)
	}

	d, err := c.Globals.Client.UpdateDomain(&c.input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated domain %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
