package iplist

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

// CreateCommand calls the Fastly API to create account-level ip lists.
type CreateCommand struct {
	argparser.Base

	// Required.
	Entries  string `required:"" help:"Entries for the list. Can either be a comma separated list or a path to a file."`
	NameFlag string `name:"name" required:"" help:"User submitted display name of a list."`

	// Optional.
	Description argparser.OptionalString `help:"User submitted description of the list."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	input := ngwaflist.ListCreateInput{
		CommandScope: scope.ScopeTypeAccount,
		Description:  c.Description,
		Entries:      c.Entries,
		Name:         c.NameFlag,
		Type:         "ip",
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

	text.Success(out, "Created Account IP List '%s' (list id: %s)", data.Name, data.ListID)
	return nil
}
