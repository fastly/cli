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

// DescribeCommand calls the Fastly API to describe a bot.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	botID       string
	categoryID  string
	workspaceID string
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a bot").Alias("get")

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

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := policy.GetBot(context.TODO(), fc, &policy.GetBotInput{
		BotID:       &c.botID,
		CategoryID:  &c.categoryID,
		WorkspaceID: &c.workspaceID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if written, err := c.WriteJSON(out, data); written {
		return err
	}

	text.PrintBotManagementBot(out, data)
	return nil
}
