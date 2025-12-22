package rule

import (
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly"

	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"
)

// ListCommand calls the Fastly API to list all workspace-level rules for your API token.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	workspaceID argparser.OptionalWorkspaceID

	// Optional.
	action  argparser.OptionalString
	enabled argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("list", "List all workspace-level rules")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.CmdClause.Flag("action", "Filter rules based on action.").Action(c.action.Set).StringVar(&c.action.Value)
	c.CmdClause.Flag("enabled", "Filter rules based on whether the rule is enabled.").Action(c.enabled.Set).StringVar(&c.enabled.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
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

	input := &rules.ListInput{
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeWorkspace,
			AppliesTo: []string{c.workspaceID.Value},
		},
	}

	if c.action.WasSet {
		input.Action = &c.action.Value
	}

	if c.enabled.WasSet {
		enabled, _ := strconv.ParseBool(c.enabled.Value)
		input.Enabled = &enabled
	}

	rules, err := rules.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, rules); ok {
		return err
	}

	text.PrintRuleTbl(out, rules.Data)
	return nil
}
