package version

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ValidateCommand calls the Fastly API to validate a service version.
type ValidateCommand struct {
	argparser.Base
	argparser.JSONOutput
	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`

	input          fastly.ValidateVersionInput
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion

	Version string `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ValidateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	return nil
}

// Exec invokes the application logic for the command.
func (c *ValidateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.input.ServiceID = serviceID
	c.input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	valid, msg, err := c.Globals.APIClient.ValidateVersion(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, map[string]any{
		"valid":   valid,
		"message": msg,
	}); ok {
		return err
	}

	if valid {
		if msg != "" {
			text.Success(out, "Service %s version %d is valid: %s", serviceID, c.input.ServiceVersion, msg)
		} else {
			text.Success(out, "Service %s version %d is valid", serviceID, c.input.ServiceVersion)
		}
	} else {
		if msg != "" {
			text.Error(out, "Service %s version %d is not valid: %s", serviceID, c.input.ServiceVersion, msg)
		} else {
			text.Error(out, "Service %s version %d is not valid", serviceID, c.input.ServiceVersion)
		}
	}

	return nil
}
