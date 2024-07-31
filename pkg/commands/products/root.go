package products

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
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
	"brotli_compression",
	"domain_inspector",
	"fanout",
	"image_optimizer",
	"origin_inspector",
	"websockets",
}

// CommandName is the string to be used to invoke this command
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
	ac := c.Globals.APIClient

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

	if c.enableProduct != "" {
		p := identifyProduct(c.enableProduct)
		if p == fastly.ProductUndefined {
			return errors.New("unrecognised product")
		}
		if _, err := ac.EnableProduct(&fastly.ProductEnablementInput{
			ProductID: p,
			ServiceID: serviceID,
		}); err != nil {
			return fmt.Errorf("failed to enable product '%s': %w", c.enableProduct, err)
		}
		text.Success(out, "Successfully enabled product '%s'", c.enableProduct)
		return nil
	}

	if c.disableProduct != "" {
		p := identifyProduct(c.disableProduct)
		if p == fastly.ProductUndefined {
			return errors.New("unrecognised product")
		}
		if err := ac.DisableProduct(&fastly.ProductEnablementInput{
			ProductID: p,
			ServiceID: serviceID,
		}); err != nil {
			return fmt.Errorf("failed to disable product '%s': %w", c.disableProduct, err)
		}
		text.Success(out, "Successfully disabled product '%s'", c.disableProduct)
		return nil
	}

	ps := ProductStatus{}

	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductBrotliCompression,
		ServiceID: serviceID,
	}); err == nil {
		ps.BrotliCompression = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductDomainInspector,
		ServiceID: serviceID,
	}); err == nil {
		ps.DomainInspector = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductFanout,
		ServiceID: serviceID,
	}); err == nil {
		ps.Fanout = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductImageOptimizer,
		ServiceID: serviceID,
	}); err == nil {
		ps.ImageOptimizer = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductOriginInspector,
		ServiceID: serviceID,
	}); err == nil {
		ps.OriginInspector = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductWebSockets,
		ServiceID: serviceID,
	}); err == nil {
		ps.WebSockets = true
	}

	if ok, err := c.WriteJSON(out, ps); ok {
		return err
	}

	t := text.NewTable(out)
	t.AddHeader("PRODUCT", "ENABLED")
	t.AddLine("Brotli Compression", ps.BrotliCompression)
	t.AddLine("Domain Inspector", ps.DomainInspector)
	t.AddLine("Fanout", ps.Fanout)
	t.AddLine("Image Optimizer", ps.ImageOptimizer)
	t.AddLine("Origin Inspector", ps.OriginInspector)
	t.AddLine("Web Sockets", ps.WebSockets)
	t.Print()
	return nil
}

func identifyProduct(product string) fastly.Product {
	switch product {
	case "brotli_compression":
		return fastly.ProductBrotliCompression
	case "domain_inspector":
		return fastly.ProductDomainInspector
	case "fanout":
		return fastly.ProductFanout
	case "image_optimizer":
		return fastly.ProductImageOptimizer
	case "origin_inspector":
		return fastly.ProductOriginInspector
	case "websockets":
		return fastly.ProductWebSockets
	default:
		return fastly.ProductUndefined
	}
}

// ProductStatus indicates the status for each product.
type ProductStatus struct {
	BrotliCompression bool `json:"brotli_compression"`
	DomainInspector   bool `json:"domain_inspector"`
	Fanout            bool `json:"fanout"`
	ImageOptimizer    bool `json:"image_optimizer"`
	OriginInspector   bool `json:"origin_inspector"`
	WebSockets        bool `json:"websockets"`
}
