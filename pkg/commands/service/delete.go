package service

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// DeleteCommand calls the Fastly API to delete services.
type DeleteCommand struct {
	cmd.Base
	manifest    manifest.Data
	Input       fastly.DeleteServiceInput
	force       bool
	serviceName cmd.OptionalServiceNameID
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("delete", "Delete a Fastly service").Alias("remove")
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("force", "Force deletion of an active service").Short('f').BoolVar(&c.force)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ID = serviceID

	if c.force {
		s, err := c.Globals.APIClient.GetServiceDetails(&fastly.GetServiceInput{
			ID: serviceID,
		})
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
			return err
		}

		if s.ActiveVersion.Number != 0 {
			_, err := c.Globals.APIClient.DeactivateVersion(&fastly.DeactivateVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: s.ActiveVersion.Number,
			})
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]any{
					"Service ID":      serviceID,
					"Service Version": s.ActiveVersion.Number,
				})
				return err
			}
		}
	}

	if err := c.Globals.APIClient.DeleteService(&c.Input); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
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
