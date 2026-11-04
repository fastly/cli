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

// UpdateCommand calls the Fastly API to update account-level custom signals.
type UpdateCommand struct {
	argparser.Base

	// Required.
	SignalID    string `required:"" help:"Custom Signal ID"`
	Description string `required:"" help:"User submitted description of a custom signal."`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	var err error
	input := &signals.UpdateInput{
		SignalID:    &c.SignalID,
		Description: &c.Description,
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: []string{"*"},
		},
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := signals.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated account-level signal '%s' (signal-id: %s)", data.Name, data.SignalID)
	return nil
}
