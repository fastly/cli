package serviceauth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v10/fastly"
)

// ListCommand calls the Fastly API to list service authorizations.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	input fastly.ListServiceAuthorizationsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.Globals = g
	c.CmdClause = parent.Command("list", "List service authorizations")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.input.PageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.input.PageSize)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.ListServiceAuthorizations(&c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Page Number": c.input.PageNumber,
			"Page Size":   c.input.PageSize,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		if len(o.Items) > 0 {
			tw := text.NewTable(out)
			tw.AddHeader("AUTH ID", "USER ID", "SERVICE ID", "PERMISSION")

			for _, s := range o.Items {
				tw.AddLine(s.ID, s.User.ID, s.Service.ID, s.Permission)
			}
			tw.Print()

			return nil
		}
	}

	for _, s := range o.Items {
		fmt.Fprintf(out, "Auth ID: %s\n", s.ID)
		fmt.Fprintf(out, "User ID: %s\n", s.User.ID)
		fmt.Fprintf(out, "Service ID: %s\n", s.Service.ID)
		fmt.Fprintf(out, "Permission: %s\n", s.Permission)

		if s.CreatedAt != nil {
			fmt.Fprintf(out, "Created (UTC): %s\n", s.CreatedAt.UTC().Format(time.Format))
		}
		if s.UpdatedAt != nil {
			fmt.Fprintf(out, "Last edited (UTC): %s\n", s.UpdatedAt.UTC().Format(time.Format))
		}
		if s.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted (UTC): %s\n", s.DeletedAt.UTC().Format(time.Format))
		}
	}

	return nil
}
