package authtoken

import (
	"io"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/fastly/kingpin"
)

// https://developer.fastly.com/reference/api/auth/#scopes
var Scopes = []string{"global", "purge_select", "purge_all", "global:read"}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Create an API token").Alias("add")
	c.Globals = globals
	c.manifest = data

	// Required flags
	//
	// NOTE: The go-fastly client internally calls `/sudo` before `/tokens` and
	// the sudo endpoint requires a password to be provided alongside an API
	// token. The password must be for the user account that created the token
	// being passed as authentication to the API endpoint.
	c.CmdClause.Flag("password", "User password corresponding with --token or $FASTLY_API_TOKEN").Required().StringVar(&c.password)

	// Optional flags
	//
	// NOTE: The API describes 'scope' as being space-delimited but we've opted
	// for comma-separated as it means users don't have to worry about how best
	// to handle issues with passing a flag value with whitespace. When
	// constructing the input for the API call we convert from a comma-separated
	// value to a space-delimited value.
	c.CmdClause.Flag("expires", "Time-stamp (UTC) of when the token will expire").HintOptions("2016-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.expires)
	c.CmdClause.Flag("name", "Name of the token").StringVar(&c.name)
	c.CmdClause.Flag("scope", "Authorization scope (repeat flag per scope)").HintOptions(Scopes...).EnumsVar(&c.scope, Scopes...)
	c.CmdClause.Flag("services", "A comma-separated list of alphanumeric strings identifying services (default: access to all services)").StringsVar(&c.services, kingpin.Separator(","))
	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	expires  time.Time
	manifest manifest.Data
	name     string
	password string
	scope    []string
	services []string
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	input := c.constructInput()

	r, err := c.Globals.Client.CreateToken(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	expires := "never"
	if r.ExpiresAt != nil {
		expires = r.ExpiresAt.String()
	}

	text.Success(out, "Created token '%s' (name: %s, id: %s, scope: %s, expires: %s)", r.AccessToken, r.Name, r.ID, r.Scope, expires)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateTokenInput {
	var input fastly.CreateTokenInput

	input.Password = c.password

	if !c.expires.IsZero() {
		input.ExpiresAt = &c.expires
	}
	if c.name != "" {
		input.Name = c.name
	}
	if len(c.scope) > 0 {
		input.Scope = fastly.TokenScope(strings.Join(c.scope, " "))
	}
	if len(c.services) > 0 {
		input.Services = c.services
	}

	return &input
}
