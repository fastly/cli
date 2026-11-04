package rule

import (
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/scope"
)

// ListCommand calls the Fastly API to list all workspace-level rules for your API token.
type ListCommand struct {
	argparser.Base

	// Required.
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	Action      argparser.OptionalString `help:"Filter rules based on action."`
	EnabledFlag argparser.OptionalString `name:"enabled" help:"Filter rules based on whether the rule is enabled."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &rules.ListInput{
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeWorkspace,
			AppliesTo: []string{c.WorkspaceID.Value},
		},
	}

	if c.Action.WasSet {
		input.Action = &c.Action.Value
	}

	if c.EnabledFlag.WasSet {
		enabled, _ := strconv.ParseBool(c.EnabledFlag.Value)
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

	text.PrintRuleTbl(out, rules)
	return nil
}
