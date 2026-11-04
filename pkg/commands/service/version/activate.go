package version

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ActivateCommand calls the Fastly API to activate a service version.
type ActivateCommand struct {
	argparser.Base
	Input          fastly.ActivateVersionInput     `kong:"-"`
	ServiceID      *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName    argparser.OptionalServiceNameID `help:"The name of the service"`
	Version        string                          `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`
	serviceVersion argparser.OptionalServiceVersion
	AutoClone      argparser.OptionalAutoClone `name:"autoclone" help:"If the selected service version is not editable, clone it and use the clone."`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ActivateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	return nil
}

// Exec invokes the application logic for the command.
func (c *ActivateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
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
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	ver, err := c.Globals.APIClient.ActivateVersion(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	text.Success(out, "Activated service %s version %d", fastly.ToValue(ver.ServiceID), c.Input.ServiceVersion)
	return nil
}
