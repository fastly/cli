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
			WantError: "error parsing arguments: required flag --domain not provided",
		},
		{
			Args: "--domain domainr-testing.com",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(status.Status{
							Domain: "domainr-testing.com",
							Zone:   "com",
							Status: "undelegated inactive",
							Tags:   "generic",
						}))),
					},
				},
			},
			WantOutput: `Domain: domainr-testing.com
Zone: com
Status: undelegated inactive
Tags: generic
`,
		},
		{
			Args: "--domain domainr-testing.org --scope estimate",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(status.Status{
							Domain: "domainr-testing.org",
							Zone:   "org",
							Status: "marketed priced transferable active",
							Tags:   "generic",
							Scope:  fastly.ToPointer(status.ScopeEstimate),
							Offers: []status.Offer{
								{
									Vendor:   "am.godaddy.com",
									Currency: "USD",
									Price:    "25000.00",
								},
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
			WantOutput: `Domain: domainr-testing.org
Zone: org
Status: marketed priced transferable active
Tags: generic
Scope: estimate
Offers:
  - Vendor: am.godaddy.com
    Currency: USD
    Price: 25000.00
  - Vendor: example.com
    Currency: USD
    Price: 20000.00
`,
		},
		{
			Args: "-j --domain domainr-testing.org --scope estimate",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(status.Status{
							Domain: "domainr-testing.org",
							Zone:   "org",
							Status: "marketed priced transferable active",
							Tags:   "generic",
							Scope:  fastly.ToPointer(status.ScopeEstimate),
							Offers: []status.Offer{
								{
									Vendor:   "am.godaddy.com",
									Currency: "USD",
									Price:    "25000.00",
								},
							},
						}))),
					},
				},
			},
			WantOutput: `{
  "domain": "domainr-testing.org",
  "zone": "org",
  "status": "marketed priced transferable active",
  "scope": "estimate",
  "tags": "generic",
  "offers": [
    {
      "vendor": "am.godaddy.com",
      "price": "25000.00",
      "currency": "USD"
    }
  ]
}
`,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, tools.CommandName, "status"}, scenarios)
}
