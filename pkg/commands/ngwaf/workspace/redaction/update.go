package redaction

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/redactions"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update redactions.
type UpdateCommand struct {
	argparser.Base

	// Required.
	RedactionID string                        `required:"" help:"A base62-encoded representation of a UUID used to uniquely identify a redaction."`
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	Field         argparser.OptionalString `help:"The name of the field that should be redacted."`
	RedactionType argparser.OptionalString `name:"type" help:"The type of field that is being redacted."`

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

	var err error
	input := &redactions.UpdateInput{
		RedactionID: &c.RedactionID,
		WorkspaceID: &c.WorkspaceID.Value,
	}

	if c.Field.WasSet {
		input.Field = &c.Field.Value
	}

	if c.RedactionType.WasSet {
		input.Type = &c.RedactionType.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := redactions.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated redaction '%s' (field: %s, type: %s)", data.RedactionID, data.Field, data.Type)
	return nil
}
