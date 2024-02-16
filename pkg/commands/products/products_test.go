package products_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/app"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/mock"
	"github.com/fastly/cli/v10/pkg/testutil"
)

func TestProductEnablement(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "validate missing Service ID",
			Args:      args("products"),
			WantError: "failed to identify Service ID: error reading service: no service ID found",
		},
		{
			Name:      "validate invalid enable/disable flag combo",
			Args:      args("products --enable fanout --disable fanout"),
			WantError: "invalid flag combination: --enable and --disable",
		},
		{
			Name: "validate API error for product status",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, testutil.Err
				},
			},
			Args: args("products --service-id 123"),
			WantOutput: `PRODUCT             ENABLED
Brotli Compression  false
Domain Inspector    false
Fanout              false
Image Optimizer     false
Origin Inspector    false
Web Sockets         false
`,
		},
		{
			Name: "validate API success for product status",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, nil
				},
			},
			Args: args("products --service-id 123"),
			WantOutput: `PRODUCT             ENABLED
Brotli Compression  true
Domain Inspector    true
Fanout              true
Image Optimizer     true
Origin Inspector    true
Web Sockets         true
`,
		},
		{
			Name:      "validate flag parsing error for enabling product",
			Args:      args("products --service-id 123 --enable foo"),
			WantError: "error parsing arguments: enum value must be one of brotli_compression,domain_inspector,fanout,image_optimizer,origin_inspector,websockets, got 'foo'",
		},
		{
			Name:      "validate flag parsing error for disabling product",
			Args:      args("products --service-id 123 --disable foo"),
			WantError: "error parsing arguments: enum value must be one of brotli_compression,domain_inspector,fanout,image_optimizer,origin_inspector,websockets, got 'foo'",
		},
		{
			Name: "validate success for enabling product",
			API: mock.API{
				EnableProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, nil
				},
			},
			Args:       args("products --service-id 123 --enable brotli_compression"),
			WantOutput: "SUCCESS: Successfully enabled product 'brotli_compression'",
		},
		{
			Name: "validate success for disabling product",
			API: mock.API{
				DisableProductFn: func(i *fastly.ProductEnablementInput) error {
					return nil
				},
			},
			Args:       args("products --service-id 123 --disable brotli_compression"),
			WantOutput: "SUCCESS: Successfully disabled product 'brotli_compression'",
		},
		{
			Name:      "validate invalid json/verbose flag combo",
			Args:      args("products --service-id 123 --json --verbose"),
			WantError: "invalid flag combination, --verbose and --json",
		},
		{
			Name: "validate API error for product status with --json output",
			API: mock.API{
				GetProductFn: func(i *fastly.ProductEnablementInput) (*fastly.ProductEnablement, error) {
					return nil, testutil.Err
				},
			},
			Args: args("products --service-id 123 --json"),
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
			Args: args("products --service-id 123 --json"),
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

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			app.Init = func(_ []string, _ io.Reader) (*global.Data, error) {
				opts := testutil.MockGlobalData(testcase.Args, &stdout)
				opts.APIClientFactory = mock.APIClient(testcase.API)
				return opts, nil
			}
			err := app.Run(testcase.Args, nil)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
