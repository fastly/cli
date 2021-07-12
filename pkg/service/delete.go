package service

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

// DeleteCommand calls the Fastly API to delete services.
type DeleteCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.DeleteServiceInput
	force    bool
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("delete", "Delete a Fastly service").Alias("remove")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("force", "Force deletion of an active service").Short('f').BoolVar(&c.force)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ID = serviceID

	if c.force {
		s, err := c.Globals.Client.GetServiceDetails(&fastly.GetServiceInput{
			ID: serviceID,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if s.ActiveVersion.Number != 0 {
			_, err := c.Globals.Client.DeactivateVersion(&fastly.DeactivateVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: s.ActiveVersion.Number,
			})
			if err != nil {
				c.Globals.ErrLog.Add(err)
				return err
			}
		}
	}

	if err := c.Globals.Client.DeleteService(&c.Input); err != nil {
		c.Globals.ErrLog.Add(err)
		return errors.RemediationError{
			Inner:       err,
			Remediation: fmt.Sprintf("Try %s\n", text.Bold("fastly service delete --force")),
		}
	}

	// Ensure that VCL service users are unaffected by checking if the Service ID
	// was acquired via the fastly.toml manifest.
	if source == manifest.SourceFile {
		if err := c.manifest.File.Read(manifest.Filename); err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error reading package manifest: %w", err)
		}
		c.manifest.File.ServiceID = ""
		if err := c.manifest.File.Write(manifest.Filename); err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error updating package manifest: %w", err)
		}
	}

	text.Success(out, "Deleted service ID %s", c.Input.ID)
	return nil
}
