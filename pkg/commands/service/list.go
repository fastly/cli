package service

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	argparser.Base

	// Optional.
	Direction string `default:"ascend" enum:"ascend,descend" hints:"ascend,descend" help:"Direction in which to sort results"`
	argparser.JSONOutput

	Page    int `help:"Page number of data set to fetch"`
	PerPage int `help:"Number of records per page"`
	input   fastly.GetServicesInput
	Sort    string `default:"created" help:"Field on which to sort"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	c.input.Direction = &c.Direction
	c.input.Page = &c.Page
	c.input.PerPage = &c.PerPage
	c.input.Sort = &c.Sort
	paginator := c.Globals.APIClient.GetServices(context.TODO(), &c.input)

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

			tw.AddLine(
				fastly.ToValue(service.Name),
				fastly.ToValue(service.ServiceID),
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
