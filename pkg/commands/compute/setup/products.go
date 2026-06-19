package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/fastly/cli/pkg/api"
	fsterrors "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/products/apidiscovery"
	"github.com/fastly/go-fastly/v15/fastly/products/botmanagement"
	"github.com/fastly/go-fastly/v15/fastly/products/brotlicompression"
	"github.com/fastly/go-fastly/v15/fastly/products/ddosprotection"
	"github.com/fastly/go-fastly/v15/fastly/products/domaininspector"
	"github.com/fastly/go-fastly/v15/fastly/products/fanout"
	"github.com/fastly/go-fastly/v15/fastly/products/imageoptimizer"
	"github.com/fastly/go-fastly/v15/fastly/products/logexplorerinsights"
	"github.com/fastly/go-fastly/v15/fastly/products/ngwaf"
	"github.com/fastly/go-fastly/v15/fastly/products/origininspector"
	"github.com/fastly/go-fastly/v15/fastly/products/websockets"
)

// Products represents the service state related to Products defined
// within the fastly.toml [setup] configuration.
//
// NOTE: It implements the setup.Interface interface.
type Products struct {
	// Public
	APIClient      api.Interface
	AcceptDefaults bool
	NonInteractive bool
	Spinner        text.Spinner
	ServiceID      string
	ServiceVersion int
	Setup          *manifest.SetupProducts
	Stdin          io.Reader
	Stdout         io.Writer

	// Private
	required ProductsMap
}

// ProductsMap represents the configuration parameters for enabling specified products
// for a service.
type ProductsMap struct {
	APIDiscovery        ProductSettings
	BotManagement       ProductSettings
	BrotliCompression   ProductSettings
	DDoSProtection      ProductSettings
	DomainInspector     ProductSettings
	Fanout              ProductSettings
	ImageOptimizer      ProductSettings
	LogExplorerInsights ProductSettings
	NGWAF               ProductSettings
	OriginInspector     ProductSettings
	WebSockets          ProductSettings
}

type ProductSettings interface {
	Enabled() bool
}

type Product struct {
	Enable bool
}

func NewProductEnabled() *Product {
	return &Product{Enable: true}
}

var _ ProductSettings = (*Product)(nil)

func (p *Product) Enabled() bool {
	return p != nil && p.Enable
}

type ProductBotManagement struct {
	Product
	ContentGuard string
}

func NewProductBotManagement(contentGuard string) *ProductBotManagement {
	return &ProductBotManagement{
		Product:      *NewProductEnabled(),
		ContentGuard: contentGuard,
	}
}

var _ ProductSettings = (*ProductBotManagement)(nil)

type ProductDDoSProtection struct {
	Product
	Mode string
}

func NewProductDDoSProtection(mode string) *ProductDDoSProtection {
	return &ProductDDoSProtection{
		Product: *NewProductEnabled(),
		Mode:    mode,
	}
}

var _ ProductSettings = (*ProductDDoSProtection)(nil)

type ProductNGWAF struct {
	Product
	WorkspaceID string
}

func NewProductNGWAF(workspaceID string) *ProductNGWAF {
	return &ProductNGWAF{
		Product:     *NewProductEnabled(),
		WorkspaceID: workspaceID,
	}
}

var _ ProductSettings = (*ProductNGWAF)(nil)

type productsSpec struct {
	id                   string
	name                 string
	getSetupProduct      func(*manifest.SetupProducts) manifest.SetupProductSettings
	configure            func(io.Writer, *ProductsMap, manifest.SetupProductSettings) error
	getConfiguredProduct func(*ProductsMap) ProductSettings
	enable               func(*fastly.Client, ProductSettings, string) error
}

var productsSpecs []productsSpec

