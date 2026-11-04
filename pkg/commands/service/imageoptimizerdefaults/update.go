package imageoptimizerdefaults

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// UpdateCommand calls the Fastly API to update Image Optimizer default settings for a service.
type UpdateCommand struct {
	argparser.Base

	Input          fastly.UpdateImageOptimizerDefaultSettingsInput `kong:"-"`
	serviceVersion argparser.OptionalServiceVersion

	// Required.
	ServiceID *string `short:"s" required:"" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	Version   string  `required:"" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version"`

	// Optional flags for Image Optimizer settings

	// Image Optimizer setting flags
	AllowVideo   argparser.OptionalString `help:"Enables GIF to MP4 transformations on this service [true, false]"`
	JpegQuality  argparser.OptionalInt    `help:"The default quality to use with JPEG output (1-100)"`
	JpegType     argparser.OptionalString `help:"The default type of JPEG output to use (auto, baseline, progressive)"`
	ResizeFilter argparser.OptionalString `help:"The type of filter to use while resizing an image (lanczos3, lanczos2, bicubic, bilinear, nearest)"`
	Upscale      argparser.OptionalString `help:"Whether or not we should allow output images to render at sizes larger than input [true, false]"`
	Webp         argparser.OptionalString `help:"Controls whether or not to default to WebP output when the client supports it [true, false]"`
	WebpQuality  argparser.OptionalInt    `help:"The default quality to use with WebP output (1-100)"`

	// Optional.
	argparser.JSONOutput

	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	c.serviceVersion.Value = c.Version
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	// Set optional fields only if they were provided
	if c.ResizeFilter.WasSet {
		// Convert string to ImageOptimizerResizeFilter constant
		switch c.ResizeFilter.Value {
		case "lanczos3":
			filter := fastly.ImageOptimizerLanczos3
			c.Input.ResizeFilter = &filter
		case "lanczos2":
			filter := fastly.ImageOptimizerLanczos2
			c.Input.ResizeFilter = &filter
		case "bicubic":
			filter := fastly.ImageOptimizerBicubic
			c.Input.ResizeFilter = &filter
		case "bilinear":
			filter := fastly.ImageOptimizerBilinear
			c.Input.ResizeFilter = &filter
		case "nearest":
			filter := fastly.ImageOptimizerNearest
			c.Input.ResizeFilter = &filter
		default:
			return fmt.Errorf("invalid resize filter: %s. Valid options: lanczos3, lanczos2, bicubic, bilinear, nearest", c.ResizeFilter.Value)
		}
	}
	if c.Webp.WasSet {
		webp, err := argparser.ConvertBoolFromStringFlag(c.Webp.Value, "webp")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.Input.Webp = webp
	}
	if c.WebpQuality.WasSet {
		c.Input.WebpQuality = &c.WebpQuality.Value
	}
	if c.JpegType.WasSet {
		// Convert string to JPEG type constant
		switch c.JpegType.Value {
		case "auto":
			jpegType := fastly.ImageOptimizerAuto
			c.Input.JpegType = &jpegType
		case "baseline":
			jpegType := fastly.ImageOptimizerBaseline
			c.Input.JpegType = &jpegType
		case "progressive":
			jpegType := fastly.ImageOptimizerProgressive
			c.Input.JpegType = &jpegType
		default:
			return fmt.Errorf("invalid jpeg type: %s. Valid options: auto, baseline, progressive", c.JpegType.Value)
		}
	}
	if c.JpegQuality.WasSet {
		c.Input.JpegQuality = &c.JpegQuality.Value
	}
	if c.Upscale.WasSet {
		upscale, err := argparser.ConvertBoolFromStringFlag(c.Upscale.Value, "upscale")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.Input.Upscale = upscale
	}
	if c.AllowVideo.WasSet {
		allowVideo, err := argparser.ConvertBoolFromStringFlag(c.AllowVideo.Value, "allow-video")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.Input.AllowVideo = allowVideo
	}

	o, err := c.Globals.APIClient.UpdateImageOptimizerDefaultSettings(context.TODO(), &c.Input)
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

	fmt.Fprintf(out, "Updated Image Optimizer default settings for service %s (version %d)\n", serviceID, fastly.ToValue(serviceVersion.Number))
	fmt.Fprintf(out, "\nAllow Video: %t\n", o.AllowVideo)
	fmt.Fprintf(out, "JPEG Quality: %d\n", o.JpegQuality)
	fmt.Fprintf(out, "JPEG Type: %s\n", o.JpegType)
	fmt.Fprintf(out, "Resize Filter: %s\n", o.ResizeFilter)
	fmt.Fprintf(out, "Upscale: %t\n", o.Upscale)
	fmt.Fprintf(out, "WebP: %t\n", o.Webp)
	fmt.Fprintf(out, "WebP Quality: %d\n", o.WebpQuality)

	return nil
}
