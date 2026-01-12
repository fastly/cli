package imageoptimizerdefaults

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/flagconversion"
	"github.com/fastly/cli/pkg/global"
)

// UpdateCommand calls the Fastly API to update Image Optimizer default settings for a service.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.UpdateImageOptimizerDefaultSettingsInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion

	// Image Optimizer setting flags
	allowVideo   argparser.OptionalString
	jpegQuality  argparser.OptionalInt
	jpegType     argparser.OptionalString
	resizeFilter argparser.OptionalString
	upscale      argparser.OptionalString
	webp         argparser.OptionalString
	webpQuality  argparser.OptionalInt
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update Image Optimizer default settings for a service")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
		Required:    true,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional flags for Image Optimizer settings
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "allow-video",
		Description: "Enables GIF to MP4 transformations on this service [true, false]",
		Action:      c.allowVideo.Set,
		Dst:         &c.allowVideo.Value,
	})
	c.RegisterFlagInt(argparser.IntFlagOpts{
		Name:        "jpeg-quality",
		Description: "The default quality to use with JPEG output (1-100)",
		Dst:         &c.jpegQuality.Value,
		Action:      c.jpegQuality.Set,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "jpeg-type",
		Description: "The default type of JPEG output to use (auto, baseline, progressive)",
		Dst:         &c.jpegType.Value,
		Action:      c.jpegType.Set,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "resize-filter",
		Description: "The type of filter to use while resizing an image (lanczos3, lanczos2, bicubic, bilinear, nearest)",
		Dst:         &c.resizeFilter.Value,
		Action:      c.resizeFilter.Set,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "upscale",
		Description: "Whether or not we should allow output images to render at sizes larger than input [true, false]",
		Action:      c.upscale.Set,
		Dst:         &c.upscale.Value,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "webp",
		Description: "Controls whether or not to default to WebP output when the client supports it [true, false]",
		Action:      c.webp.Set,
		Dst:         &c.webp.Value,
	})
	c.RegisterFlagInt(argparser.IntFlagOpts{
		Name:        "webp-quality",
		Description: "The default quality to use with WebP output (1-100)",
		Dst:         &c.webpQuality.Value,
		Action:      c.webpQuality.Set,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
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
		ServiceNameFlag:    c.serviceName,
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
	if c.resizeFilter.WasSet {
		// Convert string to ImageOptimizerResizeFilter constant
		switch c.resizeFilter.Value {
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
			return fmt.Errorf("invalid resize filter: %s. Valid options: lanczos3, lanczos2, bicubic, bilinear, nearest", c.resizeFilter.Value)
		}
	}
	if c.webp.WasSet {
		webp, err := flagconversion.ConvertBoolFromStringFlag(c.webp.Value)
		if err != nil {
			err := errors.New("'webp' flag must be one of the following [true, false]")
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.Input.Webp = webp
	}
	if c.webpQuality.WasSet {
		c.Input.WebpQuality = &c.webpQuality.Value
	}
	if c.jpegType.WasSet {
		// Convert string to JPEG type constant
		switch c.jpegType.Value {
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
			return fmt.Errorf("invalid jpeg type: %s. Valid options: auto, baseline, progressive", c.jpegType.Value)
		}
	}
	if c.jpegQuality.WasSet {
		c.Input.JpegQuality = &c.jpegQuality.Value
	}
	if c.upscale.WasSet {
		upscale, err := flagconversion.ConvertBoolFromStringFlag(c.upscale.Value)
		if err != nil {
			err := errors.New("'upscale' flag must be one of the following [true, false]")
			c.Globals.ErrLog.Add(err)
			return err
		}
		c.Input.Upscale = upscale
	}
	if c.allowVideo.WasSet {
		allowVideo, err := flagconversion.ConvertBoolFromStringFlag(c.allowVideo.Value)
		if err != nil {
			err := errors.New("'allow-video' flag must be one of the following [true, false]")
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
