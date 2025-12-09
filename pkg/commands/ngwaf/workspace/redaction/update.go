package redaction

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/redactions"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update redactions.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	redactionID string
	workspaceID argparser.OptionalWorkspaceID

	// Optional.
	field         argparser.OptionalString
	redactionType argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a redaction")

	// Required.
	c.CmdClause.Flag("redaction-id", "A base62-encoded representation of a UUID used to uniquely identify a redaction.").Required().StringVar(&c.redactionID)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.CmdClause.Flag("field", "The name of the field that should be redacted.").Action(c.field.Set).StringVar(&c.field.Value)
	c.CmdClause.Flag("type", "The type of field that is being redacted.").Action(c.redactionType.Set).StringVar(&c.redactionType.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var err error
	input := &redactions.UpdateInput{
		RedactionID: &c.redactionID,
		WorkspaceID: &c.workspaceID.Value,
	}

	if c.field.WasSet {
		input.Field = &c.field.Value
	}

	if c.redactionType.WasSet {
		input.Type = &c.redactionType.Value
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
