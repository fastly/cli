package workspace

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to get a workspace.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	workspaceID string
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("get", "Get a workspace")

	// Required.
	c.CmdClause.Flag("workspace-id", "Workspace ID").Required().StringVar(&c.workspaceID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := workspaces.Get(context.TODO(), fc, &workspaces.GetInput{
		WorkspaceID: &c.workspaceID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintWorkspace(out, data)
	return nil
}
