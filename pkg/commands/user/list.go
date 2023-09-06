package user

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all users from a specified customer id")
	c.Globals = g
	c.manifest = m
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagCustomerIDName,
		Description: cmd.FlagCustomerIDDesc,
		Dst:         &c.customerID.Value,
		Action:      c.customerID.Set,
	})
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base
	cmd.JSONOutput

	customerID cmd.OptionalCustomerID
	manifest   manifest.Data
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.customerID.Parse(); err != nil {
		return err
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListCustomerUsers(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Customer ID": c.customerID.Value,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListCustomerUsersInput {
	var input fastly.ListCustomerUsersInput

	input.CustomerID = c.customerID.Value

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, us []*fastly.User) {
	for _, u := range us {
		fmt.Fprintf(out, "\nID: %s\n", u.ID)
		fmt.Fprintf(out, "Login: %s\n", u.Login)
		fmt.Fprintf(out, "Name: %s\n", u.Name)
		fmt.Fprintf(out, "Role: %s\n", u.Role)
		fmt.Fprintf(out, "Customer ID: %s\n", u.CustomerID)
		fmt.Fprintf(out, "Email Hash: %s\n", u.EmailHash)
		fmt.Fprintf(out, "Limit Services: %t\n", u.LimitServices)
		fmt.Fprintf(out, "Locked: %t\n", u.Locked)
		fmt.Fprintf(out, "Require New Password: %t\n", u.RequireNewPassword)
		fmt.Fprintf(out, "Two Factor Auth Enabled: %t\n", u.TwoFactorAuthEnabled)
		fmt.Fprintf(out, "Two Factor Setup Required: %t\n\n", u.TwoFactorSetupRequired)

		if u.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", u.CreatedAt)
		}
		if u.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", u.UpdatedAt)
		}
		if u.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", u.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, us []*fastly.User) error {
	t := text.NewTable(out)
	t.AddHeader("LOGIN", "NAME", "ROLE", "LOCKED", "ID")
	for _, u := range us {
		t.AddLine(u.Login, u.Name, u.Role, u.Locked, u.ID)
	}
	t.Print()
	return nil
}
