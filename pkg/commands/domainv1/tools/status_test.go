package tools_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	root "github.com/fastly/cli/pkg/commands/domainv1"
	"github.com/fastly/cli/pkg/commands/domainv1/tools"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/domains/v1/tools/status"
)

func TestNewDomainsV1ToolsStatusCommand(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required argument 'domain' not provided",
		},
		{
			Args:      "fastly-cli-testing.com --scope not-estimate",
			WantError: "invalid scope provided",
		},
		{
			Args: "fastly-cli-testing.com",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(status.Status{
							Domain: "fastly-cli-testing.com",
							Zone:   "com",
							Status: "undelegated inactive",
							Tags:   "generic",
						}))),
					},
				},
			},
			WantOutput: `Domain: fastly-cli-testing.com
Zone: com
Status: undelegated inactive
Tags: generic
`,
		},
		{
			Args: "--scope estimate fastly-cli-testing-offers.com",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(status.Status{
							Domain: "fastly-cli-testing-offers.com",
							Zone:   "com",
							Status: "marketed priced transferable active",
							Tags:   "generic",
							Scope:  fastly.ToPointer(status.ScopeEstimate),
							Offers: []status.Offer{
								{
									Vendor:   "example.com",
									Currency: "USD",
									Price:    "20000.00",
								},
							},
						}))),
					},
				},
			},
			WantOutput: `Domain: fastly-cli-testing-offers.com
Zone: com
Status: marketed priced transferable active
Tags: generic
Scope: estimate
Offers:
  - Vendor: example.com
    Currency: USD
    Price: 20000.00
`,
		},
		{
			Args: "-j --scope estimate fastly-cli-testing-offers.com",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(status.Status{
							Domain: "fastly-cli-testing-offers.com",
							Zone:   "com",
							Status: "marketed priced transferable active",
							Tags:   "generic",
							Scope:  fastly.ToPointer(status.ScopeEstimate),
							Offers: []status.Offer{
								{
									Vendor:   "example.com",
									Currency: "USD",
									Price:    "20000.00",
								},
							},
						}))),
					},
				},
			},
			WantOutput: `{
  "domain": "fastly-cli-testing-offers.com",
  "zone": "com",
  "status": "marketed priced transferable active",
  "scope": "estimate",
  "tags": "generic",
  "offers": [
    {
      "vendor": "example.com",
      "price": "20000.00",
      "currency": "USD"
    }
  ]
}
`,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, tools.CommandName, "status"}, scenarios)
}
