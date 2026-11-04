package integration

import (
	"io"

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
)

// CommandName is the string to be used to invoke this command.
const CommandName = "integration"

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	List         ListCommand              `cmd:"" help:"List notification integrations"`
	Describe     DescribeCommand          `cmd:"" aliases:"get" help:"Retrieve a single notification integration"`
	Delete       DeleteCommand            `cmd:"" help:"Delete a notification integration"`
	ListTypes    ListTypesCommand         `cmd:"" help:"List supported notification integration types"`
	Datadog      datadog.RootCommand      `cmd:"" help:"Manage Datadog notification integrations"`
	Jiraissue    jiraissue.RootCommand    `cmd:"" help:"Manage Jira Issue notification integrations"`
	JSM          jsm.RootCommand          `cmd:"" help:"Manage Jira Service Management notification integrations"`
	Opsgenie     opsgenie.RootCommand     `cmd:"" help:"Manage OpsGenie notification integrations"`
	Splunkoncall splunkoncall.RootCommand `cmd:"" help:"Manage Splunk On-Call notification integrations"`
	Mail         mail.RootCommand         `cmd:"" help:"Manage Mailing List notification integrations"`
	Msteams      msteams.RootCommand      `cmd:"" help:"Manage Microsoft Teams notification integrations"`
	Newrelic     newrelic.RootCommand     `cmd:"" help:"Manage New Relic notification integrations"`
	Pagerduty    pagerduty.RootCommand    `cmd:"" help:"Manage PagerDuty notification integrations"`
	Slack        slack.RootCommand        `cmd:"" help:"Manage Slack notification integrations"`
	Webhook      webhook.RootCommand      `cmd:"" help:"Manage generic Webhook notification integrations"`
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
