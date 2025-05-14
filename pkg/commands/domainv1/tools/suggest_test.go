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
				Domain:    "domainrtest.ing",
				Subdomain: "domainrtest.",
				Zone:      "ing",
			},
			{
				Domain:    "domainrtesti.ng",
				Subdomain: "domainrtesti.",
				Zone:      "ng",
			},
			{
				Domain:    "domainrtesting.com",
				Subdomain: "domainrtesting.",
				Zone:      "com",
			},
			{
				Domain:    "domainrtesting.net",
				Subdomain: "domainrtesting.",
				Zone:      "net",
			},
			{
				Domain:    "domainrtest.in",
				Subdomain: "domainrtest.",
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
			Args: "--query `domainr testing`",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(testSuggestions))),
					},
				},
			},
			WantOutput: `Domain              Subdomain        Zone  Path
domainrtest.ing     domainrtest.     ing   
domainrtesti.ng     domainrtesti.    ng    
domainrtesting.com  domainrtesting.  com   
domainrtesting.net  domainrtesting.  net   
domainrtest.in      domainrtest.     in    /g
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
			Args: "-j --query `domainr testing`",
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
      "domain": "domainrtest.ing",
      "subdomain": "domainrtest.",
      "zone": "ing"
    },
    {
      "domain": "domainrtesti.ng",
      "subdomain": "domainrtesti.",
      "zone": "ng"
    },
    {
      "domain": "domainrtesting.com",
      "subdomain": "domainrtesting.",
      "zone": "com"
    },
    {
      "domain": "domainrtesting.net",
      "subdomain": "domainrtesting.",
      "zone": "net"
    },
    {
      "domain": "domainrtest.in",
      "subdomain": "domainrtest.",
      "zone": "in",
      "path": "/g"
    }
  ]
}
`,
		},
		{
			Args: "-v --query `domainr testing`",
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

Domain: domainrtest.ing
Subdomain: domainrtest.
Zone: ing

Domain: domainrtesti.ng
Subdomain: domainrtesti.
Zone: ng

Domain: domainrtesting.com
Subdomain: domainrtesting.
Zone: com

Domain: domainrtesting.net
Subdomain: domainrtesting.
Zone: net

Domain: domainrtest.in
Subdomain: domainrtest.
Zone: in
Path: /g
`,
		},
	}
	testutil.RunCLIScenarios(t, []string{root.CommandName, tools.CommandName, "suggest"}, scenarios)
}
