package session

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/session"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list session logs.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Optional.
	key      argparser.OptionalString
	provider argparser.OptionalString
	model    argparser.OptionalString
	from     time.Time
	to       time.Time
	limit    int
	sort     argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List session logs")

	// Optional.
	c.CmdClause.Flag("key", "Filter by a specific virtual key ID").Action(c.key.Set).StringVar(&c.key.Value)
	c.CmdClause.Flag("provider", "Filter by provider name").Action(c.provider.Set).StringVar(&c.provider.Value)
	c.CmdClause.Flag("model", "Filter by model name").Action(c.model.Set).StringVar(&c.model.Value)
	c.CmdClause.Flag("from", "Start of the time range (RFC 3339 format)").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.from)
	c.CmdClause.Flag("to", "End of the time range (RFC 3339 format). Defaults to now").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.to)
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

	input := &session.ListInput{
		Limit: &c.limit,
	}
	if c.key.WasSet {
		input.Key = &c.key.Value
	}
	if c.provider.WasSet {
		input.Provider = &c.provider.Value
	}
	if c.model.WasSet {
		input.Model = &c.model.Value
	}
	if !c.from.IsZero() {
		input.From = &c.from
	}
	if !c.to.IsZero() {
		input.To = &c.to
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	// session.List() automatically paginates through all pages internally.
	data, err := session.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintSessionsTbl(out, data)
	return nil
}
