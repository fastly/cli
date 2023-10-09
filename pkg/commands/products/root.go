package products

import (
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	cmd.JSONOutput

	disableProduct string
	enableProduct  string
	manifest       manifest.Data
	serviceName    cmd.OptionalServiceNameID
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

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.manifest = m
	c.CmdClause = parent.Command("products", "Enable, disable, and check the enablement status of products")

	// Optional.
	c.CmdClause.Flag("disable", "Disable product").HintOptions(ProductEnablementOptions...).EnumVar(&c.disableProduct, ProductEnablementOptions...)
	c.CmdClause.Flag("enable", "Enable product").HintOptions(ProductEnablementOptions...).EnumVar(&c.enableProduct, ProductEnablementOptions...)
	c.RegisterFlagBool(c.JSONFlag()) // --json
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
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	ac := c.Globals.APIClient

	if c.enableProduct != "" && c.disableProduct != "" {
		return fsterr.ErrInvalidProductEnablementFlagCombo
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, _, _, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
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

	var brotliEnabled, diEnabled, fanoutEnabled, ioEnabled, oiEnabled, wsEnabled bool

	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductBrotliCompression,
		ServiceID: serviceID,
	}); err == nil {
		brotliEnabled = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductDomainInspector,
		ServiceID: serviceID,
	}); err == nil {
		diEnabled = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductFanout,
		ServiceID: serviceID,
	}); err == nil {
		fanoutEnabled = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductImageOptimizer,
		ServiceID: serviceID,
	}); err == nil {
		ioEnabled = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductOriginInspector,
		ServiceID: serviceID,
	}); err == nil {
		oiEnabled = true
	}
	if _, err = ac.GetProduct(&fastly.ProductEnablementInput{
		ProductID: fastly.ProductWebSockets,
		ServiceID: serviceID,
	}); err == nil {
		wsEnabled = true
	}

	if ok, err := c.WriteJSON(out, struct {
		BrotliCompression bool `json:"brotli_compression"`
		DomainInspector   bool `json:"domain_inspector"`
		Fanout            bool `json:"fanout"`
		ImageOptimizer    bool `json:"image_optimizer"`
		OriginInspector   bool `json:"origin_inspector"`
		WebSockets        bool `json:"websockets"`
	}{
		BrotliCompression: brotliEnabled,
		DomainInspector:   diEnabled,
		Fanout:            fanoutEnabled,
		ImageOptimizer:    ioEnabled,
		OriginInspector:   oiEnabled,
		WebSockets:        wsEnabled,
	}); ok {
		return err
	}

	t := text.NewTable(out)
	t.AddHeader("PRODUCT", "ENABLED")
	t.AddLine("Brotli Compression", brotliEnabled)
	t.AddLine("Domain Inspector", diEnabled)
	t.AddLine("Fanout", fanoutEnabled)
	t.AddLine("Image Optimizer", ioEnabled)
	t.AddLine("Origin Inspector", oiEnabled)
	t.AddLine("Web Sockets", wsEnabled)
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
