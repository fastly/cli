package dictionaryitem

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ListCommand calls the Fastly API to list dictionary items.
type ListCommand struct {
	cmd.Base
	manifest    manifest.Data
	input       fastly.ListDictionaryItemsInput
	json        bool
	serviceName cmd.OptionalServiceNameID
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List items in a Fastly edge dictionary")
	c.CmdClause.Flag("dictionary-id", "Dictionary ID").Required().StringVar(&c.input.DictionaryID)
	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(cmd.PaginationDirection[0]).HintOptions(cmd.PaginationDirection...).EnumVar(&c.input.Direction, cmd.PaginationDirection...)
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.input.Page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.input.PerPage)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(&c.input.Sort)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := cmd.ServiceID(c.serviceName, c.manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.ServiceID = serviceID
	paginator := c.Globals.APIClient.NewListDictionaryItemsPaginator(&c.input)

	var ds []*fastly.DictionaryItem
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Dictionary ID":   c.input.DictionaryID,
				"Service ID":      serviceID,
				"Remaining Pages": paginator.Remaining(),
			})
			return err
		}
		ds = append(ds, data...)
	}

	if c.json {
		data, err := json.Marshal(ds)
		if err != nil {
			return err
		}
		fmt.Fprint(out, string(data))
		return nil
	}

	if !c.Globals.Verbose() {
		fmt.Fprintf(out, "\nService ID: %s\n", c.input.ServiceID)
	}
	for i, dictionary := range ds {
		text.Output(out, "Item: %d/%d", i+1, len(ds))
		text.PrintDictionaryItem(out, "\t", dictionary)
		text.Break(out)
	}

	return nil
}
