package category_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/category"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
)

var category = policy.Category{
	CategoryID:  "search-engines",
	Name:        "Search Engines",
	Description: "Crawlers operated by search providers",
	Action:      policy.Action{Type: "allow"},
}

func TestCategoryList(t *testing.T) {
	page := policy.Categories{
		Data: []policy.Category{
			category,
			{CategoryID: "scrapers", Name: "Scrapers", Action: policy.Action{Type: "block"}},
		},
	}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "lists categories as a table",
			Args: "--workspace-id ws123",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(page))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID              Name            Action
search-engines  Search Engines  allow
scrapers        Scrapers        block
`) + "\n",
		},
		{
			Name: "--json emits the categories",
			Args: "--workspace-id ws123 --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(page))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(page.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestCategoryDescribe(t *testing.T) {

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--category-id is required",
			Args:      "--workspace-id ws123",
			WantError: "error parsing arguments: required flag --category-id not provided",
		},
		{
			Name: "describes the category",
			Args: "--workspace-id ws123 --category-id search-engines",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(category))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID: search-engines
Name: Search Engines
Description: Crawlers operated by search providers
Action: allow
`),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestCategoryUpdate(t *testing.T) {
	updated := category
	updated.Action.Type = "challenge"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--action is required",
			Args:      "--workspace-id ws123 --category-id search-engines",
			WantError: "error parsing arguments: required flag --action not provided",
		},
		{
			Name:      "categories cannot inherit",
			Args:      "--workspace-id ws123 --category-id search-engines --action inherit",
			WantError: "enum value must be one of allow,block,challenge, got 'inherit'",
		},
		{
			Name: "prints the updated category",
			Args: "--workspace-id ws123 --category-id search-engines --action challenge",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated bot category 'Search Engines' (category-id: search-engines, action: challenge)"),
		},
		{
			Name: "API error is surfaced",
			Args: "--workspace-id ws123 --category-id nope --action block",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusNotFound,
						Status:     http.StatusText(http.StatusNotFound),
						Body:       io.NopCloser(bytes.NewReader([]byte(`{"title":"not found","status":404}`))),
					},
				},
			},
			WantError: "404 - Not Found",
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
