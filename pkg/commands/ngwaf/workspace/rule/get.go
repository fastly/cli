package rule

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/rules"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to get a workspace-level rule.
type GetCommand struct {
	argparser.Base

	// Required.
	RuleID      string                        `required:"" help:"Rule ID"`
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
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

	data, err := rules.Get(context.TODO(), fc, &rules.GetInput{
		RuleID: &c.RuleID,
		Scope: &scope.Scope{
			Type:      scope.ScopeTypeWorkspace,
			AppliesTo: []string{c.WorkspaceID.Value},
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
