package products

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly/products/botmanagement"
	"github.com/fastly/go-fastly/v12/fastly/products/brotlicompression"
	"github.com/fastly/go-fastly/v12/fastly/products/domaininspector"
	"github.com/fastly/go-fastly/v12/fastly/products/fanout"
	"github.com/fastly/go-fastly/v12/fastly/products/imageoptimizer"
	"github.com/fastly/go-fastly/v12/fastly/products/logexplorerinsights"
	"github.com/fastly/go-fastly/v12/fastly/products/origininspector"
	"github.com/fastly/go-fastly/v12/fastly/products/websockets"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	argparser.JSONOutput

	disableProduct string
	enableProduct  string
	serviceName    argparser.OptionalServiceNameID
}

// ProductEnablementOptions is a list of products that can be enabled/disabled.
var ProductEnablementOptions = []string{
	"bot_management",
	"brotli_compression",
	"domain_inspector",
	"fanout",
	"image_optimizer",
	"log_explorer_insights",
	"origin_inspector",
	"websockets",
}

// ProductStatus indicates the status for each product.
type ProductStatus struct {
	BotManagement       bool `json:"bot_management"`
	BrotliCompression   bool `json:"brotli_compression"`
	DomainInspector     bool `json:"domain_inspector"`
	Fanout              bool `json:"fanout"`
	ImageOptimizer      bool `json:"image_optimizer"`
	LogExplorerInsights bool `json:"log_explorer_insights"`
	OriginInspector     bool `json:"origin_inspector"`
	WebSockets          bool `json:"websockets"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "products"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "Enable, disable, and check the enablement status of products")

	// Optional.
	c.CmdClause.Flag("disable", "Disable product").HintOptions(ProductEnablementOptions...).EnumVar(&c.disableProduct, ProductEnablementOptions...)
	c.CmdClause.Flag("enable", "Enable product").HintOptions(ProductEnablementOptions...).EnumVar(&c.enableProduct, ProductEnablementOptions...)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.enableProduct != "" && c.disableProduct != "" {
		return fsterr.ErrInvalidEnableDisableFlagCombo
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, _, _, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return fmt.Errorf("failed to identify Service ID: %w", err)
	}

	ac, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	if c.enableProduct != "" {
		switch c.enableProduct {
		case "bot_management":
			_, err = botmanagement.Enable(context.TODO(), ac, serviceID)
		case "brotli_compression":
			_, err = brotlicompression.Enable(context.TODO(), ac, serviceID)
		case "domain_inspector":
			_, err = domaininspector.Enable(context.TODO(), ac, serviceID)
		case "fanout":
			_, err = fanout.Enable(context.TODO(), ac, serviceID)
		case "image_optimizer":
			_, err = imageoptimizer.Enable(context.TODO(), ac, serviceID)
		case "log_explorer_insights":
			_, err = logexplorerinsights.Enable(context.TODO(), ac, serviceID)
		case "origin_inspector":
			_, err = origininspector.Enable(context.TODO(), ac, serviceID)
		case "websockets":
			_, err = websockets.Enable(context.TODO(), ac, serviceID)
		default:
			return errors.New("unrecognised product")
		}
		if err != nil {
			return fmt.Errorf("failed to enable product '%s': %w", c.enableProduct, err)
		}
		text.Success(out, "Successfully enabled product '%s'", c.enableProduct)
		return nil
	}

	if c.disableProduct != "" {
		switch c.disableProduct {
		case "bot_management":
			err = botmanagement.Disable(context.TODO(), ac, serviceID)
		case "brotli_compression":
			err = brotlicompression.Disable(context.TODO(), ac, serviceID)
		case "domain_inspector":
			err = domaininspector.Disable(context.TODO(), ac, serviceID)
		case "fanout":
			err = fanout.Disable(context.TODO(), ac, serviceID)
		case "image_optimizer":
			err = imageoptimizer.Disable(context.TODO(), ac, serviceID)
		case "log_explorer_insights":
			err = logexplorerinsights.Disable(context.TODO(), ac, serviceID)
		case "origin_inspector":
			err = origininspector.Disable(context.TODO(), ac, serviceID)
		case "websockets":
			err = websockets.Disable(context.TODO(), ac, serviceID)
		default:
			return errors.New("unrecognised product")
		}
		if err != nil {
			return fmt.Errorf("failed to disable product '%s': %w", c.disableProduct, err)
		}
		text.Success(out, "Successfully disabled product '%s'", c.disableProduct)
		return nil
	}

	ps := ProductStatus{}

	if _, err = botmanagement.Get(context.TODO(), ac, serviceID); err == nil {
		ps.BotManagement = true
	}
	if _, err = brotlicompression.Get(context.TODO(), ac, serviceID); err == nil {
		ps.BrotliCompression = true
	}
	if _, err = domaininspector.Get(context.TODO(), ac, serviceID); err == nil {
		ps.DomainInspector = true
	}
	if _, err = fanout.Get(context.TODO(), ac, serviceID); err == nil {
		ps.Fanout = true
	}
	if _, err = imageoptimizer.Get(context.TODO(), ac, serviceID); err == nil {
		ps.ImageOptimizer = true
	}
	if _, err = logexplorerinsights.Get(context.TODO(), ac, serviceID); err == nil {
		ps.LogExplorerInsights = true
	}
	if _, err = origininspector.Get(context.TODO(), ac, serviceID); err == nil {
		ps.OriginInspector = true
	}
	if _, err = websockets.Get(context.TODO(), ac, serviceID); err == nil {
		ps.WebSockets = true
	}

	if ok, err := c.WriteJSON(out, ps); ok {
		return err
	}

	t := text.NewTable(out)
	t.AddHeader("PRODUCT", "ENABLED")
	t.AddLine("Bot Management", ps.BotManagement)
	t.AddLine("Brotli Compression", ps.BrotliCompression)
	t.AddLine("Domain Inspector", ps.DomainInspector)
	t.AddLine("Fanout", ps.Fanout)
	t.AddLine("Image Optimizer", ps.ImageOptimizer)
	t.AddLine("Log Explorer & Insights", ps.LogExplorerInsights)
	t.AddLine("Origin Inspector", ps.OriginInspector)
	t.AddLine("WebSockets", ps.WebSockets)
	t.Print()
	return nil
}
