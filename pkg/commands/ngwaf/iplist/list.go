package iplist

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

// ListCommand calls the Fastly API to list all ip lists for your API token.

type ListCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := ngwaflist.ListListInput{
		CommandScope: scope.ScopeTypeAccount,
		Type:         "ip",
		WorkspaceID:  nil,
	}

	var ok bool
	input.FC, ok = c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	lists, err := ngwaflist.ListList(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, *lists); ok {
		return err
	}

	text.PrintListTbl(out, lists.Data)
	return nil
}
