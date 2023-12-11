package service

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete services.
type DeleteCommand struct {
	argparser.Base
	Input       fastly.DeleteServiceInput
	force       bool
	serviceName argparser.OptionalServiceNameID
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a Fastly service").Alias("remove")

	// Optional.
	c.CmdClause.Flag("force", "Force deletion of an active service").Short('f').BoolVar(&c.force)
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
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
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
		if err := c.Globals.Manifest.File.Read(manifest.Filename); err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error reading fastly.toml: %w", err)
		}
		c.Globals.Manifest.File.ServiceID = ""
		if err := c.Globals.Manifest.File.Write(manifest.Filename); err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error updating fastly.toml: %w", err)
		}
	}

	text.Success(out, "Deleted service ID %s", c.Input.ID)
	return nil
}
