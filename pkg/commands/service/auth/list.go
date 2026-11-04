package auth

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v19/fastly"
)

// ListCommand calls the Fastly API to list service authorizations.
type ListCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput

	input fastly.ListServiceAuthorizationsInput

	Page    int `help:"Page number of data set to fetch"`
	PerPage int `help:"Number of records per page"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.input.PageNumber = c.Page
	c.input.PageSize = c.PerPage
	return nil
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.ListServiceAuthorizations(context.TODO(), &c.input)
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
