package category

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

// Actions are the accepted values for a category's action. Unlike bots,
// categories cannot `inherit`, as there is nothing above them to inherit from.
var Actions = []string{"allow", "block", "challenge"}

// UpdateCommand calls the Fastly API to set the action of a bot category.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	categoryFilterFields

	// Required payload.
	action string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Set the action applied to bots in a category that are configured to inherit")

	c.categoryFilterFields.register(c.CmdClause)

	// Required.
	c.CmdClause.Flag("action", "Action applied to every bot in the category configured with the 'inherit' action").
		Required().
		HintOptions(Actions...).
		EnumVar(&c.action, Actions...)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := policy.UpdateCategory(context.TODO(), fc, &policy.UpdateCategoryInput{
		Action:      &policy.Action{Type: c.action},
		CategoryID:  &c.categoryID,
		WorkspaceID: &c.workspaceID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated bot category '%s' (category-id: %s, action: %s)", data.Name, data.CategoryID, data.Action.Type)
	return nil
}
