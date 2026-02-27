package customsignal

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/signals"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update workspace-level custom signals.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	signalID    string
	description string
	workspaceID argparser.OptionalWorkspaceID
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a workspace")

	// Required.
	c.CmdClause.Flag("signal-id", "Custom Signal ID").Required().StringVar(&c.signalID)
	c.CmdClause.Flag("description", "User submitted description of a custom signal.").Required().StringVar(&c.description)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	var err error
	input := &signals.UpdateInput{
		SignalID:    &c.signalID,
		Description: &c.description,
		Scope: &scope.Scope{
			Type: scope.ScopeTypeWorkspace,
		},
	}

	if err := c.workspaceID.Parse(); err != nil {
		return err
	}
	input.Scope.AppliesTo = []string{c.workspaceID.Value}

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

	text.Success(out, "Updated workspace-level signal '%s' (signal-id: %s)", data.Name, data.SignalID)
	return nil
}
