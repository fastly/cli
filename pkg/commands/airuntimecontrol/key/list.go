package key

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list virtual keys.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Optional.
	model          argparser.OptionalString
	provider       argparser.OptionalString
	includeDeleted argparser.OptionalBool
	search         argparser.OptionalString
	limit          int
	sort           argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List virtual keys")

	// Optional.
	c.CmdClause.Flag("model", "Filter by AI model identifier").Action(c.model.Set).StringVar(&c.model.Value)
	c.CmdClause.Flag("provider", "Filter by AI model provider").Action(c.provider.Set).StringVar(&c.provider.Value)
	c.CmdClause.Flag("include-deleted", "Include deleted virtual keys").Action(c.includeDeleted.Set).BoolVar(&c.includeDeleted.Value)
	c.CmdClause.Flag("search", "Search virtual keys by substring match on key name").Action(c.search.Set).StringVar(&c.search.Value)
	c.RegisterFlagBool(c.JSONFlag())                 // --json
	c.RegisterFlagInt(argparser.LimitFlag(&c.limit)) // --limit
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

	input := &key.ListInput{
		Limit: &c.limit,
	}
	if c.model.WasSet {
		input.Model = &c.model.Value
	}
	if c.provider.WasSet {
		input.Provider = &c.provider.Value
	}
	if c.includeDeleted.WasSet {
		input.IncludeDeleted = &c.includeDeleted.Value
	}
	if c.search.WasSet {
		input.Search = &c.search.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	// key.List() automatically paginates through all pages internally.
	data, err := key.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintVirtualKeysTbl(out, data)
	return nil
}
