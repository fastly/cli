package bot_test

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"

	root "github.com/fastly/cli/pkg/commands/botmanagement"
	sub "github.com/fastly/cli/pkg/commands/botmanagement/bot"
	fstfmt "github.com/fastly/cli/pkg/fmt"
	"github.com/fastly/cli/pkg/testutil"
)

var bot = policy.Bot{
	BotID:  "googlebot",
	Name:   "Googlebot",
	Action: policy.Action{Type: "inherit"},
}

func TestBotList(t *testing.T) {
	// Listing across the workspace returns each bot's category.
	workspaceBots := policy.Bots{
		Data: []policy.Bot{
			{BotID: "googlebot", Name: "Googlebot", CategoryID: "search-engines", Action: policy.Action{Type: "inherit"}},
			{BotID: "scrapy", Name: "Scrapy", CategoryID: "scrapers", Action: policy.Action{Type: "block"}},
		},
	}
	// Listing within a category omits it.
	categoryBots := policy.Bots{Data: []policy.Bot{bot}}

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--workspace-id is required",
			WantError: "error parsing arguments: required flag --workspace-id not provided",
		},
		{
			Name: "lists bots as a table",
			Args: "--workspace-id ws123",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(workspaceBots))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID         Name       Category ID     Action
googlebot  Googlebot  search-engines  inherit
scrapy     Scrapy     scrapers        block
`) + "\n",
		},
		{
			Name: "--category-id fills in the category column",
			Args: "--workspace-id ws123 --category-id search-engines",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(categoryBots))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID         Name       Category ID     Action
googlebot  Googlebot  search-engines  inherit
`) + "\n",
		},
		{
			Name: "--json emits the bots exactly as returned",
			Args: "--workspace-id ws123 --category-id search-engines --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(categoryBots))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(categoryBots.Data),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "list"}, scenarios)
}

func TestBotDescribe(t *testing.T) {
	scenarios := []testutil.CLIScenario{
		{
			Name:      "--bot-id is required",
			Args:      "--workspace-id ws123 --category-id search-engines",
			WantError: "error parsing arguments: required flag --bot-id not provided",
		},
		{
			Name: "describes the bot",
			Args: "--workspace-id ws123 --category-id search-engines --bot-id googlebot",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(bot))),
					},
				},
			},
			WantOutput: strings.TrimSpace(`
ID: googlebot
Name: Googlebot
Action: inherit
`),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "describe"}, scenarios)
}

func TestBotUpdate(t *testing.T) {
	updated := bot
	updated.Action.Type = "block"

	scenarios := []testutil.CLIScenario{
		{
			Name:      "--action rejects unknown values",
			Args:      "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action deny",
			WantError: "enum value must be one of allow,block,challenge,inherit, got 'deny'",
		},
		{
			Name: "prints the updated bot",
			Args: "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action block",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated bot 'Googlebot' (bot-id: googlebot, action: block)"),
		},
		{
			Name: "--action accepts inherit",
			Args: "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action inherit",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(bot))),
					},
				},
			},
			WantOutput: fstfmt.Success("Updated bot 'Googlebot' (bot-id: googlebot, action: inherit)"),
		},
		{
			Name: "--json emits the updated bot",
			Args: "--workspace-id ws123 --category-id search-engines --bot-id googlebot --action block --json",
			Client: &http.Client{
				Transport: &testutil.MockRoundTripper{
					Response: &http.Response{
						StatusCode: http.StatusOK,
						Status:     http.StatusText(http.StatusOK),
						Body:       io.NopCloser(bytes.NewReader(testutil.GenJSON(updated))),
					},
				},
			},
			WantOutput: fstfmt.EncodeJSON(updated),
		},
	}

	testutil.RunCLIScenarios(t, []string{root.CommandName, sub.CommandName, "update"}, scenarios)
}
