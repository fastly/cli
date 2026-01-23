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
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/products/apidiscovery"
	"github.com/fastly/go-fastly/v12/fastly/products/botmanagement"
	"github.com/fastly/go-fastly/v12/fastly/products/brotlicompression"
	"github.com/fastly/go-fastly/v12/fastly/products/ddosprotection"
	"github.com/fastly/go-fastly/v12/fastly/products/domaininspector"
	"github.com/fastly/go-fastly/v12/fastly/products/fanout"
	"github.com/fastly/go-fastly/v12/fastly/products/imageoptimizer"
	"github.com/fastly/go-fastly/v12/fastly/products/logexplorerinsights"
	"github.com/fastly/go-fastly/v12/fastly/products/ngwaf"
	"github.com/fastly/go-fastly/v12/fastly/products/origininspector"
	"github.com/fastly/go-fastly/v12/fastly/products/websockets"
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
// for a service
type ProductsMap struct {
	APIDiscovery        ProductSettings
	BotManagement       ProductSettings
	BrotliCompression   ProductSettings
	DdosProtection      ProductSettings
	DomainInspector     ProductSettings
	Fanout              ProductSettings
	ImageOptimizer      ProductSettings
	LogExplorerInsights ProductSettings
	Ngwaf               ProductSettings
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

type ProductNgwaf struct {
	Product
	WorkspaceID string
}

func NewProductNgWaf(workspaceID string) *ProductNgwaf {
	return &ProductNgwaf{
		Product:     *NewProductEnabled(),
		WorkspaceID: workspaceID,
	}
}

var _ ProductSettings = (*ProductNgwaf)(nil)

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
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.APIDiscovery = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.APIDiscovery
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
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
				p.BotManagement = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.BotManagement
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				_, err := botmanagement.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   brotlicompression.ProductID,
			name: brotlicompression.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.BrotliCompression
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.BrotliCompression = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.BrotliCompression
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				_, err := brotlicompression.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   ddosprotection.ProductID,
			name: ddosprotection.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.DdosProtection
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.DdosProtection = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.DdosProtection
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				_, err := ddosprotection.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   domaininspector.ProductID,
			name: domaininspector.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.DomainInspector
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.DomainInspector = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.DomainInspector
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
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
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.Fanout = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.Fanout
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
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
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.ImageOptimizer = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.ImageOptimizer
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
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
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.LogExplorerInsights = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.LogExplorerInsights
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				_, err := logexplorerinsights.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:   ngwaf.ProductID,
			name: ngwaf.ProductName,
			getSetupProduct: func(setupProducts *manifest.SetupProducts) manifest.SetupProductSettings {
				return setupProducts.Ngwaf
			},
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				ngwafSetupProduct, ok := sp.(*manifest.SetupProductNgwaf)
				if !ok {
					return fmt.Errorf("unexpected: Incorrect type for setupProduct")
				}
				if strings.TrimSpace(ngwafSetupProduct.WorkspaceID) == "" {
					return fmt.Errorf("workspace_id is required")
				}
				text.Output(w, "  workspace_id: %s", ngwafSetupProduct.WorkspaceID)
				p.Ngwaf = NewProductNgWaf(ngwafSetupProduct.WorkspaceID)
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.Ngwaf
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
				ngwafProduct, ok := product.(*ProductNgwaf)
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
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.OriginInspector = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.OriginInspector
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
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
			configure: func(w io.Writer, p *ProductsMap, sp manifest.SetupProductSettings) error {
				p.WebSockets = NewProductEnabled()
				return nil
			},
			getConfiguredProduct: func(products *ProductsMap) ProductSettings {
				return products.WebSockets
			},
			enable: func(fc *fastly.Client, product ProductSettings, serviceID string) error {
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
