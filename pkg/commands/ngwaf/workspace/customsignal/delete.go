package customsignal

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/signals"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a workspace-level custom signal.
type DeleteCommand struct {
	argparser.Base

	// Required.
	SignalID    string                        `required:"" help:"Custom Signal ID"`
	WorkspaceID argparser.OptionalWorkspaceID `required:"" help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &signals.DeleteInput{
		SignalID: &c.SignalID,
		Scope: &scope.Scope{
			Type: scope.ScopeTypeWorkspace,
		},
	}
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	input.Scope.AppliesTo = []string{c.WorkspaceID.Value}

	err := signals.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.SignalID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted workspace-level custom signal (id: %s)", c.SignalID)
	return nil
}
