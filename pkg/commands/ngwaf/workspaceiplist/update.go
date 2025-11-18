package workspaceiplist

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/scope"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/ngwaflist"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a workspace ip list.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	listID      string
	workspaceID argparser.OptionalWorkspaceID

	// Optional.
	description argparser.OptionalString
	entries     argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a workspace level ip list")

	// Required.
	c.CmdClause.Flag("list-id", "List ID").Required().StringVar(&c.listID)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("description", "User submitted description of the list.").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("entries", "Entries for the list. Can either a comma separated list or a path to a file.").Action(c.entries.Set).StringVar(&c.entries.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	input := ngwaflist.ListUpdateInput{
		CommandScope: scope.ScopeTypeWorkspace,
		Description:  c.description,
		Entries:      c.entries,
		ListID:       c.listID,
		WorkspaceID:  &c.workspaceID,
		Out:          out,
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

	text.Success(out, "Updated Workspace IP List '%s' (list id: %s)", data.Name, data.ListID)
	return nil
}
