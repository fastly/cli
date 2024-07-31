package dictionaryentry

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list dictionary items.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	direction     string
	input         fastly.GetDictionaryItemsInput
	page, perPage int
	serviceName   argparser.OptionalServiceNameID
	sort          string
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List items in a Fastly edge dictionary")

	// Required.
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.input.DictionaryID)

	// Optional.
	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(argparser.PaginationDirection[0]).HintOptions(argparser.PaginationDirection...).EnumVar(&c.direction, argparser.PaginationDirection...)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.perPage)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(&c.sort)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.Direction = &c.direction
	c.input.Page = &c.page
	c.input.PerPage = &c.perPage
	c.input.ServiceID = serviceID
	c.input.Sort = &c.sort
	paginator := c.Globals.APIClient.GetDictionaryItems(&c.input)

	var o []*fastly.DictionaryItem
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Dictionary ID":   c.input.DictionaryID,
				"Service ID":      serviceID,
				"Remaining Pages": paginator.Remaining(),
			})
			return err
		}
		o = append(o, data...)
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", c.input.ServiceID)
	}
	for i, dictionary := range o {
		text.Output(out, "Item: %d/%d", i+1, len(o))
		text.PrintDictionaryItem(out, "\t", dictionary)
		text.Break(out)
	}

	return nil
}
