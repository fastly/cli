package service

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	cmd.Base
	input fastly.ListServicesInput
	json  bool
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.CmdClause = parent.Command("list", "List Fastly services")
	c.CmdClause.Flag("direction", "Direction in which to sort results").Default(cmd.PaginationDirection[0]).HintOptions(cmd.PaginationDirection...).EnumVar(&c.input.Direction, cmd.PaginationDirection...)
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.input.Page)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.input.PerPage)
	c.CmdClause.Flag("sort", "Field on which to sort").Default("created").StringVar(&c.input.Sort)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	paginator := c.Globals.APIClient.NewListServicesPaginator(&c.input)

	var ss []*fastly.Service
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"Remaining Pages": paginator.Remaining(),
			})
			return err
		}
		ss = append(ss, data...)
	}

	if !c.Globals.Verbose() {
		if c.json {
			data, err := json.Marshal(ss)
			if err != nil {
				return err
			}
			out.Write(data)
			return nil
		}

		tw := text.NewTable(out)
		tw.AddHeader("NAME", "ID", "TYPE", "ACTIVE VERSION", "LAST EDITED (UTC)")
		for _, service := range ss {
			updatedAt := "n/a"
			if service.UpdatedAt != nil {
				updatedAt = service.UpdatedAt.UTC().Format(time.Format)
			}

			activeVersion := fmt.Sprint(service.ActiveVersion)
			for _, v := range service.Versions {
				if uint(v.Number) == service.ActiveVersion && !v.Active {
					activeVersion = "n/a"
				}
			}

			tw.AddLine(service.Name, service.ID, service.Type, activeVersion, updatedAt)
		}
		tw.Print()
		return nil
	}

	for i, service := range ss {
		fmt.Fprintf(out, "Service %d/%d\n", i+1, len(ss))
		text.PrintService(out, "\t", service)
		fmt.Fprintln(out)
	}

	return nil
}
