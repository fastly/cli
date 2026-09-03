package providerconnection

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list provider connections.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Optional.
	sort argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List provider connections")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("sort", "Sort field. Prefix with '-' for descending order (e.g. \"-created_at\")").Default("created_at").Action(c.sort.Set).StringVar(&c.sort.Value)

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

	input := &providerconnection.ListInput{}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	// The API returns all provider connections in a single response; there is
	// no pagination to handle here.
	data, err := providerconnection.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintProviderConnectionsTbl(out, data.Data)
	return nil
}
