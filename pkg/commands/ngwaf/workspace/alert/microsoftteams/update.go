package microsoftteams

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/alerts/microsoftteams"
)

// UpdateCommand calls the Fastly API to update Microsoft Teams alerts.
type UpdateCommand struct {
	argparser.Base

	// Required.
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`
	AlertID     string                        `required:"" help:"Alphanumeric string identifying the alert"`
	Webhook     string                        `required:"" help:"Microsoft Teams webhook."`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &microsoftteams.UpdateInput{
		AlertID:     &c.AlertID,
		WorkspaceID: &c.WorkspaceID.Value,
		Config: &microsoftteams.UpdateConfig{
			Webhook: &c.Webhook,
		},
		// Set 'Events' to the only possible value, 'flag'
		Events: alert.GetDefaultEvents(),
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := microsoftteams.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated '%s' alert '%s' (workspace-id: %s)", data.Type, data.ID, c.WorkspaceID.Value)
	return nil
}
