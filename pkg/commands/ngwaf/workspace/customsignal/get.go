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

// GetCommand calls the Fastly API to get a workspace-level custom signal.
type GetCommand struct {
	argparser.Base

	// Required.
	SignalID    string                        `required:"" help:"Custom Signal ID"`
	WorkspaceID argparser.OptionalWorkspaceID `required:"" help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &signals.GetInput{
		SignalID: &c.SignalID,
		Scope: &scope.Scope{
			Type: scope.ScopeTypeWorkspace,
		},
	}

	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	input.Scope.AppliesTo = []string{c.WorkspaceID.Value}

	data, err := signals.Get(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintCustomSignal(out, data)
	return nil
}
