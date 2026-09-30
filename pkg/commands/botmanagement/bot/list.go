package bot

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces/policy"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/argparser/validate"
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

	// Optional, can't be empty
	categoryID string
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
	c.CmdClause.Flag("workspace-id", "Workspace ID").
		Required().
		Action(validate.MinLength(1)).
		StringVar(&c.workspaceID)

	// Optional.
	c.CmdClause.Flag("category-id", "Only list bots in this category").
		Action(validate.MinLength(1)).
		StringVar(&c.categoryID)

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
		// If CategoryID is provided (not nil or _empty_), go-fastly
		// will request bots belonging only to that category. This value
		// will be empty if the flag was not provided, so we don't
		// need an additional if check to conditionally set this field.
		CategoryID: &c.categoryID,
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
	if c.categoryID != "" {
		for i := range data {
			if data[i].CategoryID == "" {
				data[i].CategoryID = c.categoryID
			}
		}
	}

	text.PrintBotManagementBotTbl(out, data)
	return nil
}
