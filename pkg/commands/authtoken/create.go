package authtoken

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// Scopes is a list of purging scope options.
// https://www.fastly.com/documentation/reference/api/auth-tokens#scopes
var Scopes = []string{"global", "purge_select", "purge_all", "global:read"}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	//
	// NOTE: The go-fastly client internally calls `/sudo` before `/tokens` and
	// the sudo endpoint requires a password to be provided alongside an API
	// token. The password must be for the user account that created the token
	// being passed as authentication to the API endpoint.
	Password string `required:"" help:"User password corresponding with --token or $FASTLY_API_TOKEN"`

	// Optional.
	//
	// NOTE: The API describes 'scope' as being space-delimited but we've opted
	// for comma-separated as it means users don't have to worry about how best
	// to handle issues with passing a flag value with whitespace. When
	// constructing the input for the API call we convert from a comma-separated
	// value to a space-delimited value.
	Expires  time.Time `hints:"2016-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"Time-stamp (UTC) of when the token will expire"`
	NameFlag string    `name:"name" help:"Name of the token"`
	Scope    []string  `enum:"global,purge_select,purge_all,global:read" sep:"none" hints:"global,purge_select,purge_all,global:read" help:"Authorization scope (repeat flag per scope)"`
	Services []string  `sep:"," help:"A comma-separated list of alphanumeric strings identifying services (default: access to all services)"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if !c.Globals.Flags.Quiet {
		text.Deprecated("The 'auth-token' command tree will be removed in a future release. Use the Fastly API directly to manage API tokens.\n\n")
	}

	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateToken(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	expires := "never"
	if r.ExpiresAt != nil {
		expires = r.ExpiresAt.String()
	}

	text.Success(out, "Created token '%s' (name: %s, id: %s, scope: %s, expires: %s)", fastly.ToValue(r.AccessToken), fastly.ToValue(r.Name), fastly.ToValue(r.TokenID), fastly.ToValue(r.Scope), expires)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTokenInput {
	var input fastly.CreateTokenInput

	input.Password = new(c.Password)

	if !c.Expires.IsZero() {
		input.ExpiresAt = &c.Expires
	}
	if c.NameFlag != "" {
		input.Name = new(c.NameFlag)
	}
	if len(c.Scope) > 0 {
		input.Scope = new(fastly.TokenScope(strings.Join(c.Scope, " ")))
	}
	if len(c.Services) > 0 {
		input.Services = c.Services
	}

	return &input
}
