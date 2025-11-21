package signallist

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

// CreateCommand calls the Fastly API to create account-level signal lists.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	entries string
	name    string

	// Optional.
	description argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an account-level signal list").Alias("add")

	// Required.
	c.CmdClause.Flag("entries", "Entries for the list. Can either a comma separated list or a path to a file.").Required().StringVar(&c.entries)
	c.CmdClause.Flag("name", "User submitted display name of a list.").Required().StringVar(&c.name)

	// Optional.
	c.CmdClause.Flag("description", "User submitted description of the list.").Action(c.description.Set).StringVar(&c.description.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	input := ngwaflist.ListCreateInput{
		CommandScope: scope.ScopeTypeAccount,
		Description:  c.description,
		Entries:      c.entries,
		Name:         c.name,
		Type:         "signal",
		WorkspaceID:  nil,
	}

	var ok bool
	input.FC, ok = c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := ngwaflist.ListCreate(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created Account Signal List '%s' (list id: %s)", data.Name, data.ListID)
	return nil
}
