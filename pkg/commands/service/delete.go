package service

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete services.
type DeleteCommand struct {
	argparser.Base
	Input fastly.DeleteServiceInput `kong:"-"`

	// Optional.
	Force       bool                            `short:"f" help:"Force deletion of an active service"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DeleteCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID

	if c.Force {
		s, err := c.Globals.APIClient.GetServiceDetails(context.TODO(), &fastly.GetServiceDetailsInput{
			ServiceID: serviceID,
		})
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
			return err
		}

		if s.ActiveVersion != nil && fastly.ToValue(s.ActiveVersion.Number) != 0 {
			_, err := c.Globals.APIClient.DeactivateVersion(context.TODO(), &fastly.DeactivateVersionInput{
				ServiceID:      serviceID,
				ServiceVersion: fastly.ToValue(s.ActiveVersion.Number),
			})
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]any{
					"Service ID":      serviceID,
					"Service Version": fastly.ToValue(s.ActiveVersion.Number),
				})
				return err
			}
		}
	}

	if err := c.Globals.APIClient.DeleteService(context.TODO(), &c.Input); err != nil {
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

	text.Success(out, "Deleted service ID %s", c.Input.ServiceID)
	return nil
}
