package domain_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/domainmanagement/v1/tools/suggest"

	"github.com/fastly/cli/pkg/commands/tools"
	"github.com/fastly/cli/pkg/commands/tools/domain"
	"github.com/fastly/cli/pkg/testutil"
)

func TestNewDomainsV1ToolsSuggestCommand(t *testing.T) {
	testSuggestions := suggest.Suggestions{
		Results: []suggest.Suggestion{
			{
				Domain:    "fastlytest.ing",
				Subdomain: "fastlytest.",
				Zone:      "ing",
			},
			{
				Domain:    "fastlytesti.ng",
				Subdomain: "fastlytesti.",
				Zone:      "ng",
			},
			{
				Domain:    "fastlytesting.com",
				Subdomain: "fastlytesting.",
				Zone:      "com",
			},
			{
				Domain:    "fastlytesting.net",
				Subdomain: "fastlytesting.",
				Zone:      "net",
			},
			{
				Domain:    "fastlytest.in",
				Subdomain: "fastlytest.",
				Zone:      "in",
				Path:      fastly.ToPointer("/g"),
			},
		},
	}
	scenarios := []testutil.CLIScenario{
		{
			Args:      "",
			WantError: "error parsing arguments: required argument 'query' not provided",
		},
		{
			Args: "fastly testing",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testSuggestions))),
					},
				},
			},
			WantOutput: `Domain             Subdomain       Zone  Path
fastlytest.ing     fastlytest.     ing   
fastlytesti.ng     fastlytesti.    ng    
fastlytesting.com  fastlytesting.  com   
fastlytesting.net  fastlytesting.  net   
fastlytest.in      fastlytest.     in    /g
`,
		},
		{
			Args: "--keywords=food,kitchen --defaults=club foo",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body: io.NopCloser(bytes.NewReader(testutil.GenJSON(suggest.Suggestions{
							Results: []suggest.Suggestion{
								{
									Domain:    "foo.eat",
									Subdomain: "foo.",
									Zone:      "eat",
								},
								{
									Domain:    "foo.cafe",
									Subdomain: "foo.",
									Zone:      "cafe",
								},
								{
									Domain:    "foo.menu",
									Subdomain: "foo.",
									Zone:      "menu",
								},
								{
									Domain:    "foo.kitchen",
									Subdomain: "foo.",
									Zone:      "kitchen",
								},
								{
									Domain:    "foo.club",
									Subdomain: "foo.",
									Zone:      "club",
								},
							},
						}))),
					},
				},
			},
			WantOutput: `Domain       Subdomain  Zone     Path
foo.eat      foo.       eat      
foo.cafe     foo.       cafe     
foo.menu     foo.       menu     
foo.kitchen  foo.       kitchen  
foo.club     foo.       club     
`,
		},
		{
			Args: "-j fastly testing",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testSuggestions))),
					},
				},
			},
			WantOutput: `{
  "results": [
    {
      "domain": "fastlytest.ing",
      "subdomain": "fastlytest.",
      "zone": "ing"
    },
    {
      "domain": "fastlytesti.ng",
      "subdomain": "fastlytesti.",
      "zone": "ng"
    },
    {
      "domain": "fastlytesting.com",
      "subdomain": "fastlytesting.",
      "zone": "com"
    },
    {
      "domain": "fastlytesting.net",
      "subdomain": "fastlytesting.",
      "zone": "net"
    },
    {
      "domain": "fastlytest.in",
      "subdomain": "fastlytest.",
      "zone": "in",
      "path": "/g"
    }
  ]
}
`,
		},
		{
			Args: "-v fastly testing",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testSuggestions))),
					},
				},
			},
			WantOutput: `Fastly API endpoint: https://api.fastly.com
Fastly API token provided via config file (profile: user)

Domain: fastlytest.ing
Subdomain: fastlytest.
Zone: ing

Domain: fastlytesti.ng
Subdomain: fastlytesti.
Zone: ng

Domain: fastlytesting.com
Subdomain: fastlytesting.
Zone: com

Domain: fastlytesting.net
Subdomain: fastlytesting.
Zone: net

Domain: fastlytest.in
Subdomain: fastlytest.
Zone: in
Path: /g
`,
		},
	}
	testutil.RunCLIScenarios(t, []string{tools.CommandName, domain.CommandName, "suggest"}, scenarios)
}
