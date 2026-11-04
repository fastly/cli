package signallist

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/ngwaflist"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a workspace signal list.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ListID      string                        `required:"" help:"List ID"`
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	Description argparser.OptionalString `help:"User submitted description of the list."`
	Entries     argparser.OptionalString `help:"Entries for the list. Can either be a comma separated list or a path to a file."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	input := ngwaflist.ListUpdateInput{
		CommandScope: scope.ScopeTypeWorkspace,
		Description:  c.Description,
		Entries:      c.Entries,
		ListID:       c.ListID,
		WorkspaceID:  &c.WorkspaceID,
	}

	var ok bool
	input.FC, ok = c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := ngwaflist.ListUpdate(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated Workspace Signal List '%s' (list id: %s)", data.Name, data.ListID)
	return nil
}
