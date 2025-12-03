package alert

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// CommandName is the string to be used to invoke this command.
const CommandName = "alert"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "Manage workspace alerts")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

// WorkspaceIDFlags contains the workspace ID flag used by all alert commands.
type WorkspaceIDFlags struct {
	WorkspaceID argparser.OptionalWorkspaceID
}

// AlertIDFlags contains flags for identifying a specific alert (used in update/delete/get).
type AlertIDFlags struct {
	AlertID string
}

// AlertDataFlags contains optional data fields for alerts (used in create/update).
type AlertDataFlags struct {
	Description argparser.OptionalString
}

// KeyConfigFlags contains the Key configuration (used by opsgenie, pagerduty).
type KeyConfigFlags struct {
	Key string
}

// WebhookConfigFlags contains the Webhook configuration (used by webhook, slack, microsoftteams).
type WebhookConfigFlags struct {
	Webhook string
}

// GetDefaultEvents returns the hardcoded events value for all alerts.
// Currently the only supported value is "flag".
func GetDefaultEvents() *[]string {
	events := []string{"flag"}
	return &events
}
