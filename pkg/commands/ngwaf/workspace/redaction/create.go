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

// CreateCommand calls the Fastly API to create a redaction.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	field         string
	redactionType string
	workspaceID   argparser.OptionalWorkspaceID
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a redaction").Alias("add")

	// Required.
	c.CmdClause.Flag("field", "The name of the field that should be redacted.").Required().StringVar(&c.field)
	c.CmdClause.Flag("type", "The type of field that is being redacted.").Required().StringVar(&c.redactionType)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var err error
	input := &redactions.CreateInput{
		Field:       &c.field,
		Type:        &c.redactionType,
		WorkspaceID: &c.workspaceID.Value,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := redactions.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created redaction '%s' (field: %s, type: %s)", data.RedactionID, data.Field, data.Type)
	return nil
}
