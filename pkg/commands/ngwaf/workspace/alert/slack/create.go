package slack

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/common"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/slack"
)

// CreateCommand calls the Fastly API to create Slack alerts.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	common.BaseAlertFlags
	common.WebhookConfigFlags

	// Optional.
	common.AlertDataFlags
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Slack alert").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:          argparser.FlagNGWAFWorkspaceID,
		Description:   argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:           &c.WorkspaceID.Value,
		Action:        c.WorkspaceID.Set,
		ForceRequired: true,
	})
	c.CmdClause.Flag("webhook", "Slack webhook.").Required().StringVar(&c.Webhook)

	// Optional.
	c.CmdClause.Flag("description", "An optional description for the alert.").Action(c.Description.Set).StringVar(&c.Description.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &slack.CreateInput{
		WorkspaceID: &c.WorkspaceID.Value,
		Config: &slack.CreateConfig{
			Webhook: &c.Webhook,
		},
		// Set 'Events' to the only possible value, 'flag'
		Events: common.GetDefaultEvents(),
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := slack.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created a '%s' alert '%s' (workspace-id: %s)", data.Type, data.ID, c.WorkspaceID.Value)
	return nil
}
