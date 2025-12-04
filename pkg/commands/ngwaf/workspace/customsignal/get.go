package customsignal

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/signals"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to get an workspace-level custom signal.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	signalID    string
	workspaceID argparser.OptionalWorkspaceID
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("get", "Get a custom signal")

	// Required.
	c.CmdClause.Flag("signal-id", "Custom Signal ID").Required().StringVar(&c.signalID)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
		Required:    true,
	})

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

	input := &signals.GetInput{
		SignalID: &c.signalID,
		Scope: &scope.Scope{
			Type: scope.ScopeTypeWorkspace,
		},
	}

	if err := c.workspaceID.Parse(); err != nil {
		return err
	}
	input.Scope.AppliesTo = []string{c.workspaceID.Value}

	data, err := signals.Get(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintCustomSignal(out, data)
	return nil
}
