package workspaces

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a workspace.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	workspaceID string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("delete", "Delete a workspace")

	// Required.
	c.CmdClause.Flag("workspace-id", "Workspace ID").Required().StringVar(&c.workspaceID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	err := workspaces.Delete(context.TODO(), fc, &workspaces.DeleteInput{
		WorkspaceID: &c.workspaceID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.workspaceID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted workspace (id: %s)", c.workspaceID)
	return nil
}
