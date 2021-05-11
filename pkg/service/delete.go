package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DeleteCommand calls the Fastly API to delete services.
type DeleteCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.DeleteServiceInput
	force    bool
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent common.Registerer, globals *config.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("delete", "Delete a Fastly service").Alias("remove")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
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
			return err
		}

		if s.ActiveVersion.Number != 0 {
			_, err := c.Globals.Client.DeactivateVersion(&fastly.DeactivateVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: s.ActiveVersion.Number,
			})
			if err != nil {
				return err
			}
		}
	}

	if err := c.Globals.Client.DeleteService(&c.Input); err != nil {
		return errors.RemediationError{
			Inner:       err,
			Remediation: fmt.Sprintf("Try %s\n", text.Bold("fastly service delete --force")),
		}
	}

	text.Success(out, "Deleted service ID %s", c.Input.ID)
	return nil
}
