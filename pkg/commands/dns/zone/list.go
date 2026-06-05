package zone

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/dnszones"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list DNS Zones.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	name argparser.OptionalString
	sort argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List DNS Zones")

	// Optional.
	c.CmdClause.Flag("name", "Filter zones to only those containing the provided name.").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("sort", "Order in which to list results. Valid values are: name_asc, name_desc, created_at_asc, created_at_desc.").Action(c.sort.Set).HintOptions(SortOptions...).EnumVar(&c.sort.Value, SortOptions...)
	c.RegisterFlagBool(c.JSONFlag()) // --json

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

	input := &dnszones.ListInput{}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.sort.WasSet {
		v := sortAPIValue[c.sort.Value]
		input.Sort = &v
	}

	zones, err := dnszones.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, zones); ok {
		return err
	}

	text.PrintDNSZoneTbl(out, zones)
	return nil
}
