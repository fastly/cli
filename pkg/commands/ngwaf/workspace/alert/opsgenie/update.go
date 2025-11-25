package opsgenie

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/common"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/opsgenie"
)

// UpdateCommand calls the Fastly API to update Opsgenie alerts.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	common.AlertIDFlags
	common.BaseAlertFlags
	common.KeyConfigFlags
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Opsgenie alert")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:          argparser.FlagNGWAFWorkspaceID,
		Description:   argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:           &c.WorkspaceID.Value,
		Action:        c.WorkspaceID.Set,
		ForceRequired: true,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFAlertID,
		Description: argparser.FlagNGWAFAlertIDDesc,
		Dst:         &c.AlertID,
		Required:    true,
	})
	c.CmdClause.Flag("key", "Opsgenie integration key.").Required().StringVar(&c.Key)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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

	input := &opsgenie.UpdateInput{
		AlertID:     &c.AlertID,
		WorkspaceID: &c.WorkspaceID.Value,
		Config: &opsgenie.UpdateConfig{
			Key: &c.Key,
		},
		// Set 'Events' to the only possible value, 'flag'
		Events: common.GetDefaultEvents(),
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := opsgenie.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated '%s' alert '%s' (workspace-id: %s)", data.Type, data.ID, c.WorkspaceID.Value)
	return nil
}
