package auth

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete service authorizations.
type DeleteCommand struct {
	argparser.Base
	Input fastly.DeleteServiceAuthorizationInput `kong:"-"`

	// Required.
	ID string `required:"" help:"ID of the service authorization to delete"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DeleteCommand) AfterApply() error {
	c.Input.ID = c.ID
	return nil
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := c.Globals.APIClient.DeleteServiceAuthorization(context.TODO(), &c.Input); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.Input.ID,
		})
		return err
	}

	text.Success(out, "Deleted service authorization %s", c.Input.ID)
	return nil
}
