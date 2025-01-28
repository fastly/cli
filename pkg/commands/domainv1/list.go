package domainv1

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"
	v1 "github.com/fastly/go-fastly/v9/fastly/domains/v1"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list domains.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	cursor    argparser.OptionalString
	fqdn      argparser.OptionalString
	limit     argparser.OptionalInt
	serviceID argparser.OptionalString
	sort      argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List domains")

	// Optional.
	c.CmdClause.Flag("cursor", "Cursor value from the next_cursor field of a previous response, used to retrieve the next page").Action(c.cursor.Set).StringVar(&c.cursor.Value)
	c.CmdClause.Flag("fqdn", "Filters results by the FQDN using a fuzzy/partial match").Action(c.fqdn.Set).StringVar(&c.fqdn.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("limit", "Limit how many results are returned").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceID.Set,
		Name:        argparser.FlagServiceIDName,
		Description: "Filter results based on a service_id",
		Dst:         &c.serviceID.Value,
		Short:       's',
	})
	c.CmdClause.Flag("sort", "The order in which to list the results").Action(c.sort.Set).StringVar(&c.sort.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &v1.ListInput{}

	if c.serviceID.WasSet {
		input.ServiceID = &c.serviceID.Value
	}
	if c.cursor.WasSet {
		input.Cursor = &c.cursor.Value
	}
	if c.fqdn.WasSet {
		input.FQDN = &c.fqdn.Value
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	for {
		cl, err := v1.List(fc, input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Cursor":     c.cursor.Value,
				"FQDN":       c.fqdn.Value,
				"Limit":      c.limit.Value,
				"Service ID": c.serviceID.Value,
				"Sort":       c.sort.Value,
			})
			return err
		}

		if ok, err := c.WriteJSON(out, cl); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		if c.Globals.Verbose() {
			printVerbose(out, cl.Data)
		} else {
			printSummary(out, cl.Data)
		}

		if cl != nil && cl.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					input.Cursor = &cl.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}
