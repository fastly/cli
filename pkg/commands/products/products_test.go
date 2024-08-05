package products_test

import (
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	root "github.com/fastly/cli/pkg/commands/products"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
)

func TestProductEnablement(t *testing.T) {
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing Service ID",
			WantError: "failed to identify Service ID: error reading service: no service ID found",
		},
		{
			Name:      "validate invalid enable/disable flag combo",
			Arg:       "--enable fanout --disable fanout",
			WantError: "invalid flag combination: --enable and --disable",
		},
		{
			Name: "validate API error for product status",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, testutil.Err
				},
			},
			Arg: "--service-id 123",
			WantOutput: `PRODUCT             ENABLED
brotli_compression  false
domain_inspector    false
fanout              false
image_optimizer     false
origin_inspector    false
websockets          false
`,
		},
		{
			Name: "validate API success for product status",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, nil
				},
			},
			Arg: "--service-id 123",
			WantOutput: `PRODUCT             ENABLED
brotli_compression  true
domain_inspector    true
fanout              true
image_optimizer     true
origin_inspector    true
websockets          true
`,
		},
		{
			Name:      "validate flag parsing error for enabling product",
			Arg:       "--service-id 123 --enable foo",
			WantError: "error parsing arguments: enum value must be one of brotli_compression,domain_inspector,fanout,image_optimizer,origin_inspector,websockets, got 'foo'",
		},
		{
			Name:      "validate flag parsing error for disabling product",
			Arg:       "--service-id 123 --disable foo",
			WantError: "error parsing arguments: enum value must be one of brotli_compression,domain_inspector,fanout,image_optimizer,origin_inspector,websockets, got 'foo'",
		},
		{
			Name: "validate success for enabling product",
			API: mock.API{
				EnableProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, nil
				},
			},
			Arg:        "--service-id 123 --enable brotli_compression",
			WantOutput: "SUCCESS: Successfully enabled product 'brotli_compression'",
		},
		{
			Name: "validate success for disabling product",
			API: mock.API{
				DisableProductFn: func(i *fastly.ProductEnablementInput) error {
					return nil
				},
			},
			Arg:        "--service-id 123 --disable brotli_compression",
			WantOutput: "SUCCESS: Successfully disabled product 'brotli_compression'",
		},
		{
			Name:      "validate invalid json/verbose flag combo",
			Arg:       "--service-id 123 --json --verbose",
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate API error for product status with --json output",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, testutil.Err
				},
			},
			Arg: "--service-id 123 --json",
			WantOutput: `{
  "brotli_compression": false,
  "domain_inspector": false,
  "fanout": false,
  "image_optimizer": false,
  "origin_inspector": false,
  "websockets": false
}`,
		},
		{
			Name: "validate API success for product status with --json output",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, nil
				},
			},
			Arg: "--service-id 123 --json",
			WantOutput: `{
  "brotli_compression": true,
  "domain_inspector": true,
  "fanout": true,
  "image_optimizer": true,
  "origin_inspector": true,
  "websockets": true
}`,
		},
	}

	testutil.RunScenarios(t, []string{root.CommandName}, scenarios)
}
