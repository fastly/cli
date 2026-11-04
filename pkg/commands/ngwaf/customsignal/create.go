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

// CreateCommand calls the Fastly API to create account-level custom signals.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" required:"" help:"User submitted display name of a custom signal. Is immutable and must be between 3 and 25 characters"`

	// Optional.
	Description argparser.OptionalString `help:"User submitted description of a custom signal."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	var err error
	input := &signals.CreateInput{
		Name: &c.NameFlag,
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: []string{"*"},
		},
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := signals.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created account-level custom signal '%s' (signal-id: %s)", data.Name, data.SignalID)
	return nil
}
