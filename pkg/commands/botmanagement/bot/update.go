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

// Actions are the accepted values for a bot's action. `inherit` applies the
// action of the category the bot belongs to.
var Actions = []string{"allow", "block", "challenge", "inherit"}

// UpdateCommand calls the Fastly API to set the action of a bot.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	botID       string
	categoryID  string
	workspaceID string

	// The update payload
	action string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Set the action applied to requests from a bot")

	// Though go-fastly validates that the ids are not nil or empty,
	// we get more consistent error messages by checking length up front.
	// Otherwise there's nothing linking an API field back to the CLI
	// flag that was used to set it.

	c.CmdClause.Flag("bot-id", "Bot ID").
		Required().
		Action(validate.MinLength(1)).
		StringVar(&c.botID)

	c.CmdClause.Flag("category-id", "Category ID").
		Required().
		Action(validate.MinLength(1)).
		StringVar(&c.categoryID)

	c.CmdClause.Flag("workspace-id", "Workspace ID").
		Required().
		Action(validate.MinLength(1)).
		StringVar(&c.workspaceID)

	// Required.
	c.CmdClause.Flag("action", "Action applied to requests from the bot ('inherit' applies the category's action)").
		Required().
		HintOptions(Actions...).
		EnumVar(&c.action, Actions...)

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

	data, err := policy.UpdateBot(context.TODO(), fc, &policy.UpdateBotInput{
		Action:      &policy.Action{Type: c.action},
		BotID:       &c.botID,
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

	text.Success(out, "Updated bot '%s' (bot-id: %s, action: %s)", data.Name, data.BotID, data.Action.Type)
	return nil
}
