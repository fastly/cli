package pagerduty

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/pagerduty"
)

// GetCommand calls the Fastly API to get PagerDuty alerts.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	AlertID     string
	WorkspaceID argparser.OptionalWorkspaceID
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get", "Get a PagerDuty alert")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.WorkspaceID.Value,
		Action:      c.WorkspaceID.Set,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFAlertID,
		Description: argparser.FlagNGWAFAlertIDDesc,
		Dst:         &c.AlertID,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	input := &pagerduty.GetInput{
		AlertID:     &c.AlertID,
		WorkspaceID: &c.WorkspaceID.Value,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := pagerduty.Get(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintAlert(out, data)
	return nil
}
