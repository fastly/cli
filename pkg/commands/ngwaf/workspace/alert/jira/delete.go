package jira

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/alerts/jira"
)

// DeleteCommand calls the Fastly API to delete Jira alerts.
type DeleteCommand struct {
	argparser.Base

	// Required.
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`
	AlertID     string                        `required:"" help:"Alphanumeric string identifying the alert"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	err := jira.Delete(context.TODO(), fc, &jira.DeleteInput{
		WorkspaceID: &c.WorkspaceID.Value,
		AlertID:     &c.AlertID,
	})
	if err != nil {
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.AlertID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted alert '%s' (workspace-id: %s)", c.AlertID, c.WorkspaceID.Value)
	return nil
}
