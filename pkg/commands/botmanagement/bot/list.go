package bot

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list the bots in a workspace.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	workspaceID string

	// Optional.
	categoryID argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List the bots in a Bot Management workspace")

	// Required.
	c.CmdClause.Flag("workspace-id", "Workspace ID").Required().StringVar(&c.workspaceID)

	// Optional.
	c.CmdClause.Flag("category-id", "Only list bots in this category").Action(c.categoryID.Set).StringVar(&c.categoryID.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &policy.ListBotsInput{
		WorkspaceID: &c.workspaceID,
	}
	if c.categoryID.WasSet {
		input.CategoryID = &c.categoryID.Value
	}

	data, err := policy.ListBots(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	// The API only populates a bot's category when listing across the whole
	// workspace, so fill it in from the filter to keep the table column useful.
	if c.categoryID.WasSet {
		for i := range data {
			if data[i].CategoryID == "" {
				data[i].CategoryID = c.categoryID.Value
			}
		}
	}

	text.PrintBotManagementBotTbl(out, data)
	return nil
}
