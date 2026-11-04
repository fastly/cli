package imageoptimizerdefaults

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// GetCommand calls the Fastly API to describe the Image Optimizer default settings for a service.
type GetCommand struct {
	argparser.Base

	// Required.
	ServiceID *string `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	Version   string  `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional.
	argparser.JSONOutput

	Input          fastly.GetImageOptimizerDefaultSettingsInput `kong:"-"`
	ServiceName    argparser.OptionalServiceNameID              `help:"The name of the service"`
	serviceVersion argparser.OptionalServiceVersion
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *GetCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	return nil
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
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

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.GetImageOptimizerDefaultSettings(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	fmt.Fprintf(out, "Allow Video: %t\n", o.AllowVideo)
	fmt.Fprintf(out, "JPEG Quality: %d\n", o.JpegQuality)
	fmt.Fprintf(out, "JPEG Type: %s\n", o.JpegType)
	fmt.Fprintf(out, "Resize Filter: %s\n", o.ResizeFilter)
	fmt.Fprintf(out, "Upscale: %t\n", o.Upscale)
	fmt.Fprintf(out, "WebP: %t\n", o.Webp)
	fmt.Fprintf(out, "WebP Quality: %d\n", o.WebpQuality)

	return nil
}
