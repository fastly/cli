package auth

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update service authorizations.
type UpdateCommand struct {
	argparser.Base

	input fastly.UpdateServiceAuthorizationInput

	// Required.
	ID         string `required:"" help:"ID of the service authorization to delete"`
	Permission string `short:"p" required:"" enum:"full,read_only,purge_select,purge_all" hints:"full,read_only,purge_select,purge_all" help:"The permission the user has in relation to the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *UpdateCommand) AfterApply() error {
	c.input.ID = c.ID
	c.input.Permission = c.Permission
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	s, err := c.Globals.APIClient.UpdateServiceAuthorization(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service Authorization ID": c.input.ID,
		})
		return err
	}

	text.Success(out, "Updated service authorization %s", s.ID)
	return nil
}
