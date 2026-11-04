package pagerduty

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/alerts/pagerduty"
)

// CreateCommand calls the Fastly API to create PagerDuty alerts.
type CreateCommand struct {
	argparser.Base

	// Required.
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`
	Key         string                        `required:"" help:"PagerDuty integration key."`

	// Optional.
	Description argparser.OptionalString `help:"An optional description for the alert."`

	argparser.JSONOutput
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

	input := &pagerduty.CreateInput{
		WorkspaceID: &c.WorkspaceID.Value,
		Config: &pagerduty.CreateConfig{
			Key: &c.Key,
		},
		// Set 'Events' to the only possible value, 'flag'
		Events: alert.GetDefaultEvents(),
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := pagerduty.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created a '%s' alert '%s' (workspace-id: %s)", data.Type, data.ID, c.WorkspaceID.Value)
	return nil
}
