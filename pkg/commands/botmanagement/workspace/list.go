package workspace

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/botmanagement/v1/workspaces"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/argparser/validate"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list Bot Management workspaces.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Optional.
	serviceID string
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Bot Management workspaces")

	// Optional.
	c.CmdClause.
		Flag("service-id", "Only list workspaces attached to this service").
		Action(validate.MinLength(1)).
		StringVar(&c.serviceID)

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

	input := &workspaces.ListInput{}
	// unlike some of the other go-fastly methods, this one only treats nil
	// nil and not empty string as being missing, so we must set it conditionally
	if c.serviceID != "" {
		input.ServiceID = &c.serviceID
	}

	data, err := workspaces.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintBotManagementWorkspaceTbl(out, data)
	return nil
}
