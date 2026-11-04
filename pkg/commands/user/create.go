package user

import (
	"context"
	"fmt"
	"io"
	"slices"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply checks the values kong cannot check by itself.
func (c *CreateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Role.Value != "" || kongcompat.Given(ctx, "role")) && !slices.Contains([]string{"user", "billing", "engineer", "superuser"}, c.Role.Value) {
		return fmt.Errorf("enum value must be one of user,billing,engineer,superuser, got '%s'", c.Role.Value)
	}
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	Login    argparser.OptionalString `help:"The login associated with the user (typically, an email address)"`
	NameFlag argparser.OptionalString `name:"name" help:"The real life name of the user"`

	// Optional.
	Role argparser.OptionalString `hints:"user,billing,engineer,superuser" help:"The permissions role assigned to the user. Can be user, billing, engineer, or superuser"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := c.constructInput()

	r, err := c.Globals.APIClient.CreateUser(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User Login": c.Login,
			"User Name":  c.NameFlag,
		})
		return err
	}

	text.Success(out, "Created user '%s' (role: %s)", fastly.ToValue(r.Name), fastly.ToValue(r.Role))
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateUserInput {
	var input fastly.CreateUserInput
	if c.Login.WasSet {
		input.Login = &c.Login.Value
	}
	if c.Role.WasSet {
		input.Role = &c.Role.Value
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}

	return &input
}
