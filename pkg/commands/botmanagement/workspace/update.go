package workspace

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ProtectionModes are the accepted values for a workspace's protection mode.
var ProtectionModes = []string{"off", "log", "block"}

// UpdateCommand calls the Fastly API to update a Bot Management workspace.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	workspaceID string

	// Optional.
	description    argparser.OptionalString
	name           argparser.OptionalString
	protectionMode argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Bot Management workspace")

	// Required.
	c.CmdClause.Flag("workspace-id", "Workspace ID").Required().StringVar(&c.workspaceID)

	// Optional.
	c.CmdClause.Flag("description", "Description of the workspace").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("name", "Display name of the workspace").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("protection-mode", "Protection mode of the workspace").HintOptions(ProtectionModes...).Action(c.protectionMode.Set).EnumVar(&c.protectionMode.Value, ProtectionModes...)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if !c.description.WasSet && !c.name.WasSet && !c.protectionMode.WasSet {
		return fsterr.RemediationError{
			Inner:       errors.New("no workspace fields to update"),
			Remediation: "Provide at least one of --description, --name, or --protection-mode.",
		}
	}

	input := &workspaces.UpdateInput{
		WorkspaceID: &c.workspaceID,
	}
	if c.description.WasSet {
		input.Description = &c.description.Value
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.protectionMode.WasSet {
		input.ProtectionMode = &c.protectionMode.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := workspaces.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated Bot Management workspace '%s' (workspace-id: %s, protection-mode: %s)", data.Name, data.ID, data.ProtectionMode)
	return nil
}
