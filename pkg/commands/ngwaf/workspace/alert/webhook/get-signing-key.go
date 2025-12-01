package webhook

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/alertutil"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/webhook"
)

// GetSigningKeyCommand calls the Fastly API to get Webhook alerts.
type GetSigningKeyCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	alertutil.AlertIDFlags
	alertutil.BaseAlertFlags
}

// NewGetSigningKeyCommand returns a usable command registered under the parent.
func NewGetSigningKeyCommand(parent argparser.Registerer, g *global.Data) *GetSigningKeyCommand {
	c := GetSigningKeyCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get-signing-key", "Retrieves details of a webhook alert signing key")

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
func (c *GetSigningKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	input := &webhook.GetKeyInput{
		AlertID:     &c.AlertID,
		WorkspaceID: &c.WorkspaceID.Value,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := webhook.GetKey(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Signing Key: '%s' (Workspace: %s)", data.SigningKey, c.WorkspaceID.Value)
	return nil
}
