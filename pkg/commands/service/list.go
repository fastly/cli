package service

import (
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	input fastly.GetServicesInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List Fastly services")

	// Optional.
	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(argparser.PaginationDirection[0]).HintOptions(argparser.PaginationDirection...).EnumVar(&c.input.Direction, argparser.PaginationDirection...)
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(c.input.Page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(c.input.PerPage)
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(c.input.Sort)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	paginator := c.Globals.APIClient.GetServices(&c.input)

	var o []*fastly.Service
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
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
		tw := text.NewTable(out)
		tw.AddHeader("NAME", "ID", "TYPE", "ACTIVE VERSION", "LAST EDITED (UTC)")
		for _, service := range o {
			updatedAt := "n/a"
			if service.UpdatedAt != nil {
				updatedAt = service.UpdatedAt.UTC().Format(time.Format)
			}

			activeVersion := strconv.Itoa(fastly.ToValue(service.ActiveVersion))
			for _, v := range service.Versions {
				if fastly.ToValue(v.Number) == fastly.ToValue(service.ActiveVersion) && !fastly.ToValue(v.Active) {
					activeVersion = "n/a"
				}
			}

			tw.AddLine(service.Name,
				fastly.ToValue(service.ID),
				fastly.ToValue(service.Type),
				activeVersion,
				updatedAt,
			)
		}
		tw.Print()
		return nil
	}

	for i, service := range o {
		fmt.Fprintf(out, "Service %d/%d\n", i+1, len(o))
		text.PrintService(out, "\t", service)
		fmt.Fprintln(out)
	}

	return nil
}
