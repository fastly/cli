package wildcardlist

import (
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/ngwaflist"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/scope"
)

// GetCommand calls the Fastly API to get a workspace-level wildcard list.
type GetCommand struct {
	argparser.Base

	// Required.
	ListID      string                        `required:"" help:"List ID"`
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := ngwaflist.ListGetInput{
		CommandScope: scope.ScopeTypeWorkspace,
		ListID:       c.ListID,
		WorkspaceID:  &c.WorkspaceID,
	}

	var ok bool
	input.FC, ok = c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	list, err := ngwaflist.ListGet(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, list); ok {
		return err
	}

	text.PrintList(out, list)
	return nil
}
