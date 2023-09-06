package user

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a user of the Fastly API and web interface")
	c.Globals = g
	c.manifest = m
	c.CmdClause.Flag("id", "Alphanumeric string identifying the user").StringVar(&c.id)
	c.CmdClause.Flag("login", "The login associated with the user (typically, an email address)").StringVar(&c.login)
	c.CmdClause.Flag("name", "The real life name of the user").StringVar(&c.name)
	c.CmdClause.Flag("password-reset", "Requests a password reset for the specified user").BoolVar(&c.reset)
	c.CmdClause.Flag("role", "The permissions role assigned to the user. Can be user, billing, engineer, or superuser").EnumVar(&c.role, "user", "billing", "engineer", "superuser")

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	cmd.Base

	id       string
	login    string
	manifest manifest.Data
	name     string
	reset    bool
	role     string
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.reset {
		input, err := c.constructInputReset()
		if err != nil {
			return err
		}

		err = c.Globals.APIClient.ResetUserPassword(input)
		if err != nil {
			return err
		}

		text.Success(out, "Reset user password (login: %s)", c.login)
		return nil
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.UpdateUser(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User ID": c.id,
		})
		return err
	}

	text.Success(out, "Updated user '%s' (role: %s)", r.Name, r.Role)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() (*fastly.UpdateUserInput, error) {
	var input fastly.UpdateUserInput

	if c.id == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --id flag")
	}
	input.ID = c.id

	if c.name == "" && c.role == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide either the --name or --role with the --id flag")
	}

	if c.name != "" {
		input.Name = &c.name
	}
	if c.role != "" {
		input.Role = &c.role
	}

	return &input, nil
}

// constructInputReset transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInputReset() (*fastly.ResetUserPasswordInput, error) {
	var input fastly.ResetUserPasswordInput

	if c.login == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --login when requesting a password reset")
	}
	input.Login = c.login

	return &input, nil
}
