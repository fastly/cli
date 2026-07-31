package integration

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/integration/datadog"
	"github.com/fastly/cli/pkg/commands/integration/jiraissue"
	"github.com/fastly/cli/pkg/commands/integration/jsm"
	"github.com/fastly/cli/pkg/commands/integration/mail"
	"github.com/fastly/cli/pkg/commands/integration/msteams"
	"github.com/fastly/cli/pkg/commands/integration/newrelic"
	"github.com/fastly/cli/pkg/commands/integration/opsgenie"
	"github.com/fastly/cli/pkg/commands/integration/pagerduty"
	"github.com/fastly/cli/pkg/commands/integration/slack"
	"github.com/fastly/cli/pkg/commands/integration/splunkoncall"
	"github.com/fastly/cli/pkg/commands/integration/webhook"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// knownIntegrationTypes lists the integration type values with dedicated CLI
// sub-families, offered as shell-completion hints for --type. Other type
// values (e.g. legacy integration types) are still accepted.
var knownIntegrationTypes = []string{
	datadog.CommandName,
	jiraissue.CommandName,
	jsm.CommandName,
	mail.CommandName,
	msteams.CommandName,
	newrelic.CommandName,
	opsgenie.CommandName,
	pagerduty.CommandName,
	slack.CommandName,
	splunkoncall.CommandName,
	webhook.CommandName,
}

// ListCommand calls the Fastly API to list notification integrations.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Cursor argparser.OptionalString
	Limit  argparser.OptionalInt
	Sort   argparser.OptionalString
	Type   argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List notification integrations")

	// Optional.
	c.CmdClause.Flag("cursor", "Pagination cursor from a previous response's meta").Action(c.Cursor.Set).StringVar(&c.Cursor.Value)
	c.RegisterFlagBool(c.JSONFlag())
	c.CmdClause.Flag("limit", "Maximum number of items to return").Action(c.Limit.Set).IntVar(&c.Limit.Value)
	c.CmdClause.Flag("sort", "Field to sort results by").Action(c.Sort.Set).StringVar(&c.Sort.Value)
	c.CmdClause.Flag("type", "Filter integrations by type").HintOptions(knownIntegrationTypes...).Action(c.Type.Set).StringVar(&c.Type.Value)

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &fastly.SearchIntegrationsInput{}
	if c.Cursor.WasSet {
		input.Cursor = &c.Cursor.Value
	}
	if c.Limit.WasSet {
		input.Limit = &c.Limit.Value
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}
	if c.Type.WasSet {
		input.Type = &c.Type.Value
	}

	o, err := c.Globals.APIClient.SearchIntegrations(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	text.PrintIntegrationsTbl(out, o.Data)

	return nil
}
