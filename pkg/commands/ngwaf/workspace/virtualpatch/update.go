package virtualpatch

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/virtualpatches"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update virtual patches.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	virtualpatchID string
	workspaceID    string

	// Optional.
	enabled argparser.OptionalString
	mode    argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a virtual patch")

	// Required.
	c.CmdClause.Flag("virtual-patch-id", "Virtual Patch ID").Required().StringVar(&c.virtualpatchID)
	c.CmdClause.Flag("workspace-id", "Workspace ID").Required().StringVar(&c.workspaceID)

	// Optional.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        "enabled",
		Description: "Specify the toggle status indicator of the virtual patch.",
		Action:      c.enabled.Set,
		Dst:         &c.enabled.Value,
	})
	c.CmdClause.Flag("mode", "Specify the action to take when a signal for virtual patch is detected.").Action(c.mode.Set).StringVar(&c.mode.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var err error
	input := &virtualpatches.UpdateInput{
		VirtualPatchID: &c.virtualpatchID,
		WorkspaceID:    &c.workspaceID,
	}
	if c.enabled.WasSet {
		enabled, err := argparser.ConvertBoolFromStringFlag(c.enabled.Value)
		if err != nil {
			err := errors.New("'enabled' flag must be one of the following [true, false]")
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.Enabled = enabled
	}
	if c.mode.WasSet {
		input.Mode = &c.mode.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := virtualpatches.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated virtual patch '%s' (enabled: %t, mode: %s)", data.ID, data.Enabled, data.Mode)
	return nil
}
