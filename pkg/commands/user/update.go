package user

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	ID       string `help:"Alphanumeric string identifying the user"`
	Login    string `help:"The login associated with the user (typically, an email address)"`
	NameFlag string `name:"name" help:"The real life name of the user"`
	Reset    bool   `name:"password-reset" help:"Requests a password reset for the specified user"`
	Role     string `default:"" enum:"user,billing,engineer,superuser," help:"The permissions role assigned to the user. Can be user, billing, engineer, or superuser"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Reset {
		input, err := c.constructInputReset()
		if err != nil {
			return err
		}

		err = c.Globals.APIClient.ResetUserPassword(context.TODO(), input)
		if err != nil {
			return err
		}

		text.Success(out, "Reset user password (login: %s)", c.Login)
		return nil
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.APIClient.UpdateUser(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"User ID": c.ID,
		})
		return err
	}

	text.Success(out, "Updated user '%s' (role: %s)", fastly.ToValue(r.Name), fastly.ToValue(r.Role))
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() (*fastly.UpdateUserInput, error) {
	var input fastly.UpdateUserInput

	if c.ID == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --id flag")
	}
	input.UserID = c.ID

	if c.NameFlag == "" && c.Role == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide either the --name or --role with the --id flag")
	}

	if c.NameFlag != "" {
		input.Name = &c.NameFlag
	}
	if c.Role != "" {
		input.Role = &c.Role
	}

	return &input, nil
}

// constructInputReset transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInputReset() (*fastly.ResetUserPasswordInput, error) {
	var input fastly.ResetUserPasswordInput

	if c.Login == "" {
		return nil, fmt.Errorf("error parsing arguments: must provide --login when requesting a password reset")
	}
	input.Login = c.Login

	return &input, nil
}
