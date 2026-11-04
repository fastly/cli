package integration

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

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

	// Optional.
	Cursor argparser.OptionalString `help:"Pagination cursor from a previous response's meta"`
	argparser.JSONOutput

	Limit argparser.OptionalInt    `help:"Maximum number of items to return"`
	Sort  argparser.OptionalString `help:"Field to sort results by"`
	Type  argparser.OptionalString `hints:"datadog,jiraissue,jsm,mail,msteams,newrelic,opsgenie,pagerduty,slack,splunkoncall,webhook" help:"Filter integrations by type"`
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
