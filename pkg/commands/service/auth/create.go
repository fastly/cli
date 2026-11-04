package auth

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// Permissions is a list of supported permission values.
// https://www.fastly.com/documentation/reference/api/account/service-authorization#data-model
var Permissions = []string{"full", "read_only", "purge_select", "purge_all"}

// CreateCommand calls the Fastly API to create a service authorization.
type CreateCommand struct {
	argparser.Base
	input fastly.CreateServiceAuthorizationInput

	// Required.
	UserID string `short:"u" required:"" help:"Alphanumeric string identifying the user"`

	// Optional.
	// NOTE: We default to 'read_only' for security reasons.
	// The API otherwise defaults to 'full' permissions!
	Permission  string                          `short:"p" default:"read_only" enum:"full,read_only,purge_select,purge_all" hints:"full,read_only,purge_select,purge_all" help:"The permission the user has in relation to the service (default: read_only)"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	c.input.Permission = c.Permission
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":   c.Globals.Manifest.Flag.ServiceID,
			"Service Name": c.ServiceName.Value,
		})
		return err
	}
	if c.Globals.Flags.Verbose {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.input.Service = &fastly.SAService{
		ID: serviceID,
	}
	c.input.User = &fastly.SAUser{
		ID: c.UserID,
	}

	s, err := c.Globals.APIClient.CreateServiceAuthorization(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
			"Flag":       flag,
		})
		return err
	}

	text.Success(out, "Created service authorization %s", s.ID)
	return nil
}
