package domain

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

// UpdateCommand calls the Fastly API to update domains.
type UpdateCommand struct {
	common.Base
	manifest    manifest.Data
	getInput    fastly.GetDomainInput
	updateInput fastly.UpdateDomainInput
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.CmdClause = parent.Command("update", "Update a domain on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.getInput.ServiceVersion)
	c.CmdClause.Flag("name", "Domain name").Short('n').Required().StringVar(&c.getInput.Name)
	c.CmdClause.Flag("new-name", "New domain name").StringVar(&c.updateInput.NewName)
	c.CmdClause.Flag("comment", "A descriptive note").StringVar(c.updateInput.Comment)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.getInput.ServiceID = serviceID

	// If neither arguments are provided, error with useful message.
	if c.updateInput.NewName == "" && *c.updateInput.Comment == "" {
		return fmt.Errorf("error parsing arguments: must provide either --new-name or --comment to update domain")
	}

	d, err := c.Globals.Client.GetDomain(&c.getInput)
	if err != nil {
		return err
	}

	c.updateInput.ServiceID = d.ServiceID
	c.updateInput.ServiceVersion = d.ServiceVersion

	// Only update name if non-empty.
	if c.updateInput.NewName == "" {
		c.updateInput.NewName = d.Name
	}

	// Only update comment if non-empty.
	if *c.updateInput.Comment == "" {
		c.updateInput.Comment = fastly.String(d.Comment)
	}

	d, err = c.Globals.Client.UpdateDomain(&c.updateInput)
	if err != nil {
		return err
	}

	text.Success(out, "Updated domain %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
