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
	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/scope"
)

// ListCommand calls the Fastly API to list all account-level rules for your API token.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

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

	c.CmdClause = parent.Command("list", "List all account-level rules")

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

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &rules.ListInput{
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeAccount,
			AppliesTo: []string{"*"},
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
