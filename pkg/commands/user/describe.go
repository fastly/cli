package user

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	Current bool   `help:"Get the logged in user"`
	ID      string `help:"Alphanumeric string identifying the user"`
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.Current {
		o, err := c.Globals.APIClient.GetCurrentUser(context.TODO())
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if ok, err := c.WriteJSON(out, o); ok {
			return err
		}

		c.print(out, o)
		return nil
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	o, err := c.Globals.APIClient.GetUser(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	c.print(out, o)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() (*fastly.GetUserInput, error) {
	var input fastly.GetUserInput

	if c.ID == "" {
		return nil, fsterr.RemediationError{
			Inner:       fmt.Errorf("error parsing arguments: must provide --id flag"),
			Remediation: "Alternatively pass --current to validate the logged in user.",
		}
	}
	input.UserID = c.ID

	return &input, nil
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.User) {
	fmt.Fprintf(out, "\nID: %s\n", fastly.ToValue(r.UserID))
	fmt.Fprintf(out, "Login: %s\n", fastly.ToValue(r.Login))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(r.Name))
	fmt.Fprintf(out, "Role: %s\n", fastly.ToValue(r.Role))
	fmt.Fprintf(out, "Customer ID: %s\n", fastly.ToValue(r.CustomerID))
	fmt.Fprintf(out, "Email Hash: %s\n", fastly.ToValue(r.EmailHash))
	fmt.Fprintf(out, "Limit Services: %t\n", fastly.ToValue(r.LimitServices))
	fmt.Fprintf(out, "Locked: %t\n", fastly.ToValue(r.Locked))
	fmt.Fprintf(out, "Require New Password: %t\n", fastly.ToValue(r.RequireNewPassword))
	fmt.Fprintf(out, "Two Factor Auth Enabled: %t\n", fastly.ToValue(r.TwoFactorAuthEnabled))
	fmt.Fprintf(out, "Two Factor Setup Required: %t\n\n", fastly.ToValue(r.TwoFactorSetupRequired))

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}
	if r.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", r.UpdatedAt)
	}
	if r.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", r.DeletedAt)
	}
}