func init() {
	productsSpecs = []productsSpec{
		{
			id:   apidiscovery.ProductID,
			name: apidiscovery.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.APIDiscovery
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.APIDiscovery = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.APIDiscovery
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := apidiscovery.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   botmanagement.ProductID,
			name: botmanagement.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.BotManagement
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				botManagementSetupProduct, ok := sp.(*manifest.SetupProductBotManagement)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for setupProduct")
				}
				if strings.TrimSpace(botManagementSetupProduct.ContentGuard) == "" {
					return fmt.Errorf("contentguard is required")
				}
				text.Output(w, "  contentguard: %s", botManagementSetupProduct.ContentGuard)
				p.BotManagement = NewProductBotManagement(botManagementSetupProduct.ContentGuard)
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.BotManagement
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				botManagementProduct, ok := product.(*ProductBotManagement)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for product")
				}
				_, err := botmanagement.Enable(context.TODO(), fc, serviceID)
				if err != nil {
					return fmt.Errorf("unexpected: Unable to Enable product")
				}
				_, err = botmanagement.UpdateConfiguration(context.TODO(), fc, serviceID, botmanagement.ConfigureInput{ContentGuard: botManagementProduct.ContentGuard})
				return err
			},
		},
		{
			id:   brotlicompression.ProductID,
			name: brotlicompression.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.BrotliCompression
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.BrotliCompression = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.BrotliCompression
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := brotlicompression.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   ddosprotection.ProductID,
			name: ddosprotection.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.DDoSProtection
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				ddosProtectionSetupProduct, ok := sp.(*manifest.SetupProductDDoSProtection)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for setupProduct")
				}
				if strings.TrimSpace(ddosProtectionSetupProduct.Mode) == "" {
					return fmt.Errorf("mode is required")
				}
				text.Output(w, "  mode: %s", ddosProtectionSetupProduct.Mode)
				p.DDoSProtection = NewProductDDoSProtection(ddosProtectionSetupProduct.Mode)
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.DDoSProtection
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				ddosProtectionProduct, ok := product.(*ProductDDoSProtection)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for product")
				}
				_, err := ddosprotection.Enable(context.TODO(), fc, serviceID, ddosprotection.EnableInput{Mode: ddosProtectionProduct.Mode})
				return err
			},
		},
		{
			id:   domaininspector.ProductID,
			name: domaininspector.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.DomainInspector
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.DomainInspector = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.DomainInspector
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := domaininspector.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   fanout.ProductID,
			name: fanout.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.Fanout
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.Fanout = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.Fanout
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := fanout.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   imageoptimizer.ProductID,
			name: imageoptimizer.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.ImageOptimizer
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.ImageOptimizer = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.ImageOptimizer
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := imageoptimizer.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   logexplorerinsights.ProductID,
			name: logexplorerinsights.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.LogExplorerInsights
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.LogExplorerInsights = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.LogExplorerInsights
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := logexplorerinsights.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   ngwaf.ProductID,
			name: ngwaf.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.NGWAF
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				ngwafSetupProduct, ok := sp.(*manifest.SetupProductNGWAF)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for setupProduct")
				}
				if strings.TrimSpace(ngwafSetupProduct.WorkspaceID) == "" {
					return fmt.Errorf("workspace_id is required")
				}
				text.Output(w, "  workspace_id: %s", ngwafSetupProduct.WorkspaceID)
				p.NGWAF = NewProductNGWAF(ngwafSetupProduct.WorkspaceID)
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.NGWAF
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				ngwafProduct, ok := product.(*ProductNGWAF)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for product")
				}
				_, err := ngwaf.Enable(context.TODO(), fc, serviceID, ngwaf.EnableInput{WorkspaceID: ngwafProduct.WorkspaceID})
				return err
			},
		},
		{
			id:   origininspector.ProductID,
			name: origininspector.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.OriginInspector
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.OriginInspector = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.OriginInspector
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := origininspector.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   websockets.ProductID,
			name: websockets.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.WebSockets
			},
			configure: func(_ io.Writer, p *ProductsMap, _ manifest.SetupProductSettings) error {
				p.WebSockets = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.WebSockets
			},
			enable: func(fc *fastly.Client, _ ProductSettings, serviceID string) error {
				_, err := websockets.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
	}
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (p *Products) Predefined() bool {
	return p != nil && p.Setup != nil && p.Setup.AnyDefined()
}

// Configure prompts the user for specific values related to the service resource.
func (p *Products) Configure() error {
	text.Info(p.Stdout, "The package code will attempt to enable the following products on the service.\n")

	for _, spec := range productsSpecs {
		product := normalizeIfacePtr(spec.getSetupProduct(p.Setup))
		if product == nil || !product.Enabled() {
			continue
		}
		text.Output(p.Stdout, "%s", text.Bold(spec.name))
		if err := spec.configure(p.Stdout, &p.required, product); err != nil {
			return fmt.Errorf("%s: %w", "setup.products."+spec.id, err)
		}
	}

	return nil
}

// Create calls the relevant API to create the service resource(s).
func (p *Products) Create() error {
	anyEnabled := false
	for _, spec := range productsSpecs {
		product := normalizeIfacePtr(spec.getConfiguredProduct(&p.required))
		if product != nil && product.Enabled() {
			anyEnabled = true
			break
		}
	}
	if !anyEnabled {
		return nil
	}

	if p.Spinner == nil {
		return fsterrors.RemediationError{
			Inner:       fmt.Errorf("internal logic error: no spinner configured for setup.Products"),
			Remediation: fsterrors.BugRemediation,
		}
	}

	fc, ok := p.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	for _, spec := range productsSpecs {
		product := normalizeIfacePtr(spec.getConfiguredProduct(&p.required))
		if product == nil || !product.Enabled() {
			continue
		}
		err := p.Spinner.Process(
			fmt.Sprintf("Enabling product '%s'...", spec.id),
			func(_ *text.SpinnerWrapper) error {
				if err := spec.enable(fc, product, p.ServiceID); err != nil {
					return fmt.Errorf("error enabling product [%s]: %w", spec.id, err)
				}
				return nil
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// normalizeIfacePtr converts an interface holding a typed-nil pointer into a real nil interface.
// Works for any interface type parameter I.
func normalizeIfacePtr[I any](v I) I {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		var zero I
		return zero
	}
	return v
}
