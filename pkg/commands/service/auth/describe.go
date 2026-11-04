package auth

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/time"
)

// DescribeCommand calls the Fastly API to describe a service authorization for a service.
type DescribeCommand struct {
	argparser.Base

	// Required.
	ID string `required:"" help:"ID of the service authorization to retrieve"`

	// Optional.
	argparser.JSONOutput

	Input fastly.GetServiceAuthorizationInput `kong:"-"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.Input.ID = c.ID
	return nil
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetServiceAuthorization(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.Input.ID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(o, out)
}

func (c *DescribeCommand) print(s *fastly.ServiceAuthorization, out io.Writer) error {
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

	return nil
}
