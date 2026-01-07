package rule

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to get a workspace-level rule.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	ruleID      string
	workspaceID argparser.OptionalWorkspaceID
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("get", "Get a workspace-level rule")

	// Required.
	c.CmdClause.Flag("rule-id", "Rule ID").Required().StringVar(&c.ruleID)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
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
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := rules.Get(context.TODO(), fc, &rules.GetInput{
		RuleID: &c.ruleID,
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeWorkspace,
			AppliesTo: []string{c.workspaceID.Value},
		},
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintRule(out, data)
	return nil
}
