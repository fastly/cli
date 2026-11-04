package virtualpatch

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/virtualpatches"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update virtual patches.
type UpdateCommand struct {
	argparser.Base

	// Required.
	VirtualpatchID string `name:"virtual-patch-id" required:"" help:"Virtual Patch ID"`
	WorkspaceID    string `required:"" help:"Workspace ID"`

	// Optional.
	EnabledFlag argparser.OptionalString `name:"enabled" help:"Specify the toggle status indicator of the virtual patch."`
	Mode        argparser.OptionalString `help:"Specify the action to take when a signal for virtual patch is detected."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	var err error
	input := &virtualpatches.UpdateInput{
		VirtualPatchID: &c.VirtualpatchID,
		WorkspaceID:    &c.WorkspaceID,
	}
	if c.EnabledFlag.WasSet {
		enabled, err := argparser.ConvertBoolFromStringFlag(c.EnabledFlag.Value, "enabled")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.Enabled = enabled
	}
	if c.Mode.WasSet {
		input.Mode = &c.Mode.Value
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
