package products_test

import (
	"bytes"
	"testing"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
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
			Name:      "validate invalid flag combo",
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
	}

	for testcaseIdx := range scenarios {
		testcase := &scenarios[testcaseIdx]
		t.Run(testcase.Name, func(t *testing.T) {
			var stdout bytes.Buffer
			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			opts.APIClient = mock.APIClient(testcase.API)
			err := app.Run(opts)
			testutil.AssertErrorContains(t, err, testcase.WantError)
			testutil.AssertStringContains(t, stdout.String(), testcase.WantOutput)
		})
	}
}
