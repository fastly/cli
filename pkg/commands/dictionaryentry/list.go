package dictionaryentry

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list dictionary items.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	input       fastly.ListDictionaryItemsInput
	serviceName argparser.OptionalServiceNameID
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
	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(argparser.PaginationDirection[0]).HintOptions(argparser.PaginationDirection...).EnumVar(&c.input.Direction, argparser.PaginationDirection...)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.input.Page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.input.PerPage)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(&c.input.Sort)
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

	c.input.ServiceID = serviceID
	paginator := c.Globals.APIClient.NewListDictionaryItemsPaginator(&c.input)

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
