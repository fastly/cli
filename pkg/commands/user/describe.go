package user

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Get a specific user of the Fastly API and web interface").Alias("get")
	c.Globals = globals
	c.manifest = data
	c.CmdClause.Flag("id", "Alphanumeric string identifying the user").StringVar(&c.id)
	c.CmdClause.Flag("current", "Get the logged in user").BoolVar(&c.current)
	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	current  bool
	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	if c.current {
		r, err := c.Globals.Client.GetCurrentUser()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		c.print(out, r)
		return nil
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	r, err := c.Globals.Client.GetUser(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.print(out, r)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() (*fastly.GetUserInput, error) {
	var input fastly.GetUserInput

	if c.id == "" {
		return nil, errors.RemediationError{
			Inner:       fmt.Errorf("error parsing arguments: must provide --id flag"),
			Remediation: "Alternatively pass --current to validate the logged in user.",
		}
	}
	input.ID = c.id

	return &input, nil
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.User) {
	fmt.Fprintf(out, "\nID: %s\n", r.ID)
	fmt.Fprintf(out, "Login: %s\n", r.Login)
	fmt.Fprintf(out, "Name: %s\n", r.Name)
	fmt.Fprintf(out, "Role: %s\n", r.Role)
	fmt.Fprintf(out, "Customer ID: %s\n", r.CustomerID)
	fmt.Fprintf(out, "Email Hash: %s\n", r.EmailHash)
	fmt.Fprintf(out, "Limit Services: %t\n", r.LimitServices)
	fmt.Fprintf(out, "Locked: %t\n", r.Locked)
	fmt.Fprintf(out, "Require New Password: %t\n", r.RequireNewPassword)
	fmt.Fprintf(out, "Two Factor Auth Enabled: %t\n", r.TwoFactorAuthEnabled)
	fmt.Fprintf(out, "Two Factor Setup Required: %t\n\n", r.TwoFactorSetupRequired)

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
