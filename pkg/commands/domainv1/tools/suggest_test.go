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
	"github.com/fastly/go-fastly/v10/fastly/domains/v1/tools/suggest"
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
			WantError: "error parsing arguments: required flag --query not provided",
		},
		{
			Args: "--query `fastly testing`",
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
			Args: "--query foo --keywords=food,kitchen --defaults=club",
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
			Args: "-j --query `fastly testing`",
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
			Args: "-v --query `fastly testing`",
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
	testutil.RunCLIScenarios(t, []string{root.CommandName, tools.CommandName, "suggest"}, scenarios)
}
