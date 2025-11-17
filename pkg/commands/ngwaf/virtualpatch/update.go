package virtualpatch

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/virtualpatches"

	"github.com/fastly/cli/pkg/argparser"
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
		Description: "Specify the the toggle status indicator of the VirtualPatch.",
		Action:      c.enabled.Set,
		Dst:         &c.enabled.Value,
	})
	c.CmdClause.Flag("mode", "Specify the action to take when a signal for virtual patch is detected.").Action(c.mode.Set).StringVar(&c.mode.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	var err error
	input := &virtualpatches.UpdateInput{
		VirtualPatchID: &c.virtualpatchID,
		WorkspaceID:    &c.workspaceID,
	}
	if c.enabled.WasSet {
		var enableToggle bool
		switch c.enabled.Value {
		case "true":
			enableToggle = true
		case "false":
			enableToggle = false
		default:
			return fmt.Errorf("'enabled' flag must be one of the following [true, false]")
		}
		input.Enabled = &enableToggle
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
