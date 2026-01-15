package setup

import (
	"context"
	"errors"
	"fmt"
	"io"

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
	required Product
}

// Product represents the configuration parameters for creating a KV Store via
// the API client.
type Product struct {
	APIDiscovery        *ProductSettingsEnable
	BotManagement       *ProductSettingsEnable
	BrotliCompression   *ProductSettingsEnable
	DdosProtection      *ProductSettingsEnable
	DomainInspector     *ProductSettingsEnable
	Fanout              *ProductSettingsEnable
	ImageOptimizer      *ProductSettingsEnable
	LogExplorerInsights *ProductSettingsEnable
	Ngwaf               *ProductSettingsNgwaf
	OriginInspector     *ProductSettingsEnable
	WebSockets          *ProductSettingsEnable
}

type ProductSettings interface {
	Enabled() bool
}

type ProductSettingsEnable struct {
	Enable bool
}

var _ ProductSettings = (*ProductSettingsEnable)(nil)

func (p *ProductSettingsEnable) Enabled() bool {
	return p != nil && p.Enable
}

type ProductSettingsNgwaf struct {
	ProductSettingsEnable
	WorkspaceID string
}

var _ ProductSettings = (*ProductSettingsNgwaf)(nil)

func (p *ProductSettingsNgwaf) Enabled() bool {
	if p == nil {
		return false
	}
	return p.ProductSettingsEnable.Enabled()
}

// Predefined indicates if the service resource has been specified within the
// fastly.toml file using a [setup] configuration block.
func (p *Products) Predefined() bool {
	return p != nil && p.Setup != nil && p.Setup.AnyEnabled()
}

// Configure prompts the user for specific values related to the service resource.
func (p *Products) Configure() error {
	if !p.Predefined() {
		return nil
	}

	text.Info(p.Stdout, "The package code will attempt to enable the following products on the service.\n\n")

	type productSpec struct {
		run func() error
	}

	specs := []productSpec{
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.APIDiscovery,
					&p.required.APIDiscovery,
					apidiscovery.ProductName,
					"setup.products."+apidiscovery.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.BotManagement,
					&p.required.BotManagement,
					botmanagement.ProductName,
					"setup.products."+botmanagement.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.BrotliCompression,
					&p.required.BrotliCompression,
					brotlicompression.ProductName,
					"setup.products."+brotlicompression.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.DdosProtection,
					&p.required.DdosProtection,
					ddosprotection.ProductName,
					"setup.products."+ddosprotection.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.DomainInspector,
					&p.required.DomainInspector,
					domaininspector.ProductName,
					"setup.products."+domaininspector.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.Fanout,
					&p.required.Fanout,
					fanout.ProductName,
					"setup.products."+fanout.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.ImageOptimizer,
					&p.required.ImageOptimizer,
					imageoptimizer.ProductName,
					"setup.products."+imageoptimizer.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.LogExplorerInsights,
					&p.required.LogExplorerInsights,
					logexplorerinsights.ProductName,
					"setup.products."+logexplorerinsights.ProductID,
				)
			},
		},
		{
			run: func() error {
				return whenEnabledDo(
					p.Stdout,
					p.Setup.Ngwaf,
					func(productSettingsEnable ProductSettingsEnable) error {
						text.Output(p.Stdout, "  %s", p.Setup.Ngwaf.WorkspaceID)
						p.required.Ngwaf = &ProductSettingsNgwaf{
							ProductSettingsEnable: productSettingsEnable,
							WorkspaceID:           p.Setup.Ngwaf.WorkspaceID,
						}
						return nil
					},
					ngwaf.ProductName,
					"setup.products."+ngwaf.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.OriginInspector,
					&p.required.OriginInspector,
					origininspector.ProductName,
					"setup.products."+origininspector.ProductID,
				)
			},
		},
		{
			run: func() error {
				return configureIfEnabled(
					p.Stdout,
					p.Setup.WebSockets,
					&p.required.WebSockets,
					websockets.ProductName,
					"setup.products."+websockets.ProductID,
				)
			},
		},
	}

	for _, spec := range specs {
		if err := spec.run(); err != nil {
			return err
		}
	}

	return nil
}

func whenEnabledDo(w io.Writer, s manifest.SetupProduct, fn func(productSettingsEnable ProductSettingsEnable) error, label string, path string) error {
	if s == nil || !s.Enabled() {
		return nil
	}
	if err := s.Validate(); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	text.Output(w, "%s", text.Bold(label))
	return fn(ProductSettingsEnable{Enable: true})
}

func configureIfEnabled(w io.Writer, setupProduct manifest.SetupProduct, product **ProductSettingsEnable, label string, path string) error {
	return whenEnabledDo(w, setupProduct, func(productSettingsEnable ProductSettingsEnable) error {
		*product = &productSettingsEnable
		return nil
	}, label, path)
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

	type enableSpec struct {
		id      string
		enabled func() bool
		enable  func(fc *fastly.Client, serviceID string) error
	}

	specs := []enableSpec{
		{
			id:      apidiscovery.ProductID,
			enabled: func() bool { return p.required.APIDiscovery.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := apidiscovery.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      botmanagement.ProductID,
			enabled: func() bool { return p.required.BotManagement.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := botmanagement.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      brotlicompression.ProductID,
			enabled: func() bool { return p.required.BrotliCompression.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := brotlicompression.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      ddosprotection.ProductID,
			enabled: func() bool { return p.required.DdosProtection.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := ddosprotection.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      domaininspector.ProductID,
			enabled: func() bool { return p.required.DomainInspector.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := domaininspector.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      fanout.ProductID,
			enabled: func() bool { return p.required.Fanout.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := fanout.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      imageoptimizer.ProductID,
			enabled: func() bool { return p.required.ImageOptimizer.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := imageoptimizer.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      logexplorerinsights.ProductID,
			enabled: func() bool { return p.required.LogExplorerInsights.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := logexplorerinsights.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      ngwaf.ProductID,
			enabled: func() bool { return p.required.Ngwaf.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := ngwaf.Enable(context.TODO(), fc, serviceID, ngwaf.EnableInput{WorkspaceID: p.required.Ngwaf.WorkspaceID})
				return err
			},
		},
		{
			id:      origininspector.ProductID,
			enabled: func() bool { return p.required.OriginInspector.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := origininspector.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
		{
			id:      websockets.ProductID,
			enabled: func() bool { return p.required.WebSockets.Enabled() },
			enable: func(fc *fastly.Client, serviceID string) error {
				_, err := websockets.Enable(context.TODO(), fc, serviceID)
				return err
			},
		},
	}

	for _, s := range specs {
		if err := p.enableRequiredProduct(fc, s.id, s.enabled(), s.enable); err != nil {
			return err
		}
	}
	return nil
}

func (p *Products) enableRequiredProduct(
	fc *fastly.Client,
	productID string,
	isEnabled bool,
	enableFn func(*fastly.Client, string) error,
) error {
	if !isEnabled {
		return nil
	}

	return p.Spinner.Process(
		fmt.Sprintf("Enabling product '%s'...", productID),
		func(_ *text.SpinnerWrapper) error {
			if err := enableFn(fc, p.ServiceID); err != nil {
				return fmt.Errorf("error enabling product [%s]: %w", productID, err)
			}
			return nil
		},
	)
}
